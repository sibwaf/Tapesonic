package responses

type SearchResult3 struct {
	Artist []ArtistId3     `json:"artist" xml:"artist"`
	Album  []AlbumId3      `json:"album" xml:"album"`
	Song   []SubsonicChild `json:"song" xml:"song"`
}

func NewSearchResult3(
	artists []ArtistId3,
	albums []AlbumId3,
	songs []SubsonicChild,
) *SearchResult3 {
	return &SearchResult3{
		Artist: artists,
		Album:  albums,
		Song:   songs,
	}
}
