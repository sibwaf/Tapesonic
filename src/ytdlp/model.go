package ytdlp

type YtdlpMetadata struct {
	Title string `json:"title"`
	Channel string `json:"channel"`
	WebpageUrl string `json:"webpage_url"`
	Formats []YtdlpFormat `json:"formats"`
	Chapters []YtdlpChapter `json:"chapters"`
	Tags []string `json:"tags"`
}

type YtdlpFormat struct {
	Format string `json:"format"`
	FormatId string `json:"format_id"`
	FormatNote string `json:"format_note"`
	VCodec string `json:"vcodec"`
	ACodec string `json:"acodec"`
	Filesize int `json:"filesize"`
	AudioBitrate float64 `json:"abr"`
	Quality float64 `json:"quality"`
}

type YtdlpChapter struct {
	Title string `json:"title"`
	StartTime float64 `json:"start_time"`
	EndTime float64 `json:"end_time"`
}
