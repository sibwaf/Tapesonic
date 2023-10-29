package responses

import (
	"time"
)

type SubsonicPlaylist struct {
	Id        string `json:"id" xml:"id,attr"`
	Name      string `json:"name" xml:"name,attr"`
	SongCount int    `json:"songCount" xml:"songCount,attr"`
	Duration  int    `json:"duration" xml:"duration,attr"`
	Created   string `json:"created" xml:"created,attr"`
	Changed   string `json:"changed" xml:"changed,attr"`

	CoverArt string `json:"coverArt" xml:"coverArt,attr"`
	Owner    string `json:"owner" xml:"owner,attr"`

	Entry []SubsonicChild `json:"entry,omitempty" xml:"entry"`
}

func NewSubsonicPlaylist(
	id string,
	name string,
	songCount int,
	durationSec int,
	created time.Time,
	changed time.Time,
) *SubsonicPlaylist {
	return &SubsonicPlaylist{
		Id:        id,
		Name:      name,
		SongCount: songCount,
		Duration:  durationSec,
		Created:   created.Format(time.RFC3339),
		Changed:   changed.Format(time.RFC3339),
	}
}
