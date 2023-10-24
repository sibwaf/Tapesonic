package responses

type SubsonicChild struct {
	Id       string `json:"id" xml:"id,attr"`
	IsDir    bool   `json:"isDir" xml:"isDir,attr"`
	Title    string `json:"title" xml:"title,attr"`
	Duration int    `json:"duration" xml:"duration,attr"`
}

func NewSubsonicChild(
	id string,
	isDir bool,
	title string,
	durationSec int,
) *SubsonicChild {
	return &SubsonicChild{
		Id:       id,
		IsDir:    isDir,
		Title:    title,
		Duration: durationSec,
	}
}
