package admin

import (
	"encoding/json"
	"net/http"
	"tapesonic/model"
	"tapesonic/sources"
	"tapesonic/util"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type SourceFileRs struct {
	Codec string
}

func toSourceFileRs(file sources.SourceFile) SourceFileRs {
	return SourceFileRs{
		Codec: file.Codec,
	}
}

type SourceListRs struct {
	Id uuid.UUID

	Url         string
	Title       string
	Uploader    string
	DurationMs  int64
	ThumbnailId *uuid.UUID
	File        *SourceFileRs
}

func toSourceListRs(source sources.SourceForApi) SourceListRs {
	result := SourceListRs{
		Id:          source.Source.Id,
		Url:         source.Source.Url,
		Title:       source.Source.Title,
		Uploader:    source.Source.Uploader,
		DurationMs:  source.Source.DurationMs,
		ThumbnailId: source.Source.ThumbnailId,
	}

	if source.File != nil {
		file := toSourceFileRs(*source.File)
		result.File = &file
	}

	return result
}

func GetSources(auth *authenticator, sources *sources.SourceService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		managementPolicies := []string{}

		for key, values := range r.URL.Query() {
			if key == "managementPolicy" {
				managementPolicies = append(managementPolicies, values...)
			}
		}

		sources, err := sources.GetListForApi(managementPolicies)
		if err != nil {
			return nil, err
		}

		return util.Map(sources, toSourceListRs), nil
	}
}

type SourceFullRs struct {
	Id uuid.UUID

	Url         string
	Title       string
	Uploader    string
	AlbumArtist string
	AlbumTitle  string
	AlbumIndex  int
	TrackArtist string
	TrackTitle  string
	DurationMs  int64
	ReleaseDate *time.Time
	ThumbnailId *uuid.UUID
	File        *SourceFileRs
}

func toSourceFullRs(source sources.Source, file *sources.SourceFile) SourceFullRs {
	result := SourceFullRs{
		Id:          source.Id,
		Url:         source.Url,
		Title:       source.Title,
		Uploader:    source.Uploader,
		AlbumArtist: source.AlbumArtist,
		AlbumTitle:  source.AlbumTitle,
		AlbumIndex:  source.AlbumIndex,
		TrackArtist: source.TrackArtist,
		TrackTitle:  source.TrackTitle,
		DurationMs:  source.DurationMs,
		ReleaseDate: source.ReleaseDate.UnwrapNullable(),
		ThumbnailId: source.ThumbnailId,
	}

	if file != nil {
		fileRs := toSourceFileRs(*file)
		result.File = &fileRs
	}

	return result
}

func GetSource(auth *authenticator, sources *sources.SourceService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		sourceId, err := uuid.Parse(mux.Vars(r)["sourceId"])
		if err != nil {
			return nil, model.ErrNotFound
		}

		source, err := sources.GetByIdForApi(sourceId)
		if err != nil {
			return nil, err
		}

		return toSourceFullRs(source.Source, source.File), nil
	}
}

type SourceHierarchyListRs struct {
	Id       uuid.UUID
	ParentId *uuid.UUID

	Url         string
	Title       string
	Uploader    string
	ListIndex   int
	ThumbnailId *uuid.UUID
}

func toSourceHierarchyListRs(source sources.SourceForHierarchy) SourceHierarchyListRs {
	return SourceHierarchyListRs{
		Id:       source.Id,
		ParentId: source.ParentId,

		Url:         source.Url,
		Title:       source.Title,
		Uploader:    source.Uploader,
		ListIndex:   source.ListIndex,
		ThumbnailId: source.ThumbnailId,
	}
}

func GetSourceHierarchy(auth *authenticator, sources *sources.SourceService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		sourceId, err := uuid.Parse(mux.Vars(r)["sourceId"])
		if err != nil {
			return nil, model.ErrNotFound
		}

		hierarchy, err := sources.GetHierarchy(sourceId)
		if err != nil {
			return nil, err
		}

		return util.Map(hierarchy, toSourceHierarchyListRs), nil
	}
}

type SourceTrackRsArtist struct {
	Id   uuid.UUID
	Name string
}

type SourceTrackRs struct {
	Id       uuid.UUID
	SourceId uuid.UUID

	Artist *SourceTrackRsArtist
	Title  string

	StartOffsetMs int64
	EndOffsetMs   int64
}

func toSourceTrackRs(track sources.SavedSourceTrack) SourceTrackRs {
	trackRs := SourceTrackRs{
		Id:            track.Id,
		SourceId:      track.SourceId,
		Title:         track.Title,
		StartOffsetMs: track.StartOffsetMs,
		EndOffsetMs:   track.EndOffsetMs,
	}

	if track.ArtistId != nil {
		trackRs.Artist = &SourceTrackRsArtist{
			Id:   *track.ArtistId,
			Name: track.ArtistName,
		}
	}

	return trackRs
}

func GetSourceTracks(auth *authenticator, sourcesSvc *sources.SourceService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		sourceId, err := uuid.Parse(mux.Vars(r)["sourceId"])
		if err != nil {
			return nil, model.ErrNotFound
		}

		recursive := util.StringToBoolOrDefault(r.URL.Query().Get("recursive"), false)

		var tracks []sources.SavedSourceTrack
		if recursive {
			tracks, err = sourcesSvc.GetAllTracks(sourceId)
		} else {
			tracks, err = sourcesSvc.GetDirectTracks(sourceId)
		}

		if err != nil {
			return nil, err
		}

		return util.Map(tracks, toSourceTrackRs), nil
	}
}

type SourceTrackRq struct {
	Id *uuid.UUID

	ArtistId *uuid.UUID
	Title    string

	StartOffsetMs int64
	EndOffsetMs   int64
}

func sourceTrackRqToTrack(rq SourceTrackRq) sources.SourceTrack {
	track := sources.SourceTrack{
		ArtistId:      rq.ArtistId,
		Title:         rq.Title,
		StartOffsetMs: rq.StartOffsetMs,
		EndOffsetMs:   rq.EndOffsetMs,
	}

	if rq.Id != nil {
		track.Id = *rq.Id
	}

	return track
}

func PutSourceTracks(auth *authenticator, sourcesSvc *sources.SourceService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		sourceId, err := uuid.Parse(mux.Vars(r)["sourceId"])
		if err != nil {
			return nil, model.ErrNotFound
		}

		rq := []SourceTrackRq{}
		if err := json.NewDecoder(r.Body).Decode(&rq); err != nil {
			return nil, err
		}

		tracks := util.Map(rq, sourceTrackRqToTrack)

		savedTracks, err := sourcesSvc.ReplaceTracks(sourceId, tracks, sources.SOURCE_MANAGEMENT_POLICY_MANUAL)
		if err != nil {
			return nil, err
		}

		return util.Map(savedTracks, toSourceTrackRs), nil
	}
}

func DeleteSourceFile(auth *authenticator, sources *sources.SourceService) WebappHandler {
	return func(r *http.Request) (any, error) {
		_, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		sourceId, err := uuid.Parse(mux.Vars(r)["sourceId"])
		if err != nil {
			return nil, model.ErrNotFound
		}

		return nil, sources.DeleteFile(sourceId)
	}
}
