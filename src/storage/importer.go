package storage

import (
	"fmt"
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
	downloadedPlaylist, err := i.ytdlp.Download(url, format, i.mediaDir)
	if err != nil {
		return &Tape{}, err
	}

	if len(downloadedPlaylist.Files) == 0 {
		return &Tape{}, fmt.Errorf("no files downloaded from %s", url)
	}

	playlistMetadata, err := downloadedPlaylist.Metadata.Parse()
	if err != nil {
		return &Tape{}, err
	}

	tape := Tape{
		Metadata: string(downloadedPlaylist.Metadata.Raw),
		Url:      playlistMetadata.WebpageUrl,

		Name:       playlistMetadata.Title,
		AuthorName: util.Coalesce(playlistMetadata.Uploader, playlistMetadata.UploaderId),

		ThumbnailPath: downloadedPlaylist.ThumbnailPath,
	}

	for _, downloadedFile := range downloadedPlaylist.Files {
		tapeFile, err := extractFile(
			downloadedFile.Metadata,
			downloadedFile.ThumbnailPath,
			downloadedFile.MediaPath,
		)
		if err != nil {
			return &tape, err
		}

		tape.Files = append(tape.Files, tapeFile)

		tape.Metadata = util.Coalesce(tape.Metadata, tapeFile.Metadata)
		tape.Url = util.Coalesce(tape.Url, tapeFile.Url)
		tape.Name = util.Coalesce(tape.Name, tapeFile.Name)
		tape.AuthorName = util.Coalesce(tape.AuthorName, tapeFile.AuthorName)
		tape.ThumbnailPath = util.Coalesce(tape.ThumbnailPath, tapeFile.ThumbnailPath)
	}

	return &tape, i.tapeStorage.UpsertTape(&tape)
}

func extractFile(
	metadataWrapper ytdlp.YtdlpFileWrapper,
	thumbnailPath string,
	mediaPath string,
) (*TapeFile, error) {
	metadata, err := metadataWrapper.Parse()
	if err != nil {
		return nil, err
	}

	tapeFile := TapeFile{
		Metadata: string(metadataWrapper.Raw),
		Url:      metadata.WebpageUrl,

		Name:       metadata.Title,
		AuthorName: metadata.Channel,

		ThumbnailPath: thumbnailPath,
		MediaPath:     mediaPath,
	}

	for _, chapter := range metadata.Chapters {
		track := TapeTrack{
			RawStartOffsetMs: int(chapter.StartTime * 1000),
			StartOffsetMs:    int(chapter.StartTime * 1000),
			RawEndOffsetMs:   int(chapter.EndTime * 1000),
			EndOffsetMs:      int(chapter.EndTime * 1000),

			Artist: metadata.Artist,
			Title:  chapter.Title,
		}
		tapeFile.Tracks = append(tapeFile.Tracks, &track)
	}

	if len(tapeFile.Tracks) == 0 {
		track := TapeTrack{
			RawStartOffsetMs: 0,
			StartOffsetMs:    0,
			RawEndOffsetMs:   int(metadata.Duration * 1000),
			EndOffsetMs:      int(metadata.Duration * 1000),

			Artist: metadata.Artist,
			Title:  util.Coalesce(metadata.Track, metadata.Title),
		}
		tapeFile.Tracks = append(tapeFile.Tracks, &track)
	}

	return &tapeFile, nil
}
