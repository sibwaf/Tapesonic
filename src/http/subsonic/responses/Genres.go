package responses

type Genres struct {
	Genre []Genre `json:"genre" xml:"genre"`
}

func NewGenres(genres []Genre) *Genres {
	return &Genres{
		Genre: genres,
	}
}
