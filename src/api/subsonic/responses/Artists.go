package responses

type Artists struct {
	IgnoredArticles string     `json:"ignoredArticles" xml:"ignoredArticles,attr"`
	Index           []IndexId3 `json:"index" xml:"index"`
}

func NewArtists(
	ignoredArticles string,
	index []IndexId3,
) *Artists {
	return &Artists{
		IgnoredArticles: ignoredArticles,
		Index:           index,
	}
}
