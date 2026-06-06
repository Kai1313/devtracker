package database

import (
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

//go:embed migrations/*.up.sql
var migrationFiles embed.FS

type migration struct {
	Version string
	Name    string
	Path    string
}

func RunMigrations(db *gorm.DB, log zerolog.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	tx, err := sqlDB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(32) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`); err != nil {
		return err
	}

	for _, item := range migrations {
		applied, err := isMigrationApplied(tx, item.Version)
		if err != nil {
			return err
		}

		if applied {
			continue
		}

		content, err := migrationFiles.ReadFile(item.Path)
		if err != nil {
			return err
		}

		if _, err = tx.Exec(string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", item.Path, err)
		}

		if _, err = tx.Exec(
			`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
			item.Version,
			item.Name,
		); err != nil {
			return err
		}

		log.Info().Str("version", item.Version).Str("name", item.Name).Msg("database migration applied")
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func loadMigrations() ([]migration, error) {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	migrations := make([]migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		version, name, err := parseMigrationName(entry.Name())
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, migration{
			Version: version,
			Name:    name,
			Path:    filepath.ToSlash(filepath.Join("migrations", entry.Name())),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func parseMigrationName(fileName string) (string, string, error) {
	trimmed := strings.TrimSuffix(fileName, ".up.sql")
	parts := strings.SplitN(trimmed, "_", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid migration file name: %s", fileName)
	}

	return parts[0], strings.ReplaceAll(parts[1], "_", " "), nil
}

func isMigrationApplied(tx *sql.Tx, version string) (bool, error) {
	var exists bool
	err := tx.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`,
		version,
	).Scan(&exists)

	return exists, err
}
