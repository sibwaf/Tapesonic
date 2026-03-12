package artists

import (
	"encoding/json"
	"errors"
	"fmt"
	"tapesonic/model"
	"tapesonic/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type artistStorage struct {
	db *gorm.DB
}

func newArtistStorage(db *gorm.DB) *artistStorage {
	return &artistStorage{db: db}
}

func (store *artistStorage) PrepareDatabase() error {
	return store.db.AutoMigrate(&Artist{})
}

func (store *artistStorage) CreateOrGet(id uuid.UUID, name string, aliases []string, musicBrainzId string) (Artist, error) {
	aliasesJson, err := json.Marshal(aliases)
	if err != nil {
		return Artist{}, err
	}

	result := Artist{}

	return result, store.db.Transaction(func(tx *gorm.DB) error {
		sql := `
			INSERT INTO artists (id, name, aliases, search_name, music_brainz_id)
			VALUES (@id, @name, @aliases, @searchName, @musicBrainzId)
			ON CONFLICT (music_brainz_id) DO UPDATE
			SET music_brainz_id = artists.music_brainz_id
			RETURNING id, name, aliases, music_brainz_id
		`
		params := map[string]any{
			"id":            id,
			"name":          name,
			"aliases":       string(aliasesJson),
			"searchName":    util.MakeTextSearchString(name, aliases...),
			"musicBrainzId": util.TakeIf(&musicBrainzId, musicBrainzId != ""),
		}
		if err := store.db.Raw(sql, params).Take(&result).Error; err != nil {
			return err
		}

		return nil
	})
}

func (store *artistStorage) Update(id uuid.UUID, name string, aliases []string, musicBrainzId string) (Artist, error) {
	aliasesJson, err := json.Marshal(aliases)
	if err != nil {
		return Artist{}, err
	}

	sql := `
		UPDATE artists
		SET name = @name, aliases = @aliases, search_name = @searchName, music_brainz_id = @musicBrainzId
		WHERE id = @id
		RETURNING id, name, aliases, music_brainz_id
	`
	params := map[string]any{
		"id":            id,
		"name":          name,
		"aliases":       string(aliasesJson),
		"musicBrainzId": util.TakeIf(&musicBrainzId, musicBrainzId != ""),
		"searchName":    util.MakeTextSearchString(name, aliases...),
	}

	result := Artist{}
	err = store.db.Raw(sql, params).Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Artist{}, model.ErrNotFound
	} else {
		return result, err
	}
}

func (store *artistStorage) ReplaceUsages(oldIds []uuid.UUID, newId uuid.UUID) error {
	if len(oldIds) == 0 {
		return nil
	}

	return store.db.Transaction(func(tx *gorm.DB) error {
		params := map[string]any{
			"oldIds": oldIds,
			"newId":  newId,
		}

		if err := tx.Exec("UPDATE remote_artists SET tapesonic_artist_id = @newId WHERE tapesonic_artist_id IN @oldIds", params).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE source_tracks SET artist_id = @newId WHERE artist_id IN @oldIds", params).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE tapes SET artist_id = @newId WHERE artist_id IN @oldIds", params).Error; err != nil {
			return err
		}

		return nil
	})
}

func (store *artistStorage) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	return store.db.Exec("DELETE FROM artists WHERE id IN @ids", map[string]any{"ids": ids}).Error
}

func (store *artistStorage) SearchByQuery(query string, count int, offset int) ([]Artist, error) {
	filter := util.MakeTextSearchCondition([]string{"artists.search_name"}, query)
	if filter == "" {
		filter = "1 = 1"
	}

	sql := fmt.Sprintf(
		`
			SELECT artists.id, artists.name, artists.aliases, artists.music_brainz_id
			FROM artists
			WHERE %s
			ORDER BY lower(artists.name)
			LIMIT %d OFFSET %d
		`,
		filter,
		count,
		offset,
	)

	result := []Artist{}
	return result, store.db.Raw(sql).Find(&result).Error
}

func (store *artistStorage) GetById(id uuid.UUID) (Artist, error) {
	sql := `
		SELECT artists.id, artists.name, artists.aliases, artists.music_brainz_id
		FROM artists
		WHERE id = @id
	`
	params := map[string]any{"id": id}

	result := Artist{}
	err := store.db.Raw(sql, params).Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Artist{}, model.ErrNotFound
	} else {
		return result, err
	}
}

func (store *artistStorage) SearchByName(name string) ([]Artist, error) {
	filter := util.MakeTextSearchCondition([]string{"artists.search_name"}, name)
	if filter == "" {
		return []Artist{}, nil
	}

	sql := fmt.Sprintf(
		`
			SELECT artists.id, artists.name, artists.aliases, artists.music_brainz_id
			FROM artists
			WHERE %s
			ORDER BY length(artists.name)
		`,
		filter,
	)

	result := []Artist{}
	return result, store.db.Raw(sql).Find(&result).Error
}

func (store *artistStorage) FindByMusicBrainzId(musicBrainzId string) (*Artist, error) {
	sql := `
		SELECT artists.id, artists.name, artists.aliases, artists.music_brainz_id
		FROM artists
		WHERE music_brainz_id = @musicBrainzId
	`
	params := map[string]any{"musicBrainzId": musicBrainzId}

	result := Artist{}
	err := store.db.Raw(sql, params).Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else {
		return &result, err
	}
}
