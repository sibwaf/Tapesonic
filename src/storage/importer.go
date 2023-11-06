package storage

import (
	"tapesonic/ytdlp"
)

type Importer struct {
	mediaDir    string
	ytdlp       *ytdlp.Ytdlp
	dataStorage *DataStorage
}

func NewImporter(
	mediaDir string,
	ytdlp *ytdlp.Ytdlp,
	dataStorage *DataStorage,
) *Importer {
	return &Importer{
		mediaDir:    mediaDir,
		ytdlp:       ytdlp,
		dataStorage: dataStorage,
	}
}

func (i *Importer) ImportTape(url string, format string) (string, error) {
	downloadInfo, err := i.ytdlp.Download(url, format, i.mediaDir)
	if err != nil {
		return "", err
	}

	metadata, err := downloadInfo.ParseMetadata()
	if err != nil {
		return "", err
	}

	tracks := []*TapeTrack{}
	for index, chapter := range metadata.Chapters {
		track := TapeTrack{
			TapeTrackIndex: index,

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
		Id:            metadata.Id,
		Metadata:      string(downloadInfo.RawMetadata),
		Url:           metadata.WebpageUrl,
		Name:          metadata.Title,
		AuthorName:    metadata.Channel,
		ThumbnailPath: downloadInfo.ThumbnailPath,
		Tracks:        tracks,
	}

	err = i.dataStorage.CreateTape(&tape)

	return tape.Id, err
}
