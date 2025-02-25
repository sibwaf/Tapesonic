package lastfm

type TokenWrapper struct {
	Token string `json:"token"`
}

type SessionWrapper struct {
	Session Session `json:"session"`
}

type Session struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type UpdateNowPlayingRq struct {
	Artist string
	Track  string
	Album  string
}

type ScrobbleRq struct {
	Artist    string
	Track     string
	Album     string
	Timestamp int64
}
