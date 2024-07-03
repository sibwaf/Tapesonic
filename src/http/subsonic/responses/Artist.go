package responses

import "time"

type Artist struct {
	Id   string `json:"id" xml:"id,attr"`
	Name string `json:"name" xml:"name,attr"`

	ArtistImageUrl string `json:"artistImageUrl" xml:"artistImageUrl,attr"`

	Starred *time.Time `json:"starred" xml:"starred,attr,omitempty"`

	UserRating    int     `json:"userRating" xml:"userRating,attr"`
	AverageRating float32 `json:"averageRating" xml:"averageRating,attr"`

	Album []AlbumId3 `json:"album" xml:"album"`
}

func NewArtist(
	id string,
	name string,
) *Artist {
	return &Artist{
		Id:   id,
		Name: name,
	}
}
