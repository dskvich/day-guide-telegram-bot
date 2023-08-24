package database

import (
	"database/sql"
	"embed"
	"fmt"

	migrate "github.com/rubenv/sql-migrate"
	_ "modernc.org/sqlite"
)

const dbName = "day-guide.db"

//go:embed migrations/*.sql
var migrationsFS embed.FS

func NewSQLite() (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return nil, fmt.Errorf("connecting to db: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("verifying connection: %v", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("running migrations: %v", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	source := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationsFS,
		Root:       "migrations",
	}
	if _, err := migrate.Exec(db, "sqlite3", source, migrate.Up); err != nil {
		return err
	}
	return nil
}
