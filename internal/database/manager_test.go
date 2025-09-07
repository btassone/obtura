package database

import (
	"path/filepath"
	"testing"

	"github.com/btassone/obtura/internal/config"
	"github.com/btassone/obtura/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	// This test requires mocking config.GetDatabaseConfig()
	// For now we'll test with SQLite in memory
	t.Run("sqlite in memory", func(t *testing.T) {
		// Mock the database config
		originalGetConfig := config.GetDatabaseConfig
		defer func() { config.GetDatabaseConfig = originalGetConfig }()

		config.GetDatabaseConfig = func() *database.Config {
			return &database.Config{
				Driver:     "sqlite",
				SQLitePath: ":memory:",
			}
		}

		manager, err := NewManager()
		require.NoError(t, err)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.db)
		assert.NotNil(t, manager.migrationRunner)
		assert.NotNil(t, manager.seederRunner)

		// Cleanup
		manager.Close()
	})

	t.Run("sqlite with file path", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "data", "test.db")

		originalGetConfig := config.GetDatabaseConfig
		defer func() { config.GetDatabaseConfig = originalGetConfig }()

		config.GetDatabaseConfig = func() *database.Config {
			return &database.Config{
				Driver:     "sqlite",
				SQLitePath: dbPath,
			}
		}

		manager, err := NewManager()
		require.NoError(t, err)
		assert.NotNil(t, manager)

		// Verify directory was created
		assert.DirExists(t, filepath.Dir(dbPath))

		// Cleanup
		manager.Close()
	})
}

func TestManager_DB(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	db := manager.DB()
	assert.NotNil(t, db)
}

func TestManager_Migrate(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	// Run migrations
	err := manager.Migrate()
	require.NoError(t, err)

	// Verify migrations table exists
	gormDB := manager.db.GetGorm()
	var exists bool
	err = gormDB.Raw("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='migration_versions')").Scan(&exists).Error
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestManager_MigrationStatus(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	// Get initial status
	status, err := manager.MigrationStatus()
	require.NoError(t, err)
	assert.NotEmpty(t, status)

	// All migrations should be pending initially
	for _, s := range status {
		assert.False(t, s.Applied)
	}

	// Run migrations
	err = manager.Migrate()
	require.NoError(t, err)

	// Check status again
	status, err = manager.MigrationStatus()
	require.NoError(t, err)

	// All migrations should be applied now
	for _, s := range status {
		assert.True(t, s.Applied)
	}
}

func TestManager_Rollback(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	// Run migrations first
	err := manager.Migrate()
	require.NoError(t, err)

	// Get status before rollback
	statusBefore, err := manager.MigrationStatus()
	require.NoError(t, err)
	appliedCount := 0
	for _, s := range statusBefore {
		if s.Applied {
			appliedCount++
		}
	}

	// Rollback 1 migration
	err = manager.Rollback(1)
	require.NoError(t, err)

	// Verify one migration was rolled back
	statusAfter, err := manager.MigrationStatus()
	require.NoError(t, err)
	newAppliedCount := 0
	for _, s := range statusAfter {
		if s.Applied {
			newAppliedCount++
		}
	}

	assert.Equal(t, appliedCount-1, newAppliedCount)
}

func TestManager_Seed(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	// Run migrations first
	err := manager.Migrate()
	require.NoError(t, err)

	// Run seeders
	err = manager.Seed()
	require.NoError(t, err)

	// TODO: Verify data was seeded
	// This would require checking specific tables
}

func TestManager_ListSeeders(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	seeders := manager.ListSeeders()
	assert.NotEmpty(t, seeders)

	// Verify seeder info
	for _, seeder := range seeders {
		assert.NotEmpty(t, seeder.Name)
		assert.NotEmpty(t, seeder.Description)
		assert.GreaterOrEqual(t, seeder.Priority, 0)
	}
}

func TestManager_Query(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	query := manager.Query()
	assert.NotNil(t, query)

	// Test that query builder works
	// This would require running migrations first
	err := manager.Migrate()
	require.NoError(t, err)

	// Try a simple query
	var count int64
	err = query.Table("users").Count(&count).Error()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(0))
}

// Helper function to create a test manager
func createTestManager(t *testing.T) *Manager {
	// Mock the database config
	originalGetConfig := config.GetDatabaseConfig
	t.Cleanup(func() { config.GetDatabaseConfig = originalGetConfig })

	config.GetDatabaseConfig = func() *database.Config {
		return &database.Config{
			Driver:     "sqlite",
			SQLitePath: ":memory:",
		}
	}

	manager, err := NewManager()
	require.NoError(t, err)

	return manager
}

func TestManager_TransactionSupport(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()

	// Run migrations
	err := manager.Migrate()
	require.NoError(t, err)

	db := manager.DB().GetGorm()

	// Test transaction rollback
	err = db.Transaction(func(tx *database.Gorm) error {
		// Try to create a user
		err := tx.Exec("INSERT INTO users (name, email, password, role, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
			"Test User", "test@example.com", "hashed", "user", true).Error
		if err != nil {
			return err
		}

		// Force rollback
		return assert.AnError
	})

	assert.Error(t, err)

	// Verify user was not created
	var count int64
	db.Raw("SELECT COUNT(*) FROM users WHERE email = ?", "test@example.com").Scan(&count)
	assert.Equal(t, int64(0), count)

	// Test successful transaction
	err = db.Transaction(func(tx *database.Gorm) error {
		return tx.Exec("INSERT INTO users (name, email, password, role, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
			"Test User", "test2@example.com", "hashed", "user", true).Error
	})

	require.NoError(t, err)

	// Verify user was created
	db.Raw("SELECT COUNT(*) FROM users WHERE email = ?", "test2@example.com").Scan(&count)
	assert.Equal(t, int64(1), count)
}
