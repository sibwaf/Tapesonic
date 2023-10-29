package responses

type MusicFolder struct {
	Id   string `json:"id" xml:"id,attr"`
	Name string `json:"name" xml:"name,attr"`
}

func NewMusicFolder(
	id string,
	name string,
) *MusicFolder {
	return &MusicFolder{
		Id:   id,
		Name: name,
	}
}
