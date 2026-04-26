package migration

import (
	"embed"
	"fmt"
	"io"
	"log/slog"
	"path"
	"regexp"
	"slices"
	"strings"
	"tapesonic/util"
	"time"

	"gorm.io/gorm"
)

//go:embed schema/*
var scripts embed.FS
var baseDir = "schema"

var versionRegex = regexp.MustCompile(`^V(\d+)(?:_.+)\.sql$`)
var orderRegex = regexp.MustCompile(`(\d+)`)

type script struct {
	Name       string
	Version    int
	Order      int
	Statements []parsedStatement
}

func Migrate(db *gorm.DB) error {
	if err := createMigrationHistoryTable(db); err != nil {
		return err
	}

	schemaVersion := -1
	if err := db.Raw("SELECT coalesce(max(version), -1) FROM schema_history").Take(&schemaVersion).Error; err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Current database schema version: %d", schemaVersion))

	scripts, err := collectScripts()
	if err != nil {
		return err
	}

	for _, script := range scripts {
		if script.Version > 0 && script.Version <= schemaVersion {
			slog.Info(fmt.Sprintf("Migration %s: already applied", script.Name))
			continue
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			for _, statement := range script.Statements {
				err := tx.Exec(statement.Sql).Error
				if err != nil {
					return err
				}
			}

			if script.Version > 0 {
				sql := `
					INSERT INTO schema_history (name, version, applied_at)
					VALUES (@name, @version, @appliedAt)
				`
				params := map[string]any{
					"name":      script.Name,
					"version":   script.Version,
					"appliedAt": util.NewTimestampWrapper(time.Now()),
				}
				return tx.Exec(sql, params).Error
			} else {
				return nil
			}
		})
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to apply migration %s: %s", script.Name, err.Error()))
			return err
		}

		slog.Info(fmt.Sprintf("Migration %s: successfully applied", script.Name))
	}

	slog.Info("Database schema is up-to-date")

	return nil
}

func createMigrationHistoryTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_history (
			name TEXT,
			version INTEGER,
			applied_at INTEGER
		) STRICT
	`).Error
}

func collectScripts() ([]script, error) {
	files, err := scripts.ReadDir(baseDir)
	if err != nil {
		return []script{}, err
	}

	preScripts := []script{}
	versionScripts := []script{}
	postScripts := []script{}

	for _, entry := range files {
		reader, err := scripts.Open(path.Join(baseDir, entry.Name()))
		if err != nil {
			return []script{}, err
		}
		defer reader.Close()

		bytes, err := io.ReadAll(reader)
		if err != nil {
			return []script{}, err
		}

		parser := parser{
			lexer: lexer{
				runes: []rune(string(bytes)),
			},
		}

		parsedScript, err := parser.parseScript()
		if err != nil {
			return []script{}, err
		}

		s := script{
			Name:       entry.Name(),
			Statements: parsedScript.Statements,
		}

		if strings.HasPrefix(entry.Name(), "Pre") {
			s.Order = parseOrder(entry.Name())
			preScripts = append(preScripts, s)
		} else if strings.HasPrefix(entry.Name(), "Post") {
			s.Order = parseOrder(entry.Name())
			postScripts = append(postScripts, s)
		} else if strings.HasPrefix(entry.Name(), "V") {
			s.Version, err = parseVersion(entry.Name())
			if err != nil {
				return []script{}, err
			}
			s.Order = s.Version
			versionScripts = append(versionScripts, s)
		} else {
			return []script{}, fmt.Errorf("unexpected migration filename %s", entry.Name())
		}
	}

	compareScriptOrder := func(a script, b script) int { return a.Order - b.Order }
	slices.SortFunc(preScripts, compareScriptOrder)
	slices.SortFunc(versionScripts, compareScriptOrder)
	slices.SortFunc(postScripts, compareScriptOrder)

	result := []script{}
	result = append(result, preScripts...)
	result = append(result, versionScripts...)
	result = append(result, postScripts...)

	return result, nil
}

func parseVersion(text string) (int, error) {
	match := versionRegex.FindStringSubmatch(text)
	if match == nil {
		return -1, fmt.Errorf("invalid migration name '%s'", text)
	}

	version := util.StringToIntOrDefault(match[1], 0)
	if version < 1 {
		return -1, fmt.Errorf("invalid migration version '%s'", match[1])
	} else {
		return version, nil
	}
}

func parseOrder(text string) int {
	match := orderRegex.FindStringSubmatch(text)
	if match == nil {
		return -1
	} else {
		return util.StringToIntOrDefault(match[1], -1)
	}
}
