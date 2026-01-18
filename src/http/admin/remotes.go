package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"tapesonic/model"
	"tapesonic/remotes"
	"tapesonic/util"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type RemoteRq struct {
	Name string
	Type remotes.RemoteType
	Url  string

	IsScrobbleReplicationEnabled bool
	IsExternalScrobblingEnabled  bool
}

func toRemoteSettings(remoteRq RemoteRq) remotes.RemoteSettings {
	return remotes.RemoteSettings{
		Name:                         remoteRq.Name,
		Type:                         remoteRq.Type,
		Url:                          remoteRq.Url,
		IsScrobbleReplicationEnabled: remoteRq.IsScrobbleReplicationEnabled,
		IsExternalScrobblingEnabled:  remoteRq.IsExternalScrobblingEnabled,
	}
}

type RemoteListRs struct {
	Id string

	Type remotes.RemoteType
	Name string
	Url  string

	IsScrobbleReplicationEnabled bool
	IsExternalScrobblingEnabled  bool
}

func toRemoteListRs(remote remotes.Remote) RemoteListRs {
	return RemoteListRs{
		Id:                           remote.Id.String(),
		Type:                         remote.Type,
		Name:                         remote.Name,
		Url:                          remote.Url,
		IsScrobbleReplicationEnabled: remote.IsScrobbleReplicationEnabled,
		IsExternalScrobblingEnabled:  remote.IsExternalScrobblingEnabled,
	}
}

type RemoteFullRs struct {
	Id string

	Type remotes.RemoteType
	Name string
	Url  string

	IsScrobbleReplicationEnabled bool
	IsExternalScrobblingEnabled  bool

	Username string

	Description string
	Status      string
}

func toRemoteFullRs(remote remotes.Remote, status remotes.RemoteInfo) RemoteFullRs {
	return RemoteFullRs{
		Id:                           remote.Id.String(),
		Type:                         remote.Type,
		Name:                         remote.Name,
		Url:                          remote.Url,
		IsScrobbleReplicationEnabled: remote.IsScrobbleReplicationEnabled,
		IsExternalScrobblingEnabled:  remote.IsExternalScrobblingEnabled,
		Username:                     status.Username,
		Description:                  strings.TrimSpace(fmt.Sprintf("%s %s", status.Name, status.Version)),
		Status:                       status.Status,
	}
}

func GetRemotes(auth *authenticator, remoteSvc *remotes.RemoteService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		remotes, err := remoteSvc.GetAllForApi()
		if err != nil {
			return nil, err
		}

		return util.Map(remotes, toRemoteListRs), nil
	}
}

func PostRemotes(auth *authenticator, remoteSvc *remotes.RemoteService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authorize(r, model.ROLE_ADMIN)
		if err != nil {
			return nil, err
		}

		var remoteRq RemoteRq
		err = json.NewDecoder(r.Body).Decode(&remoteRq)
		if err != nil {
			return nil, err
		}

		remote, err := remoteSvc.Create(toRemoteSettings(remoteRq))
		if err != nil {
			return nil, err
		}

		status, err := remoteSvc.GetRemoteStatus(user, remote)
		if err != nil {
			return nil, err
		}

		return toRemoteFullRs(remote, status), nil
	}
}

func GetRemote(auth *authenticator, remoteSvc *remotes.RemoteService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		remote, err := getRemoteByRawId(mux.Vars(r)["remoteId"], remoteSvc)
		if err != nil {
			return nil, err
		}

		status, err := remoteSvc.GetRemoteStatus(user, remote)
		if err != nil {
			return nil, err
		}

		return toRemoteFullRs(remote, status), nil
	}
}

func PutRemote(auth *authenticator, remoteSvc *remotes.RemoteService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authorize(r, model.ROLE_ADMIN)
		if err != nil {
			return nil, err
		}

		id, err := uuid.Parse(mux.Vars(r)["remoteId"])
		if err != nil {
			return nil, model.ErrNotFound
		}

		var remoteRq RemoteRq
		err = json.NewDecoder(r.Body).Decode(&remoteRq)
		if err != nil {
			return nil, err
		}

		remote, err := remoteSvc.Update(id, toRemoteSettings(remoteRq))
		if err != nil {
			return nil, err
		}

		status, err := remoteSvc.GetRemoteStatus(user, remote)
		if err != nil {
			return nil, err
		}

		return toRemoteFullRs(remote, status), nil
	}
}

func DeleteRemote(auth *authenticator, remoteSvc *remotes.RemoteService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authorize(r, model.ROLE_ADMIN)
		if err != nil {
			return nil, err
		}

		id, err := uuid.Parse(mux.Vars(r)["remoteId"])
		if err != nil {
			return nil, nil
		}

		return nil, remoteSvc.Delete(id)
	}
}

func PutRemoteAuth(auth *authenticator, remoteSvc *remotes.RemoteService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		remote, err := getRemoteByRawId(mux.Vars(r)["remoteId"], remoteSvc)
		if err != nil {
			return nil, err
		}

		var credentials remotes.RemoteCredentials
		err = json.NewDecoder(r.Body).Decode(&credentials)
		if err != nil {
			return nil, err
		}

		err = remoteSvc.Authenticate(user, remote, credentials)
		if err != nil {
			return nil, err
		}

		status, err := remoteSvc.GetRemoteStatus(user, remote)
		if err != nil {
			return nil, err
		}

		return toRemoteFullRs(remote, status), nil
	}
}

func DeleteRemoteAuth(auth *authenticator, remoteSvc *remotes.RemoteService) WebappHandler {
	return func(r *http.Request) (any, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		remote, err := getRemoteByRawId(mux.Vars(r)["remoteId"], remoteSvc)
		if err != nil {
			return nil, err
		}

		err = remoteSvc.Deauthenticate(user, remote)
		if err != nil {
			return nil, err
		}

		status, err := remoteSvc.GetRemoteStatus(user, remote)
		if err != nil {
			return nil, err
		}

		return toRemoteFullRs(remote, status), nil
	}
}

func getRemoteByRawId(rawId string, remoteSvc *remotes.RemoteService) (remotes.Remote, error) {
	remoteId, err := uuid.Parse(rawId)
	if err != nil {
		return remotes.Remote{}, model.ErrNotFound
	}

	return remoteSvc.GetById(remoteId)
}
