package storage

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type cascade struct {
	ReferencingTable string
	ReferencedTable  string

	Keys []cascadeField
}

type cascadeField struct {
	ReferencingField string
	ReferencedField  string
}

func DropCascades(db *gorm.DB) error {
	type trigger struct {
		Name string
	}

	triggers := []trigger{}
	err := db.Raw("SELECT name FROM sqlite_schema WHERE type = 'trigger' AND name LIKE 'deletecascade_%'").Find(&triggers).Error
	if err != nil {
		return err
	}

	for _, trigger := range triggers {
		if err := db.Exec(fmt.Sprintf("DROP TRIGGER %s", trigger.Name)).Error; err != nil {
			return err
		}
	}

	return nil
}

func CreateCascades(db *gorm.DB) error {
	config := []cascade{
		{
			ReferencingTable: "all_track_ids",
			ReferencedTable:  "source_tracks",
			Keys:             []cascadeField{{ReferencingField: "source_track_id", ReferencedField: "id"}},
		},
		{
			ReferencingTable: "all_track_ids",
			ReferencedTable:  "remote_tracks",
			Keys:             []cascadeField{{ReferencingField: "remote_track_id", ReferencedField: "id"}},
		},
		{
			ReferencingTable: "listen_stats",
			ReferencedTable:  "all_track_ids",
			Keys:             []cascadeField{{ReferencingField: "track_id", ReferencedField: "id"}},
		},
		{
			ReferencingTable: "recommended_playlist_tracks",
			ReferencedTable:  "all_track_ids",
			Keys:             []cascadeField{{ReferencingField: "track_id", ReferencedField: "id"}},
		},
	}

	for i, cascade := range config {
		name := fmt.Sprintf("deletecascade_%d_%s", i, cascade.ReferencingTable)

		keyConditions := []string{}
		for _, key := range cascade.Keys {
			referencingColumn := fmt.Sprintf("%s.%s", cascade.ReferencingTable, key.ReferencingField)
			referencedColumn := fmt.Sprintf("old.%s", key.ReferencedField)
			keyConditions = append(keyConditions, fmt.Sprintf("%s = %s", referencingColumn, referencedColumn))
		}

		sql := fmt.Sprintf(
			`
			CREATE TRIGGER %s
			BEFORE DELETE ON %s FOR EACH ROW
			BEGIN
				DELETE FROM %s WHERE %s;
			END
			`,
			name,
			cascade.ReferencedTable,
			cascade.ReferencingTable,
			strings.Join(keyConditions, " AND "),
		)

		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}

	return nil
}
