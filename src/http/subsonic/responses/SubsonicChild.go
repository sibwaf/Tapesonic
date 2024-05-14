package responses

type SubsonicChild struct {
	Id        string `json:"id" xml:"id,attr"`
	IsDir     bool   `json:"isDir" xml:"isDir,attr"`
	Artist    string `json:"artist" xml:"artist,attr"`
	Title     string `json:"title" xml:"title,attr"`
	Track     int    `json:"track" xml:"track,attr"`
	Duration  int    `json:"duration" xml:"duration,attr"`
	PlayCount int    `json:"playCount" xml:"playCount,attr"`
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
