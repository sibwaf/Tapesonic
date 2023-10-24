package responses

type SubsonicPlaylists struct {
	Playlist []SubsonicPlaylist `json:"playlist" xml:"playlist"`
}

func NewSubsonicPlaylists(playlists []SubsonicPlaylist) *SubsonicPlaylists {
	return &SubsonicPlaylists{
		Playlist: playlists,
	}
}
