package data

import (
	"embed"

	migrate "github.com/rubenv/sql-migrate"
)

//go:embed ../../migrations/*.sql
var migFS embed.FS

// Migrate applies all *Up* migrations in embedFS.
func Migrate(db *DB, driver string) error {
	src := migrate.EmbedFileSystemMigrationSource{
		FileSystem: migFS,
		Root:       ".",
	}
	// driver must be "mysql" for MariaDB, "postgres" for Cockroach
	_, err := migrate.Exec(db.DB, driver, src, migrate.Up)
	return err
}
