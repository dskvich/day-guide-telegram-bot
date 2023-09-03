package database

import (
	"database/sql"
	"embed"
	"fmt"

	migrate "github.com/rubenv/sql-migrate"
	_ "modernc.org/sqlite"
)

const (
	dbName = "day-guide.db"
	// Connect to the database with some sane settings:
	// - No shared-cache: it's obsolete; WAL journal mode is a better solution.
	// - No foreign key constraints: it's currently disabled by default, but it's a
	// good practice to be explicit and prevent future surprises on SQLite upgrades.
	// - Journal mode set to WAL: it's the recommended journal mode for most applications
	// as it prevents locking issues.
	//
	// Notes:
	// - When using the `modernc.org/sqlite` driver, each pragma must be prefixed with `_pragma=`.
	//
	// References:
	// - https://pkg.go.dev/modernc.org/sqlite#Driver.Open
	// - https://www.sqlite.org/sharedcache.html
	// - https://www.sqlite.org/pragma.html
	settings = "?_pragma=foreign_keys(0)&_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func NewSQLite() (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbName+settings)
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
