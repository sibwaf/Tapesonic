package responses

type MusicFolders struct {
	MusicFolder []MusicFolder `json:"musicFolder" xml:"musicFolder"`
}

func NewMusicFolders(musicFolders []MusicFolder) *MusicFolders {
	return &MusicFolders{
		MusicFolder: musicFolders,
	}
}
