package responses

type AlbumList2 struct {
	Album []AlbumId3 `json:"album" xml:"album"`
}

func NewAlbumList2(albums []AlbumId3) *AlbumList2 {
	return &AlbumList2{
		Album: albums,
	}
}
