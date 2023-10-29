package responses

type ArtistId3 struct {
	Id   string `json:"id" xml:"id,attr"`
	Name string `json:"name" xml:"name,attr"`
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
