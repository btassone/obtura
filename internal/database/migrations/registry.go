package migrations

import (
	"github.com/btassone/obtura/pkg/database"
)

var (
	// migrations holds all registered migrations
	migrations []database.Migration
	
	// DriverName is set by the migration runner to indicate which database is being used
	DriverName string
)

// RegisterMigration adds a migration to the registry
func RegisterMigration(migration *database.Migration) {
	migrations = append(migrations, *migration)
}

// GetMigrations returns all registered migrations
func GetMigrations() []database.Migration {
	return migrations
}

// SetDriverName sets the database driver name for migrations to use
func SetDriverName(driver string) {
	DriverName = driver
}