package ytdlp

type YtdlpFile struct {
	Id         string  `json:"id"`
	Title      string  `json:"title"`
	Uploader   string  `json:"uploader"`
	UploaderId string  `json:"uploader_id"`
	Channel    string  `json:"channel"`
	Timestamp  float64 `json:"timestamp"`
	Url        string  `json:"url"`
	WebpageUrl string  `json:"webpage_url"`
	Thumbnail  string  `json:"thumbnail"`

	PlaylistIndex int `json:"playlist_index"`

	Artist      string `json:"artist"`
	Album       string `json:"album"`
	AlbumArtist string `json:"album_artist"`
	Track       string `json:"track"`
	TrackNumber int    `json:"track_number"`

	Duration float64 `json:"duration"`

	ReleaseDate string `json:"release_date"`

	Ext                string                   `json:"ext"`
	Formats            []YtdlpFormat            `json:"formats"`
	Chapters           []YtdlpChapter           `json:"chapters"`
	Tags               []string                 `json:"tags"`
	Entries            []YtdlpFile              `json:"entries"`
	RequestedDownloads []YtdlpRequestedDownload `json:"requested_downloads"`

	ACodec   string `json:"acodec"`
	AudioExt string `json:"audio_ext"`

	ExtractorKey string `json:"extractor_key"`
	Type         string `json:"_type"`
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
	Url          string  `json:"url"`
}

type YtdlpChapter struct {
	Title     string  `json:"title"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
}

type YtdlpRequestedDownload struct {
	ACodec   string `json:"acodec"`
	Ext      string `json:"ext"`
	Filename string `json:"filename"`
}
