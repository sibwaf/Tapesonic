package artists

import (
	"math/rand"
	"slices"
	"tapesonic/util"

	"github.com/google/uuid"
)

type ArtistService struct {
	artists *artistStorage
}

func newArtistService(artists *artistStorage) *ArtistService {
	return &ArtistService{
		artists: artists,
	}
}

func (svc *ArtistService) Search(query string, count int, offset int) ([]Artist, error) {
	return svc.artists.SearchByQuery(query, count, offset)
}

func (svc *ArtistService) FindAllMatches(name string, mbid string) ([]Artist, error) {
	if mbid != "" {
		artist, err := svc.artists.FindByMusicBrainzId(mbid)
		if err != nil {
			return []Artist{}, err
		}

		if artist != nil {
			return []Artist{*artist}, nil
		}
	}

	options, err := svc.artists.SearchByName(name)
	if err != nil {
		return []Artist{}, err
	}

	options = slices.DeleteFunc(options, func(option Artist) bool {
		for _, optionName := range getAllNames(option) {
			if util.MatchText(name, optionName) {
				return false
			}
		}
		return true
	})

	return options, nil
}

func (svc *ArtistService) GetById(id uuid.UUID) (Artist, error) {
	return svc.artists.GetById(id)
}

// todo: probably needs a mutex to prevent duplicates on concurrent operations
func (svc *ArtistService) Create(name string, mbid string) (Artist, error) {
	id := uuid.New()

	createdArtist, err := svc.artists.CreateOrGet(id, name, []string{}, mbid)
	if err != nil {
		return Artist{}, err
	}

	if createdArtist.Id != id {
		return svc.updateAliases(createdArtist, name)
	}

	return createdArtist, nil
}

func (svc *ArtistService) Update(id uuid.UUID, name string, aliases []string, mbid string) (Artist, error) {
	return svc.artists.Update(id, name, aliases, mbid)
}

func (svc *ArtistService) FindMatchOrCreate(name string, mbid string) (Artist, error) {
	existingArtists, err := svc.FindAllMatches(name, mbid)
	if err != nil {
		return Artist{}, err
	}

	if len(existingArtists) == 0 {
		return svc.Create(name, mbid)
	}

	if len(existingArtists) == 1 {
		artist := existingArtists[0]
		if mbid != "" && artist.MusicBrainzId != nil {
			return svc.updateAliases(artist, name)
		} else {
			return artist, nil
		}
	}

	if mbid != "" {
		// if we have multiple options with a matching name,
		// they didn't match the mbid and this mbid is not in our library;
		// create a new artist so the user decides which artist it is;
		// this should be pretty rare, so it shouldn't be too painful
		return svc.Create(name, mbid)
	}

	// we don't really have many other options left if multiple artists match
	// we could create a new one but this will lead to a lot of copies,
	// so we just choose a random one and hope for the best

	index := rand.Int() % len(existingArtists)
	return existingArtists[index], nil
}

func (svc *ArtistService) updateAliases(artist Artist, name string) (Artist, error) {
	isNewAlias := true
	for _, existingName := range getAllNames(artist) {
		isNewAlias = isNewAlias && !util.MatchText(name, existingName)
	}

	if isNewAlias {
		artist.Aliases = append(artist.Aliases, name)

		mbid := ""
		if artist.MusicBrainzId != nil {
			mbid = *artist.MusicBrainzId
		}

		return svc.artists.Update(artist.Id, artist.Name, artist.Aliases, mbid)
	} else {
		return artist, nil
	}
}

func getAllNames(artist Artist) []string {
	allNames := []string{artist.Name}
	allNames = append(allNames, artist.Aliases...)
	return allNames
}
