package migrations

import (
	"database/sql"

	"github.com/btassone/obtura/pkg/database"
)

func init() {
	RegisterMigration(&database.Migration{
		Version:     "004_create_settings_table",
		Description: "Create settings table",
		Up: func(tx *sql.Tx) error {
			query := `
				CREATE TABLE settings (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					key VARCHAR(255) NOT NULL UNIQUE,
					value TEXT,
					type VARCHAR(50) DEFAULT 'string',
					group_name VARCHAR(100) DEFAULT 'general',
					description TEXT,
					options TEXT, -- JSON array for select/multiselect types
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)
			`
			// Adjust for different databases
			if DriverName == "mysql" {
				query = `
					CREATE TABLE settings (
						id INT AUTO_INCREMENT PRIMARY KEY,
						` + "`key`" + ` VARCHAR(255) NOT NULL UNIQUE,
						value TEXT,
						type VARCHAR(50) DEFAULT 'string',
						group_name VARCHAR(100) DEFAULT 'general',
						description TEXT,
						options JSON,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
						INDEX idx_key (` + "`key`" + `),
						INDEX idx_group (group_name)
					)
				`
			} else if DriverName == "postgres" || DriverName == "postgresql" {
				query = `
					CREATE TABLE settings (
						id SERIAL PRIMARY KEY,
						key VARCHAR(255) NOT NULL UNIQUE,
						value TEXT,
						type VARCHAR(50) DEFAULT 'string',
						group_name VARCHAR(100) DEFAULT 'general',
						description TEXT,
						options JSONB,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
					);
					
					CREATE INDEX idx_settings_key ON settings(key);
					CREATE INDEX idx_settings_group ON settings(group_name);
				`
			}

			_, err := tx.Exec(query)
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec("DROP TABLE IF EXISTS settings")
			return err
		},
	})
}