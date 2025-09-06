package migrations

import (
	"database/sql"

	"github.com/btassone/obtura/pkg/database"
)

func init() {
	RegisterMigration(&database.Migration{
		Version:     "005_create_media_table",
		Description: "Create media table",
		Up: func(tx *sql.Tx) error {
			query := `
				CREATE TABLE media (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					filename VARCHAR(255) NOT NULL,
					original_name VARCHAR(255) NOT NULL,
					mime_type VARCHAR(100) NOT NULL,
					size INTEGER NOT NULL,
					path VARCHAR(500) NOT NULL,
					url VARCHAR(500) NOT NULL,
					alt_text VARCHAR(255),
					title VARCHAR(255),
					description TEXT,
					dimensions VARCHAR(50), -- e.g., "1920x1080"
					user_id INTEGER,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
				)
			`
			// Adjust for different databases
			if DriverName == "mysql" {
				query = `
					CREATE TABLE media (
						id INT AUTO_INCREMENT PRIMARY KEY,
						filename VARCHAR(255) NOT NULL,
						original_name VARCHAR(255) NOT NULL,
						mime_type VARCHAR(100) NOT NULL,
						size BIGINT NOT NULL,
						path VARCHAR(500) NOT NULL,
						url VARCHAR(500) NOT NULL,
						alt_text VARCHAR(255),
						title VARCHAR(255),
						description TEXT,
						dimensions VARCHAR(50),
						user_id INT,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
						INDEX idx_mime_type (mime_type),
						INDEX idx_user_id (user_id)
					)
				`
			} else if DriverName == "postgres" || DriverName == "postgresql" {
				query = `
					CREATE TABLE media (
						id SERIAL PRIMARY KEY,
						filename VARCHAR(255) NOT NULL,
						original_name VARCHAR(255) NOT NULL,
						mime_type VARCHAR(100) NOT NULL,
						size BIGINT NOT NULL,
						path VARCHAR(500) NOT NULL,
						url VARCHAR(500) NOT NULL,
						alt_text VARCHAR(255),
						title VARCHAR(255),
						description TEXT,
						dimensions VARCHAR(50),
						user_id INTEGER,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
					);
					
					CREATE INDEX idx_media_mime_type ON media(mime_type);
					CREATE INDEX idx_media_user_id ON media(user_id);
				`
			}

			_, err := tx.Exec(query)
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec("DROP TABLE IF EXISTS media")
			return err
		},
	})
}