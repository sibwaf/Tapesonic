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

type PlaylistWrapper struct {
	Items []PlaylistItem `json:"playlist"`
}

type PlaylistItem struct {
	Url  string `json:"url"`
	Name string `json:"name"`

	Artists   []Artist   `json:"artists"`
	Playlinks []Playlink `json:"playlinks"`
}

type Artist struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

type Playlink struct {
	Affiliate string `json:"affiliate"`
	Url       string `json:"url"`
}
