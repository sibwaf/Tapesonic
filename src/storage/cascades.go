package storage

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

const (
	action_delete = iota
	action_set_null
)

type cascade struct {
	ReferencingTable string
	ReferencedTable  string

	Keys []cascadeField

	OnConstraint int
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
			OnConstraint:     action_delete,
		},
		{
			ReferencingTable: "all_track_ids",
			ReferencedTable:  "remote_tracks",
			Keys:             []cascadeField{{ReferencingField: "remote_track_id", ReferencedField: "id"}},
			OnConstraint:     action_delete,
		},
		{
			ReferencingTable: "tape_to_tracks",
			ReferencedTable:  "all_track_ids",
			Keys:             []cascadeField{{ReferencingField: "track_id", ReferencedField: "id"}},
			OnConstraint:     action_delete,
		},
		{
			ReferencingTable: "listen_stats",
			ReferencedTable:  "all_track_ids",
			Keys:             []cascadeField{{ReferencingField: "track_id", ReferencedField: "id"}},
			OnConstraint:     action_delete,
		},
		{
			ReferencingTable: "recommended_playlist_tracks",
			ReferencedTable:  "all_track_ids",
			Keys:             []cascadeField{{ReferencingField: "track_id", ReferencedField: "id"}},
			OnConstraint:     action_delete,
		},
		{
			ReferencingTable: "tapes",
			ReferencedTable:  "thumbnails",
			Keys:             []cascadeField{{ReferencingField: "thumbnail_id", ReferencedField: "id"}},
			OnConstraint:     action_set_null,
		},
		{
			ReferencingTable: "tapes",
			ReferencedTable:  "remote_covers",
			Keys:             []cascadeField{{ReferencingField: "thumbnail_id", ReferencedField: "id"}},
			OnConstraint:     action_set_null,
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

		var actionSql string
		switch cascade.OnConstraint {
		case action_delete:
			actionSql = fmt.Sprintf(
				"DELETE FROM %s WHERE %s",
				cascade.ReferencingTable,
				strings.Join(keyConditions, " AND "),
			)
		case action_set_null:
			fieldAssignments := []string{}
			for _, key := range cascade.Keys {
				fieldAssignments = append(fieldAssignments, fmt.Sprintf("%s = NULL", key.ReferencingField))
			}

			actionSql = fmt.Sprintf(
				"UPDATE %s SET %s WHERE %s",
				cascade.ReferencingTable,
				strings.Join(fieldAssignments, ", "),
				strings.Join(keyConditions, " AND "),
			)
		default:
			panic(fmt.Sprintf("unknown cascade action: %d", cascade.OnConstraint))
		}

		sql := fmt.Sprintf(
			`
			CREATE TRIGGER %s
			BEFORE DELETE ON %s FOR EACH ROW
			BEGIN
				%s;
			END
			`,
			name,
			cascade.ReferencedTable,
			actionSql,
		)

		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}

	return nil
}
