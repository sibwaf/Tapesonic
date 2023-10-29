package responses

type RandomSongs struct {
	Song []SubsonicChild `json:"song" xml:"song"`
}

func NewRandomSongs(songs []SubsonicChild) *RandomSongs {
	return &RandomSongs{
		Song: songs,
	}
}
