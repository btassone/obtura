package database

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     string
	Description string
	Up          func(*sql.Tx) error
	Down        func(*sql.Tx) error
}

// MigrationRunner handles running migrations
type MigrationRunner struct {
	db         *DB
	migrations []Migration
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *DB) *MigrationRunner {
	return &MigrationRunner{
		db:         db,
		migrations: make([]Migration, 0),
	}
}

// AddMigration adds a migration to the runner
func (r *MigrationRunner) AddMigration(migration Migration) {
	r.migrations = append(r.migrations, migration)
}

// Run executes all pending migrations
func (r *MigrationRunner) Run() error {
	// Ensure migrations table exists
	if err := r.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	applied, err := r.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Sort migrations by version
	sort.Slice(r.migrations, func(i, j int) bool {
		return r.migrations[i].Version < r.migrations[j].Version
	})

	// Run pending migrations
	for _, migration := range r.migrations {
		if _, ok := applied[migration.Version]; ok {
			continue // Already applied
		}

		if err := r.runMigration(migration); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration.Version, err)
		}
	}

	return nil
}

// Rollback rolls back the last n migrations
func (r *MigrationRunner) Rollback(steps int) error {
	applied, err := r.getAppliedMigrationsOrdered()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if steps > len(applied) {
		steps = len(applied)
	}

	// Get migrations to rollback
	toRollback := applied[len(applied)-steps:]

	// Find migration definitions
	migrationMap := make(map[string]Migration)
	for _, m := range r.migrations {
		migrationMap[m.Version] = m
	}

	// Rollback migrations
	for i := len(toRollback) - 1; i >= 0; i-- {
		version := toRollback[i]
		migration, ok := migrationMap[version]
		if !ok {
			return fmt.Errorf("migration %s not found", version)
		}

		if err := r.rollbackMigration(migration); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", version, err)
		}
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func (r *MigrationRunner) createMigrationsTable() error {
	var query string

	// Adjust for different databases
	switch r.db.Driver() {
	case "sqlite", "sqlite3":
		query = `CREATE TABLE IF NOT EXISTS migrations (
			version TEXT PRIMARY KEY,
			description TEXT,
			checksum TEXT,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "mysql":
		query = `
			CREATE TABLE IF NOT EXISTS migrations (
				version VARCHAR(255) PRIMARY KEY,
				description VARCHAR(255),
				checksum VARCHAR(32),
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
	case "postgres", "postgresql":
		query = `
			CREATE TABLE IF NOT EXISTS migrations (
				version VARCHAR(255) PRIMARY KEY,
				description VARCHAR(255),
				checksum VARCHAR(32),
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
	default:
		query = `
			CREATE TABLE IF NOT EXISTS migrations (
				version VARCHAR(255) PRIMARY KEY,
				description VARCHAR(255),
				checksum VARCHAR(32),
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
	}

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	return nil
}

// getAppliedMigrations returns a map of applied migration versions
func (r *MigrationRunner) getAppliedMigrations() (map[string]bool, error) {
	rows, err := r.db.Query("SELECT version FROM migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// getAppliedMigrationsOrdered returns a slice of applied migration versions in order
func (r *MigrationRunner) getAppliedMigrationsOrdered() ([]string, error) {
	rows, err := r.db.Query("SELECT version FROM migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applied []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied = append(applied, version)
	}

	return applied, rows.Err()
}

// runMigration executes a single migration
func (r *MigrationRunner) runMigration(migration Migration) error {
	return r.db.Transaction(func(tx *sql.Tx) error {
		// Run migration
		if err := migration.Up(tx); err != nil {
			return err
		}

		// Calculate checksum
		checksum := calculateChecksum(migration)

		// Record migration
		_, err := tx.Exec(
			"INSERT INTO migrations (version, description, checksum) VALUES (?, ?, ?)",
			migration.Version, migration.Description, checksum,
		)
		return err
	})
}

// rollbackMigration rolls back a single migration
func (r *MigrationRunner) rollbackMigration(migration Migration) error {
	if migration.Down == nil {
		return fmt.Errorf("migration %s does not support rollback", migration.Version)
	}

	return r.db.Transaction(func(tx *sql.Tx) error {
		// Run rollback
		if err := migration.Down(tx); err != nil {
			return err
		}

		// Remove migration record
		_, err := tx.Exec("DELETE FROM migrations WHERE version = ?", migration.Version)
		return err
	})
}

// calculateChecksum calculates a checksum for a migration
func calculateChecksum(migration Migration) string {
	data := fmt.Sprintf("%s:%s", migration.Version, migration.Description)
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

// Status returns the status of migrations
func (r *MigrationRunner) Status() ([]MigrationStatus, error) {
	applied, err := r.getAppliedMigrations()
	if err != nil {
		return nil, err
	}

	// Sort migrations
	sort.Slice(r.migrations, func(i, j int) bool {
		return r.migrations[i].Version < r.migrations[j].Version
	})

	var status []MigrationStatus
	for _, m := range r.migrations {
		s := MigrationStatus{
			Version:     m.Version,
			Description: m.Description,
			Applied:     applied[m.Version],
		}
		status = append(status, s)
	}

	return status, nil
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version     string
	Description string
	Applied     bool
	AppliedAt   *time.Time
}