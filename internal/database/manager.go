package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/btassone/obtura/internal/config"
	"github.com/btassone/obtura/internal/database/migrations"
	"github.com/btassone/obtura/internal/database/seeders"
	"github.com/btassone/obtura/pkg/database"
)

// Manager handles database operations
type Manager struct {
	db              *database.DB
	migrationRunner *database.MigrationRunner
	seederRunner    *database.SeederRunner
}

// NewManager creates a new database manager
func NewManager() (*Manager, error) {
	// Get database configuration
	dbConfig := config.GetDatabaseConfig()

	// Ensure data directory exists for SQLite
	if dbConfig.Driver == "sqlite" || dbConfig.Driver == "sqlite3" {
		dir := filepath.Dir(dbConfig.SQLitePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	// Connect to database
	db, err := database.New(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set driver name for migrations
	migrations.SetDriverName(dbConfig.Driver)

	// Create migration runner
	migrationRunner := database.NewMigrationRunner(db)
	for _, migration := range migrations.GetMigrations() {
		migrationRunner.AddMigration(migration)
	}

	// Create seeder runner
	seederRunner := database.NewSeederRunner(db)
	for _, seeder := range seeders.GetSeeders() {
		seederRunner.AddSeeder(seeder)
	}

	return &Manager{
		db:              db,
		migrationRunner: migrationRunner,
		seederRunner:    seederRunner,
	}, nil
}

// DB returns the database connection
func (m *Manager) DB() *database.DB {
	return m.db
}

// Migrate runs all pending migrations
func (m *Manager) Migrate() error {
	fmt.Println("Running database migrations...")
	if err := m.migrationRunner.Run(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	fmt.Println("Migrations completed successfully!")
	return nil
}

// Rollback rolls back migrations
func (m *Manager) Rollback(steps int) error {
	fmt.Printf("Rolling back %d migration(s)...\n", steps)
	if err := m.migrationRunner.Rollback(steps); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}
	fmt.Println("Rollback completed successfully!")
	return nil
}

// MigrationStatus returns the status of all migrations
func (m *Manager) MigrationStatus() ([]database.MigrationStatus, error) {
	return m.migrationRunner.Status()
}

// Seed runs database seeders
func (m *Manager) Seed(specific ...string) error {
	if len(specific) > 0 {
		fmt.Printf("Running specific seeders: %v\n", specific)
		return m.seederRunner.RunSpecific(specific...)
	}

	fmt.Println("Running all seeders...")
	if err := m.seederRunner.Run(); err != nil {
		return fmt.Errorf("seeding failed: %w", err)
	}
	fmt.Println("Seeding completed successfully!")
	return nil
}

// ListSeeders returns information about available seeders
func (m *Manager) ListSeeders() []database.SeederInfo {
	return m.seederRunner.List()
}

// Close closes the database connection
func (m *Manager) Close() error {
	return m.db.Close()
}

// Query creates a new query builder
func (m *Manager) Query() *database.Query {
	return database.NewQuery(m.db)
}