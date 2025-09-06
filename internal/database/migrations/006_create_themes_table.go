package migrations

import (
	"database/sql"

	"github.com/btassone/obtura/pkg/database"
)

func init() {
	RegisterMigration(&database.Migration{
		Version:     "006_create_themes_table",
		Description: "Create themes table",
		Up: func(tx *sql.Tx) error {
			query := `
				CREATE TABLE themes (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name VARCHAR(255) NOT NULL UNIQUE,
					version VARCHAR(50) NOT NULL,
					description TEXT,
					author VARCHAR(255),
					screenshot VARCHAR(500),
					active BOOLEAN DEFAULT false,
					settings TEXT, -- JSON data
					constraints TEXT, -- JSON data for theme constraints
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				)
			`
			// Adjust for different databases
			if DriverName == "mysql" {
				query = `
					CREATE TABLE themes (
						id INT AUTO_INCREMENT PRIMARY KEY,
						name VARCHAR(255) NOT NULL UNIQUE,
						version VARCHAR(50) NOT NULL,
						description TEXT,
						author VARCHAR(255),
						screenshot VARCHAR(500),
						active BOOLEAN DEFAULT false,
						settings JSON,
						constraints JSON,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
						INDEX idx_active (active)
					)
				`
			} else if DriverName == "postgres" || DriverName == "postgresql" {
				query = `
					CREATE TABLE themes (
						id SERIAL PRIMARY KEY,
						name VARCHAR(255) NOT NULL UNIQUE,
						version VARCHAR(50) NOT NULL,
						description TEXT,
						author VARCHAR(255),
						screenshot VARCHAR(500),
						active BOOLEAN DEFAULT false,
						settings JSONB,
						constraints JSONB,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
					);
					
					CREATE INDEX idx_themes_active ON themes(active);
				`
			}

			_, err := tx.Exec(query)
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec("DROP TABLE IF EXISTS themes")
			return err
		},
	})
}