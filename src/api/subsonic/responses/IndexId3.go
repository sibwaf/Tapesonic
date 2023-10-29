package responses

type IndexId3 struct {
	Name   string      `json:"name" xml:"name,attr"`
	Artist []ArtistId3 `json:"artist" xml:"artist"`
}

func NewIndexId3(
	name string,
	artists []ArtistId3,
) *IndexId3 {
	return &IndexId3{
		Name:   name,
		Artist: artists,
	}
}
