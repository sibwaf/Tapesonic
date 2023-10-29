package storage

import (
	"tapesonic/ytdlp"
)

type Importer struct {
	mediaDir string
	ytdlp    *ytdlp.Ytdlp
}

func NewImporter(
	mediaDir string,
	ytdlp *ytdlp.Ytdlp,
) *Importer {
	return &Importer{
		mediaDir: mediaDir,
		ytdlp:    ytdlp,
	}
}

func (i *Importer) ImportMixtape(url string, format string) error {
	_, err := i.ytdlp.Download(url, format, i.mediaDir)
	return err
}
