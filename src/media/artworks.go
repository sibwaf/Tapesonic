package media

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"syscall"
	"tapesonic/artworks"
	"tapesonic/model"
	"tapesonic/remotes"
	"tapesonic/subsonic"
	"tapesonic/users"

	"github.com/google/uuid"
)

type ArtworkService struct {
	remotes  *remotes.RemoteService
	artworks *artworks.ArtworkService
}

func newArtworkService(
	remotes *remotes.RemoteService,
	artworks *artworks.ArtworkService,
) *ArtworkService {
	return &ArtworkService{
		remotes:  remotes,
		artworks: artworks,
	}
}

func (svc *ArtworkService) ServeArtwork(user users.User, r *http.Request, w http.ResponseWriter, artwork model.LibraryArtwork) error {
	if artwork.RemoteId == uuid.Nil {
		descriptor, err := svc.artworks.GetFileDescriptor(artwork.Id)
		if err != nil {
			return err
		}

		http.ServeFile(w, r, descriptor.LocalPath)
		return nil
	} else {
		remote, err := svc.remotes.GetById(artwork.RemoteId)
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

			res, err := client.GetCoverArt(r.Context(), auth, artwork.RemoteArtworkId, FilterProxyHeaders(r.Header))
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
