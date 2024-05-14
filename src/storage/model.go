package storage

type RelatedItems struct {
	Tapes     []*Tape
	Playlists []*Playlist
	Albums    []*Album
}

type SubsonicAlbumItem struct {
	Album

	SongCount   int
	DurationSec int

	Tracks []SubsonicTrackItem `gorm:"-"`
}

type SubsonicTrackItem struct {
	AlbumTrack

	Artist string
	Title  string

	DurationSec int
}
