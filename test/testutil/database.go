package testutil

import (
	"database/sql"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDB creates an in-memory SQLite database for testing
func TestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	})

	return db
}

// TestDBWithSchema creates a test database and runs migrations
func TestDBWithSchema(t *testing.T, models ...interface{}) *gorm.DB {
	db := TestDB(t)

	err := db.AutoMigrate(models...)
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

// SeedTestData seeds the database with test data
func SeedTestData(t *testing.T, db *gorm.DB, records ...interface{}) {
	for _, record := range records {
		err := db.Create(record).Error
		if err != nil {
			t.Fatalf("failed to seed test data: %v", err)
		}
	}
}

// AssertRecordExists checks if a record exists in the database
func AssertRecordExists(t *testing.T, db *gorm.DB, model interface{}, conditions ...interface{}) {
	var count int64
	query := db.Model(model)
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}
	err := query.Count(&count).Error
	if err != nil {
		t.Fatalf("failed to count records: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected record to exist, but it doesn't")
	}
}

// AssertRecordNotExists checks if a record does not exist in the database
func AssertRecordNotExists(t *testing.T, db *gorm.DB, model interface{}, conditions ...interface{}) {
	var count int64
	query := db.Model(model)
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}
	err := query.Count(&count).Error
	if err != nil {
		t.Fatalf("failed to count records: %v", err)
	}
	if count > 0 {
		t.Fatalf("expected record not to exist, but found %d", count)
	}
}

// TruncateTable removes all records from a table
func TruncateTable(t *testing.T, db *gorm.DB, model interface{}) {
	err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(model).Error
	if err != nil {
		t.Fatalf("failed to truncate table: %v", err)
	}
}

// ExecSQL executes raw SQL for test setup
func ExecSQL(t *testing.T, db *gorm.DB, query string, args ...interface{}) {
	err := db.Exec(query, args...).Error
	if err != nil {
		t.Fatalf("failed to execute SQL: %v", err)
	}
}

// QueryRow executes a raw SQL query and returns a single row
func QueryRow(t *testing.T, db *gorm.DB, dest interface{}, query string, args ...interface{}) {
	err := db.Raw(query, args...).Scan(dest).Error
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("failed to query row: %v", err)
	}
}

// MustCreate creates a record and fails the test if there's an error
func MustCreate(t *testing.T, db *gorm.DB, record interface{}) {
	err := db.Create(record).Error
	if err != nil {
		t.Fatalf("failed to create record: %v", err)
	}
}

// MustUpdate updates a record and fails the test if there's an error
func MustUpdate(t *testing.T, db *gorm.DB, record interface{}) {
	err := db.Save(record).Error
	if err != nil {
		t.Fatalf("failed to update record: %v", err)
	}
}

// MustDelete deletes a record and fails the test if there's an error
func MustDelete(t *testing.T, db *gorm.DB, record interface{}) {
	err := db.Delete(record).Error
	if err != nil {
		t.Fatalf("failed to delete record: %v", err)
	}
}

// AssertCount asserts the number of records matching conditions
func AssertCount(t *testing.T, db *gorm.DB, model interface{}, expected int64, conditions ...interface{}) {
	var count int64
	query := db.Model(model)
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}
	err := query.Count(&count).Error
	if err != nil {
		t.Fatalf("failed to count records: %v", err)
	}
	if count != expected {
		t.Fatalf("expected %d records, got %d", expected, count)
	}
}

// WithTransaction runs a test within a database transaction that gets rolled back
func WithTransaction(t *testing.T, db *gorm.DB, fn func(*gorm.DB)) {
	tx := db.Begin()
	defer tx.Rollback()

	fn(tx)
}
