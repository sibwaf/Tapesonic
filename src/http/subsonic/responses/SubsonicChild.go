package responses

type SubsonicChild struct {
	Id       string `json:"id" xml:"id,attr"`
	IsDir    bool   `json:"isDir" xml:"isDir,attr"`
	Artist   string `json:"artist" xml:"artist,attr"`
	Title    string `json:"title" xml:"title,attr"`
	Duration int    `json:"duration" xml:"duration,attr"`
}

func NewSubsonicChild(
	id string,
	isDir bool,
	artist string,
	title string,
	durationSec int,
) *SubsonicChild {
	return &SubsonicChild{
		Id:       id,
		IsDir:    isDir,
		Artist:   artist,
		Title:    title,
		Duration: durationSec,
	}
}
