package config

type TapesonicConfig struct {
	Username string
	Password string

	WebappDir       string
	MediaStorageDir string

	YtdlpPath  string
	FfmpegPath string
}
