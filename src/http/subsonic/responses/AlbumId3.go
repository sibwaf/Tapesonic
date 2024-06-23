package responses

import "time"

type AlbumId3 struct {
	Id string `json:"id" xml:"id,attr"`

	Name   string `json:"name" xml:"name,attr"`
	Artist string `json:"artist" xml:"artist,attr"`

	CoverArt string `json:"coverArt" xml:"coverArt,attr"`

	SongCount int `json:"songCount" xml:"songCount,attr"`
	Duration  int `json:"duration" xml:"duration,attr"`
	PlayCount int `json:"playCount" xml:"playCount,attr"`

	Created time.Time `json:"created" xml:"created,attr"`

	Year int `json:"year" xml:"year,attr"`

	Song []SubsonicChild `json:"song,omitempty" xml:"song"`
}

func NewAlbumId3(
	id string,
	name string,
	artist string,
	coverArt string,
	songCount int,
	durationSec int,
	created time.Time,
) *AlbumId3 {
	return &AlbumId3{
		Id:        id,
		Name:      name,
		Artist:    artist,
		CoverArt:  coverArt,
		SongCount: songCount,
		Duration:  durationSec,
		Created:   created,
	}
}
