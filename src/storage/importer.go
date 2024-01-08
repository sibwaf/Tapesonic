package storage

import (
	"encoding/json"
	"tapesonic/util"
	"tapesonic/ytdlp"
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

	tapeFile, err := extractFile(
		downloadInfo.RawMetadata,
		downloadInfo.ThumbnailPath,
		downloadInfo.MediaPath,
	)
	if err != nil {
		return &Tape{}, err
	}

	files := []*TapeFile{tapeFile}

	tapeMetadata, err := downloadInfo.ParseMetadata()
	if err != nil {
		return &Tape{}, err
	}

	tape := Tape{
		Metadata:      string(downloadInfo.RawMetadata),
		Url:           tapeMetadata.WebpageUrl,
		Name:          tapeMetadata.Title,
		AuthorName:    tapeMetadata.Channel,
		ThumbnailPath: downloadInfo.ThumbnailPath,
		Files:         files,
	}

	return &tape, i.tapeStorage.UpsertTape(&tape)
}

func extractFile(
	rawMetadata []byte,
	thumbnailPath string,
	mediaPath string,
) (*TapeFile, error) {
	var metadata *ytdlp.YtdlpMetadata
	err := json.Unmarshal(rawMetadata, &metadata)
	if err != nil {
		return nil, err
	}

	tracks := []*TapeTrack{}
	for _, chapter := range metadata.Chapters {
		track := TapeTrack{
			RawStartOffsetMs: int(chapter.StartTime) * 1000,
			StartOffsetMs:    int(chapter.StartTime) * 1000,
			RawEndOffsetMs:   int(chapter.EndTime) * 1000,
			EndOffsetMs:      int(chapter.EndTime) * 1000,

			Artist: metadata.Artist,
			Title:  chapter.Title,
		}
		tracks = append(tracks, &track)
	}

	if len(tracks) == 0 {
		track := TapeTrack{
			RawStartOffsetMs: 0,
			StartOffsetMs:    0,
			RawEndOffsetMs:   metadata.Duration * 1000,
			EndOffsetMs:      metadata.Duration * 1000,

			Artist: metadata.Artist,
			Title:  util.Coalesce(metadata.Track, metadata.Title),
		}
		tracks = append(tracks, &track)
	}

	return &TapeFile{
		Metadata: string(rawMetadata),
		Url:      metadata.WebpageUrl,

		Name:       metadata.Title,
		AuthorName: metadata.Channel,

		ThumbnailPath: thumbnailPath,
		MediaPath:     mediaPath,

		Tracks: tracks,
	}, nil
}
