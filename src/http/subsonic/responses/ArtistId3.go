package responses

import "time"

type ArtistId3 struct {
	Id   string `json:"id" xml:"id,attr"`
	Name string `json:"name" xml:"name,attr"`

	CoverArt       string `json:"coverArt" xml:"coverArt,attr"`
	ArtistImageUrl string `json:"artistImageUrl" xml:"artistImageUrl,attr"`

	AlbumCount int `json:"albumCount" xml:"albumCount,attr"`

	Starred *time.Time `json:"starred" xml:"starred,attr,omitempty"`
}

func NewArtistId3(
	id string,
	name string,
) *ArtistId3 {
	return &ArtistId3{
		Id:   id,
		Name: name,
	}
}
