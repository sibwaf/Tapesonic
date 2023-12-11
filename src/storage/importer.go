package storage

import (
	"tapesonic/ytdlp"

	"github.com/google/uuid"
)

type Importer struct {
	mediaDir    string
	ytdlp       *ytdlp.Ytdlp
	tapeStorage *TapeStorage
}

func NewImporter(
	mediaDir string,
	ytdlp *ytdlp.Ytdlp,
	tapeStorage *TapeStorage,
) *Importer {
	return &Importer{
		mediaDir:    mediaDir,
		ytdlp:       ytdlp,
		tapeStorage: tapeStorage,
	}
}

func (i *Importer) ImportTape(url string, format string) (*Tape, error) {
	downloadInfo, err := i.ytdlp.Download(url, format, i.mediaDir)
	if err != nil {
		return &Tape{}, err
	}

	metadata, err := downloadInfo.ParseMetadata()
	if err != nil {
		return &Tape{}, err
	}

	tracks := []*TapeTrack{}
	for _, chapter := range metadata.Chapters {
		track := TapeTrack{
			Id: uuid.New(),

			FilePath: downloadInfo.MediaPath,

			RawStartOffsetMs: int(chapter.StartTime) * 1000,
			StartOffsetMs:    int(chapter.StartTime) * 1000,
			RawEndOffsetMs:   int(chapter.EndTime) * 1000,
			EndOffsetMs:      int(chapter.EndTime) * 1000,

			Title: chapter.Title,
		}
		tracks = append(tracks, &track)
	}

	tape := Tape{
		Id:            uuid.New(),
		Metadata:      string(downloadInfo.RawMetadata),
		Url:           metadata.WebpageUrl,
		Name:          metadata.Title,
		AuthorName:    metadata.Channel,
		ThumbnailPath: downloadInfo.ThumbnailPath,
		Tracks:        tracks,
	}

	err = i.tapeStorage.UpsertTape(&tape)

	return &tape, err
}
