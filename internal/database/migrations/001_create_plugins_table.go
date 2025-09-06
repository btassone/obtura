package migrations

import (
	"database/sql"

	"github.com/btassone/obtura/pkg/database"
)

func init() {
	RegisterMigration(&database.Migration{
		Version:     "001_create_plugins_table",
		Description: "Create plugins table",
		Up: func(tx *sql.Tx) error {
			query := `
				CREATE TABLE plugins (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name VARCHAR(255) NOT NULL UNIQUE,
					version VARCHAR(50) NOT NULL,
					description TEXT,
					author VARCHAR(255),
					active BOOLEAN DEFAULT true,
					settings TEXT, -- JSON data
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)
			`
			// Adjust for different databases
			if DriverName == "mysql" {
				query = `
					CREATE TABLE plugins (
						id INT AUTO_INCREMENT PRIMARY KEY,
						name VARCHAR(255) NOT NULL UNIQUE,
						version VARCHAR(50) NOT NULL,
						description TEXT,
						author VARCHAR(255),
						active BOOLEAN DEFAULT true,
						settings JSON,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
					)
				`
			} else if DriverName == "postgres" || DriverName == "postgresql" {
				query = `
					CREATE TABLE plugins (
						id SERIAL PRIMARY KEY,
						name VARCHAR(255) NOT NULL UNIQUE,
						version VARCHAR(50) NOT NULL,
						description TEXT,
						author VARCHAR(255),
						active BOOLEAN DEFAULT true,
						settings JSONB,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
					)
				`
			}

			_, err := tx.Exec(query)
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec("DROP TABLE IF EXISTS plugins")
			return err
		},
	})
}