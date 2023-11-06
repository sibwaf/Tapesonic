package ytdlp

import "encoding/json"

type DownloadInfo struct {
	MediaPath     string
	ThumbnailPath string
	RawMetadata   []byte
}

func (di DownloadInfo) ParseMetadata() (YtdlpMetadata, error) {
	var metadata YtdlpMetadata
	err := json.Unmarshal(di.RawMetadata, &metadata)
	return metadata, err
}

type YtdlpMetadata struct {
	Id         string         `json:"id"`
	Title      string         `json:"title"`
	Channel    string         `json:"channel"`
	WebpageUrl string         `json:"webpage_url"`
	Ext        string         `json:"ext"`
	Formats    []YtdlpFormat  `json:"formats"`
	Chapters   []YtdlpChapter `json:"chapters"`
	Tags       []string       `json:"tags"`
}

type YtdlpFormat struct {
	Format       string  `json:"format"`
	FormatId     string  `json:"format_id"`
	FormatNote   string  `json:"format_note"`
	VCodec       string  `json:"vcodec"`
	ACodec       string  `json:"acodec"`
	Filesize     int     `json:"filesize"`
	AudioBitrate float64 `json:"abr"`
	Quality      float64 `json:"quality"`
}

type YtdlpChapter struct {
	Title     string  `json:"title"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
}
