package responses

type SubsonicChild struct {
	Id        string `json:"id" xml:"id,attr"`
	IsDir     bool   `json:"isDir" xml:"isDir,attr"`
	Artist    string `json:"artist" xml:"artist,attr"`
	ArtistId  string `json:"artistId" xml:"artistId,attr"`
	Title     string `json:"title" xml:"title,attr"`
	Album     string `json:"album" xml:"album,attr"`
	Track     int    `json:"track" xml:"track,attr"`
	CoverArt  string `json:"coverArt" xml:"coverArt,attr"`
	Duration  int    `json:"duration" xml:"duration,attr"`
	PlayCount int    `json:"playCount" xml:"playCount,attr"`
	AlbumId   string `json:"albumId" xml:"albumId,attr"`
}

func NewSubsonicChild(
	id string,
	isDir bool,
	artist string,
	title string,
	track int,
	durationSec int,
) *SubsonicChild {
	return &SubsonicChild{
		Id:       id,
		IsDir:    isDir,
		Artist:   artist,
		Title:    title,
		Track:    track,
		Duration: durationSec,
	}
}
