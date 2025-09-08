package database

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// DB represents a database connection
type DB struct {
	*sql.DB
	config     *Config
	driverName string
	mu         sync.RWMutex
}

// Config holds database configuration
type Config struct {
	Driver   string // sqlite, mysql, postgres
	Host     string
	Port     int
	Database string
	Username string
	Password string
	
	// Connection pool settings
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	
	// SQLite specific
	SQLitePath string
}

// DefaultConfig returns a default database configuration
func DefaultConfig() *Config {
	return &Config{
		Driver:          "sqlite",
		SQLitePath:      "obtura.db",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

// New creates a new database connection
func New(config *Config) (*DB, error) {
	if config == nil {
		config = DefaultConfig()
	}

	dsn := buildDSN(config)
	
	// Map sqlite to sqlite3 for the sql driver
	driverName := config.Driver
	if driverName == "sqlite" {
		driverName = "sqlite3"
	}
	
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		DB:         db,
		config:     config,
		driverName: config.Driver,
	}, nil
}

// buildDSN builds the data source name for the database connection
func buildDSN(config *Config) string {
	switch config.Driver {
	case "sqlite", "sqlite3":
		return config.SQLitePath
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Username, config.Password, config.Host, config.Port, config.Database)
	case "postgres", "postgresql":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.Username, config.Password, config.Database)
	default:
		return ""
	}
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Driver returns the database driver name
func (db *DB) Driver() string {
	return db.driverName
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Exec executes a query without returning any rows
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.DB.Exec(query, args...)
}

// Query executes a query that returns rows
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.Query(query, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRow(query, args...)
}