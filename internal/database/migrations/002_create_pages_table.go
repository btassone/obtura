package migrations

import (
	"database/sql"

	"github.com/btassone/obtura/pkg/database"
)

func init() {
	RegisterMigration(&database.Migration{
		Version:     "002_create_pages_table",
		Description: "Create pages table",
		Up: func(tx *sql.Tx) error {
			query := `
				CREATE TABLE pages (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					title VARCHAR(255) NOT NULL,
					slug VARCHAR(255) NOT NULL UNIQUE,
					content TEXT,
					excerpt TEXT,
					status VARCHAR(50) DEFAULT 'draft',
					layout VARCHAR(100) DEFAULT 'default',
					parent_id INTEGER,
					menu_order INTEGER DEFAULT 0,
					meta_title VARCHAR(255),
					meta_description TEXT,
					meta_keywords TEXT,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					published_at TIMESTAMP,
					FOREIGN KEY (parent_id) REFERENCES pages(id) ON DELETE SET NULL
				)
			`
			// Adjust for different databases
			if DriverName == "mysql" {
				query = `
					CREATE TABLE pages (
						id INT AUTO_INCREMENT PRIMARY KEY,
						title VARCHAR(255) NOT NULL,
						slug VARCHAR(255) NOT NULL UNIQUE,
						content TEXT,
						excerpt TEXT,
						status VARCHAR(50) DEFAULT 'draft',
						layout VARCHAR(100) DEFAULT 'default',
						parent_id INT,
						menu_order INT DEFAULT 0,
						meta_title VARCHAR(255),
						meta_description TEXT,
						meta_keywords TEXT,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
						published_at TIMESTAMP NULL,
						FOREIGN KEY (parent_id) REFERENCES pages(id) ON DELETE SET NULL,
						INDEX idx_slug (slug),
						INDEX idx_status (status)
					)
				`
			} else if DriverName == "postgres" || DriverName == "postgresql" {
				query = `
					CREATE TABLE pages (
						id SERIAL PRIMARY KEY,
						title VARCHAR(255) NOT NULL,
						slug VARCHAR(255) NOT NULL UNIQUE,
						content TEXT,
						excerpt TEXT,
						status VARCHAR(50) DEFAULT 'draft',
						layout VARCHAR(100) DEFAULT 'default',
						parent_id INTEGER,
						menu_order INTEGER DEFAULT 0,
						meta_title VARCHAR(255),
						meta_description TEXT,
						meta_keywords TEXT,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						published_at TIMESTAMP,
						FOREIGN KEY (parent_id) REFERENCES pages(id) ON DELETE SET NULL
					);
					
					CREATE INDEX idx_pages_slug ON pages(slug);
					CREATE INDEX idx_pages_status ON pages(status);
				`
			}

			_, err := tx.Exec(query)
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec("DROP TABLE IF EXISTS pages")
			return err
		},
	})
}