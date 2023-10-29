package responses

type Genre struct {
	Value      string `json:"value" xml:"value,attr"`
	SongCount  int    `json:"songCount" xml:"songCount,attr"`
	AlbumCount int    `json:"albumCount" xml:"albumCount,attr"`
}

func NewGenre(
	value string,
	songCount int,
	albumCount int,
) *Genre {
	return &Genre{
		Value:      value,
		SongCount:  songCount,
		AlbumCount: albumCount,
	}
}
