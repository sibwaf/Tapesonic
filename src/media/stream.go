package media

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"syscall"
	"tapesonic/ffmpeg"
	"tapesonic/model"
	"tapesonic/remotes"
	"tapesonic/sources"
	"tapesonic/storage"
	"tapesonic/subsonic"
	"tapesonic/users"
	"tapesonic/util"
	"tapesonic/ytdlp"

	"github.com/google/uuid"
)

// some codecs like mp4/alac are not supported by Chromium-based clients
var allowedStreamingCodecs = []string{"mp3", "flac", "opus"}

type StreamService struct {
	remotes *remotes.RemoteService
	sources *sources.SourceService
	cache   *storage.StreamCacheStorage
	ffmpeg  *ffmpeg.Ffmpeg
	ytdlp   *ytdlp.YtdlpService
}

func newStreamService(
	remotes *remotes.RemoteService,
	sources *sources.SourceService,
	cache *storage.StreamCacheStorage,
	ffmpeg *ffmpeg.Ffmpeg,
	ytdlp *ytdlp.YtdlpService,
) *StreamService {
	return &StreamService{
		remotes: remotes,
		sources: sources,
		cache:   cache,
		ffmpeg:  ffmpeg,
		ytdlp:   ytdlp,
	}
}

func (svc *StreamService) ServeStream(user users.User, r *http.Request, w http.ResponseWriter, track model.LibraryTrack) error {
	if track.RemoteId == nil {
		uuidId, err := uuid.Parse(track.Id)
		if err != nil {
			return model.ErrNotFound
		}

		descriptor, err := svc.sources.GetTrackFileDescriptor(uuidId)
		if err != nil {
			return err
		}

		if descriptor.LocalPath != "" {
			allowDirectStreaming := true
			switch {
			case descriptor.StartOffsetMs > 0:
				slog.Debug(fmt.Sprintf("Direct streaming for track id=`%s` (%s) is forbidden because StartOffsetMs > 0 (%d)", track.Id, descriptor.LocalPath, descriptor.StartOffsetMs))
				allowDirectStreaming = false
			case descriptor.EndOffsetMs != descriptor.SourceDurationMs:
				slog.Debug(fmt.Sprintf("Direct streaming for track id=`%s` (%s) is forbidden because EndOffsetMs != SourceDurationMs (%d != %d)", track.Id, descriptor.LocalPath, descriptor.EndOffsetMs, descriptor.SourceDurationMs))
				allowDirectStreaming = false
			case !slices.Contains(allowedStreamingCodecs, descriptor.LocalCodec):
				slog.Debug(fmt.Sprintf("Direct streaming for track id=`%s` (%s) is forbidden because codec `%s` is not allowed", track.Id, descriptor.LocalPath, descriptor.LocalCodec))
				allowDirectStreaming = false
			}

			if allowDirectStreaming {
				slog.Debug(fmt.Sprintf("Streaming downloaded track id=`%s` (%s) directly from file", track.Id, descriptor.LocalPath))

				http.ServeFile(w, r, descriptor.LocalPath)
				return nil
			}

			slog.Debug(fmt.Sprintf("Streaming downloaded track id=`%s` (%s) via ffmpeg, start=%d, end=%d", track.Id, descriptor.LocalPath, descriptor.StartOffsetMs, descriptor.EndOffsetMs))

			item, reader, err := svc.cache.GetOrSave(fmt.Sprintf("tapesonic-%s", track.Id), func() (string, io.ReadCloser, error) {
				slog.Debug(fmt.Sprintf("Populating stream cache for track id=`%s`", track.Id))

				format, reader, err := svc.ffmpeg.StreamFrom(
					r.Context(),
					descriptor.LocalCodec,
					ffmpeg.ANY_FORMAT,
					descriptor.StartOffsetMs,
					descriptor.EndOffsetMs-descriptor.StartOffsetMs,
					descriptor.LocalPath,
				)
				if err != nil {
					return "", nil, err
				}

				return util.FormatToMediaType(format), reader, nil
			})
			if reader != nil {
				defer reader.Close()
			}
			if err != nil {
				return err
			}

			slog.Debug(fmt.Sprintf("Got streaming data for track id=`%s`", track.Id))

			w.Header().Add("Content-Type", item.ContentType)
			_, err = io.Copy(w, reader)
			return err
		} else if descriptor.RemoteUrl != "" {
			streamInfo, err := svc.ytdlp.GetStreamInfo(r.Context(), descriptor.RemoteUrl, "ba")
			if err != nil {
				return err
			}

			slog.Debug(fmt.Sprintf("Streaming remote track id=`%s` (%s) via ffmpeg, start=%d, end=%d", track.Id, descriptor.RemoteUrl, descriptor.StartOffsetMs, descriptor.EndOffsetMs))

			format, reader, err := svc.ffmpeg.StreamFrom(
				r.Context(),
				streamInfo.ACodec,
				ffmpeg.SEEKABLE_FORMAT,
				descriptor.StartOffsetMs,
				descriptor.EndOffsetMs-descriptor.StartOffsetMs,
				streamInfo.Url,
			)
			if reader != nil {
				defer reader.Close()
			}
			if err != nil {
				return err
			}

			slog.Debug(fmt.Sprintf("Got streaming data for track id=`%s`", track.Id))

			w.Header().Add("Content-Type", util.FormatToMediaType(format))
			_, err = io.Copy(w, reader)
			return err
		} else {
			return fmt.Errorf("no local path or remote url for track id=`%s`", track.Id)
		}
	} else {
		if track.RemoteTrackId == nil {
			return fmt.Errorf("track id=%s is remote, but is missing remote_track_id", track.Id)
		}

		remote, err := svc.remotes.GetById(*track.RemoteId)
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

			res, err := client.Stream(r.Context(), auth, *track.RemoteTrackId, FilterProxyHeaders(r.Header))
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
