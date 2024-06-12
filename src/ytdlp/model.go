package ytdlp

type YtdlpPlaylist struct {
	Id string `json:"id"`

	Title      string `json:"title"`
	Uploader   string `json:"uploader"`
	UploaderId string `json:"uploader_id"`
	WebpageUrl string `json:"webpage_url"`

	ExtractorKey string `json:"extractor_key"`
}

type YtdlpFile struct {
	Id         string `json:"id"`
	Title      string `json:"title"`
	Channel    string `json:"channel"`
	WebpageUrl string `json:"webpage_url"`

	PlaylistIndex int `json:"playlist_index"`

	Artist string `json:"artist"`
	Album  string `json:"album"`
	Track  string `json:"track"`

	Duration float64 `json:"duration"`

	Ext      string         `json:"ext"`
	Formats  []YtdlpFormat  `json:"formats"`
	Chapters []YtdlpChapter `json:"chapters"`
	Tags     []string       `json:"tags"`

	ExtractorKey string `json:"extractor_key"`
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
