package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"tapesonic/ytdlp"
)

type Storage struct {
	dir string
}

func NewStorage(dir string) *Storage {
	return &Storage{
		dir: dir,
	}
}

func (s *Storage) GetTapes() ([]TapeInfo, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return []TapeInfo{}, err
	}

	result := []TapeInfo{}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".info.json") {
			continue
		}

		tapeId := strings.TrimSuffix(entry.Name(), ".info.json")
		tape, err := s.GetTape(tapeId)
		if err != nil {
			return []TapeInfo{}, err
		}

		result = append(result, tape)
	}

	return result, nil
}

func (s *Storage) GetTape(id string) (TapeInfo, error) {
	metadataRaw, err := os.ReadFile(getMetadataPath(id, s.dir))
	if err != nil {
		return TapeInfo{}, err
	}

	var metadata ytdlp.YtdlpMetadata
	err = json.Unmarshal(metadataRaw, &metadata)
	if err != nil {
		return TapeInfo{}, err
	}

	tracks := []TrackInfo{}
	totalOffsetMs := 0
	for index, chapter := range metadata.Chapters {
		lengthMs := int(chapter.EndTime-chapter.StartTime) * 1000

		track := TrackInfo{
			Name:     chapter.Title,
			Index:    index,
			OffsetMs: totalOffsetMs,
			LengthMs: lengthMs,
		}

		tracks = append(tracks, track)
		totalOffsetMs += lengthMs
	}

	return TapeInfo{
		Id:     id,
		Name:   metadata.Title,
		Author: metadata.Channel,
		Tracks: tracks,
	}, nil
}

func (s *Storage) GetStreamableTrack(id string) (StreamableTrack, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		return StreamableTrack{}, fmt.Errorf("invalid id `%s`, expected format `x/y`", id)
	}

	tape, err := s.GetTape(idParts[0])
	if err != nil {
		return StreamableTrack{}, err
	}

	trackIndex, err := strconv.Atoi(idParts[1])
	if err != nil {
		return StreamableTrack{}, fmt.Errorf("track index `%s` is not a number", idParts[1])
	}
	if trackIndex < 0 || trackIndex >= len(tape.Tracks) {
		return StreamableTrack{}, fmt.Errorf("tape `%s` doesn't contain track with index `%d`", id, trackIndex)
	}

	track := tape.Tracks[trackIndex]
	return StreamableTrack{
		Path:  getMediaPath(tape.Id, s.dir),
		Track: track,
	}, nil
}

func (s *Storage) GetCover(id string) (Cover, error) {
	return Cover{
		Path: getCoverPath(id, s.dir),
	}, nil
}

func getMediaPath(tapeId string, storageDir string) string {
	return path.Join(storageDir, tapeId+".webm")
}

func getMetadataPath(tapeId string, storageDir string) string {
	return path.Join(storageDir, tapeId+".info.json")
}

func getCoverPath(tapeId string, storageDir string) string {
	return path.Join(storageDir, tapeId+".png")
}
