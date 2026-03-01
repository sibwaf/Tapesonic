package subsonic

import (
	"maps"
	"net/http"
	"slices"
	"strings"

	"tapesonic/library"
	"tapesonic/subsonic"
)

func GetArtists(auth *authenticator, artists *library.LibraryService) SubsonicHandler {
	return func(r *http.Request) (*subsonic.Response, error) {
		user, err := auth.Authenticate(r)
		if err != nil {
			return nil, err
		}

		// todo: limit/offset?
		allArtists, err := artists.GetArtistsSortByName(user, 9999, 0)
		if err != nil {
			return nil, err
		}

		indexes := map[string]subsonic.IndexId3{}
		for _, artist := range allArtists {
			indexName := strings.TrimSpace(artist.Name)
			if len(indexName) > 0 {
				indexName = strings.ToUpper(string([]rune(indexName)[0]))
			}

			index, ok := indexes[indexName]
			if !ok {
				index = subsonic.IndexId3{
					Name:   indexName,
					Artist: []subsonic.ArtistId3{},
				}
			}

			index.Artist = append(index.Artist, ToArtistId3(artist))
			indexes[indexName] = index
		}

		sortedIndexes := slices.Collect(maps.Values(indexes))
		slices.SortFunc(
			sortedIndexes,
			func(a subsonic.IndexId3, b subsonic.IndexId3) int {
				return strings.Compare(a.Name, b.Name)
			},
		)

		response := subsonic.NewOkResponse()
		response.Artists = &subsonic.Artists{
			IgnoredArticles: "",
			Index:           sortedIndexes,
		}
		return response, nil
	}
}
