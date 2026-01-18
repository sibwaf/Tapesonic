package remotes

import (
	"errors"
	"fmt"
	"tapesonic/model"
	"tapesonic/subsonic"
	"tapesonic/users"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
)

type RemoteService struct {
	remotes *RemoteStorage
	planner *TaskPlanner
}

func newRemoteService(remotes *RemoteStorage, planner *TaskPlanner) *RemoteService {
	return &RemoteService{
		remotes: remotes,
		planner: planner,
	}
}

func (svc *RemoteService) GetAllForApi() ([]Remote, error) {
	return svc.remotes.GetAll()
}

func (svc *RemoteService) GetById(id uuid.UUID) (Remote, error) {
	remote, err := svc.remotes.FindById(id)
	if err != nil {
		return Remote{}, err
	}
	if remote == nil {
		return Remote{}, model.ErrNotFound
	}

	return *remote, nil
}

func (svc *RemoteService) Create(settings RemoteSettings) (Remote, error) {
	remote := Remote{
		Id:                           uuid.New(),
		Name:                         settings.Name,
		Type:                         settings.Type,
		Url:                          settings.Url,
		IsScrobbleReplicationEnabled: settings.IsScrobbleReplicationEnabled,
		IsExternalScrobblingEnabled:  settings.IsExternalScrobblingEnabled,
		CreatedAt:                    util.NewTimestampWrapper(time.Now()),
	}

	return svc.remotes.Create(remote)
}

func (svc *RemoteService) Update(id uuid.UUID, settings RemoteSettings) (Remote, error) {
	return svc.remotes.UpdateSettings(id, settings)
}

func (svc *RemoteService) Delete(id uuid.UUID) error {
	return svc.remotes.Delete(id)
}

func (svc *RemoteService) Authenticate(user users.User, remote Remote, credentials RemoteCredentials) error {
	err := svc.remotes.SetCredentials(user.Id, remote.Id, credentials)
	if err != nil {
		return err
	}

	switch remote.Type {
	case REMOTE_TYPE_SUBSONIC:
		return svc.planner.ScheduleSubsonicLibrarySync(user.Id, remote.Id)
	default:
		return nil
	}
}

func (svc *RemoteService) Deauthenticate(user users.User, remote Remote) error {
	err := svc.remotes.RemoveCredentials(user.Id, remote.Id)
	if err != nil {
		return err
	}

	switch remote.Type {
	case REMOTE_TYPE_SUBSONIC:
		return svc.planner.CancelSubsonicLibrarySync(user.Id, remote.Id)
	default:
		return nil
	}
}

func (svc *RemoteService) GetCredentials(user users.User, remote Remote) (RemoteCredentials, error) {
	credentials, err := svc.remotes.FindCredentials(user.Id, remote.Id)
	if err != nil {
		return RemoteCredentials{}, err
	} else if credentials == nil {
		return RemoteCredentials{}, model.ErrNotAuthenticated
	} else {
		return *credentials, err
	}
}

func (svc *RemoteService) GetRemoteStatus(user users.User, remote Remote) (RemoteInfo, error) {
	credentials, err := svc.remotes.FindCredentials(user.Id, remote.Id)
	if err != nil {
		return RemoteInfo{}, err
	}

	switch remote.Type {
	case REMOTE_TYPE_SUBSONIC:
		return getSubsonicRemoteStatus(remote, credentials)
	default:
		return RemoteInfo{}, fmt.Errorf("unknown remote type %s", remote.Type)
	}
}

func getSubsonicRemoteStatus(remote Remote, credentials *RemoteCredentials) (RemoteInfo, error) {
	client := subsonic.NewSubsonicClient(remote.Url)
	auth := GetSubsonicAuthMethod(credentials)

	response, err := client.Ping(auth)

	result := RemoteInfo{
		Name:    response.Type,
		Version: response.ServerVersion,
		Status:  REMOTE_STATUS_OK,
	}

	if credentials != nil {
		result.Username = credentials.Username
	}

	if errors.Is(err, subsonic.ErrInvalidResponse) {
		result.Status = REMOTE_STATUS_BAD_RESPONSE
		result.ErrorDescription = "server returned invalid response"
	} else if errors.Is(err, subsonic.ErrUnreachable) {
		result.Status = REMOTE_STATUS_UNREACHABLE
		result.ErrorDescription = "server is not reachable"
	} else if errors.Is(err, subsonic.NewSubsonicError(subsonic.ERROR_CODE_NOT_AUTHENTICATED, "")) {
		result.Status = REMOTE_STATUS_NOT_AUTHENTICATED
		result.ErrorDescription = "not authenticated"
	} else if errors.Is(err, subsonic.NewSubsonicError(subsonic.ERROR_CODE_PARAMETER_MISSING, "")) && credentials == nil {
		result.Status = REMOTE_STATUS_NOT_AUTHENTICATED
		result.ErrorDescription = "not authenticated"
	} else if err != nil {
		return RemoteInfo{}, err
	}

	return result, nil
}
