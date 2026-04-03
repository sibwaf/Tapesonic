package media

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"syscall"
	"tapesonic/logic"
	"tapesonic/model"
	"tapesonic/remotes"
	"tapesonic/subsonic"
	"tapesonic/users"

	"github.com/google/uuid"
)

type CoverService struct {
	remotes    *remotes.RemoteService
	thumbnails *logic.ThumbnailService
}

func newCoverService(
	remotes *remotes.RemoteService,
	thumbnails *logic.ThumbnailService,
) *CoverService {
	return &CoverService{
		remotes:    remotes,
		thumbnails: thumbnails,
	}
}

func (svc *CoverService) ServeCover(user users.User, r *http.Request, w http.ResponseWriter, cover model.LibraryCover) error {
	if cover.RemoteId == uuid.Nil {
		// todo: proper file streaming
		mime, reader, err := svc.thumbnails.GetThumbnailContent(cover.Id)
		if err != nil {
			return err
		}

		defer reader.Close()

		w.Header().Add("Content-Type", mime)

		_, err = io.Copy(w, reader)
		if errors.Is(err, syscall.EPIPE) {
			// client cancelled the request
			return nil
		}

		return err
	} else {
		remote, err := svc.remotes.GetById(cover.RemoteId)
		if err != nil {
			return err
		}

		credentials, err := svc.remotes.GetCredentials(user, remote)
		if err != nil {
			return err
		}

		switch remote.Type {
		case remotes.REMOTE_TYPE_SUBSONIC:
			client := subsonic.NewSubsonicClient(remote.Url)
			auth := remotes.GetSubsonicAuthMethod(&credentials)

			res, err := client.GetCoverArt(r.Context(), auth, cover.RemoteCoverId, FilterProxyHeaders(r.Header))
			if err != nil {
				return err
			}

			defer res.Body.Close()

			for key, values := range FilterProxyHeaders(res.Header) {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}

			w.WriteHeader(res.StatusCode)

			_, err = io.Copy(w, res.Body)
			if errors.Is(err, syscall.EPIPE) {
				// client cancelled the request
				return nil
			}

			return err
		default:
			return fmt.Errorf("unknown remote type %s", remote.Type)
		}
	}
}
