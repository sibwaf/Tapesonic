package ytdlp

import "encoding/json"

type DownloadedPlaylist struct {
	Metadata YtdlpPlaylistWrapper

	Files []DownloadedFile
}

type DownloadedFile struct {
	Metadata YtdlpFileWrapper

	MediaPath     string
	ThumbnailPath string
}

type YtdlpPlaylistWrapper struct {
	Raw      []byte
	metadata *YtdlpPlaylist
}

func (w *YtdlpPlaylistWrapper) Parse() (YtdlpPlaylist, error) {
	if len(w.Raw) == 0 {
		return YtdlpPlaylist{}, nil
	}
	if w.metadata != nil {
		return *w.metadata, nil
	}

	w.metadata = &YtdlpPlaylist{}
	err := json.Unmarshal(w.Raw, w.metadata)
	if err != nil {
		w.metadata = nil
		return YtdlpPlaylist{}, err
	}

	return *w.metadata, nil
}

type YtdlpFileWrapper struct {
	Raw      []byte
	metadata *YtdlpFile
}

func (w *YtdlpFileWrapper) Parse() (YtdlpFile, error) {
	if len(w.Raw) == 0 {
		return YtdlpFile{}, nil
	}
	if w.metadata != nil {
		return *w.metadata, nil
	}

	w.metadata = &YtdlpFile{}
	err := json.Unmarshal(w.Raw, w.metadata)
	if err != nil {
		w.metadata = nil
		return YtdlpFile{}, err
	}

	return *w.metadata, nil
}
