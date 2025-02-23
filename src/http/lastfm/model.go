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
