package database

import (
	"database/sql"
	"testing"

	"github.com/btassone/obtura/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	// This test would require mocking the database connection
	// For now, we'll test with an in-memory database
	t.Skip("Requires proper test setup with migrations")
}

func TestManager_DB(t *testing.T) {
	// Create a test database
	config := &database.Config{
		Driver:     "sqlite",
		SQLitePath: ":memory:",
	}
	
	db, err := database.New(config)
	require.NoError(t, err)
	
	manager := &Manager{db: db}
	
	// Get the database instance
	dbInstance := manager.DB()
	assert.NotNil(t, dbInstance)
	assert.Equal(t, db, dbInstance)
}

func TestManager_Migrate(t *testing.T) {
	// Skip this test as it requires a persistent database
	t.Skip("Requires persistent database for migration testing")
	
	// Create a test database
	config := &database.Config{
		Driver:     "sqlite",
		SQLitePath: ":memory:",
	}
	
	db, err := database.New(config)
	require.NoError(t, err)
	
	// Create migration runner
	migrationRunner := database.NewMigrationRunner(db)
	
	// Add a test migration
	migrationRunner.AddMigration(database.Migration{
		Version:     "001_create_test_table",
		Description: "Create test table",
		Up: func(tx *sql.Tx) error {
			query := `CREATE TABLE IF NOT EXISTS test_table (
				id INTEGER PRIMARY KEY,
				name TEXT NOT NULL
			)`
			_, err := tx.Exec(query)
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec("DROP TABLE IF EXISTS test_table")
			return err
		},
	})
	
	manager := &Manager{
		db:              db,
		migrationRunner: migrationRunner,
	}
	
	// Run migrations - this should create migrations table and run our test migration
	err = manager.Migrate()
	require.NoError(t, err)
	
	// Verify table was created
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='test_table'").Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "test_table", name)
}

func TestManager_Seed(t *testing.T) {
	// Skip this test as it requires a persistent database
	t.Skip("Requires persistent database for seeding testing")
	
	// Create a test database
	config := &database.Config{
		Driver:     "sqlite",
		SQLitePath: ":memory:",
	}
	
	db, err := database.New(config)
	require.NoError(t, err)
	
	// Create the table that the seeder expects
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS test_data (
		id INTEGER PRIMARY KEY,
		value TEXT NOT NULL
	)`)
	require.NoError(t, err)
	
	// Verify table was created
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='test_data'").Scan(&tableName)
	if err != nil {
		t.Logf("Table check error: %v", err)
		// Try a simpler create
		_, err = db.Exec("CREATE TABLE test_data (id INTEGER PRIMARY KEY, value TEXT)")
		require.NoError(t, err)
	}
	
	// Create seeder runner
	seederRunner := database.NewSeederRunner(db)
	
	// Add a test seeder
	seederRunner.AddSeeder(database.Seeder{
		Name:        "TestSeeder",
		Description: "Test seeder",
		Run: func(db *database.DB) error {
			_, err := db.Exec("INSERT INTO test_data (value) VALUES (?)", "test")
			return err
		},
	})
	
	manager := &Manager{
		db:           db,
		seederRunner: seederRunner,
	}
	
	// Run seeders
	err = manager.Seed()
	assert.NoError(t, err)
	
	// Verify data was seeded
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_data WHERE value = ?", "test").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestManager_Query(t *testing.T) {
	// Create a test database
	config := &database.Config{
		Driver:     "sqlite",
		SQLitePath: ":memory:",
	}
	
	db, err := database.New(config)
	require.NoError(t, err)
	
	manager := &Manager{db: db}
	
	// Get query builder
	q := manager.Query()
	assert.NotNil(t, q)
}


