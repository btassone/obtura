package migrations

import (
	"database/sql"

	"github.com/btassone/obtura/pkg/database"
)

func init() {
	RegisterMigration(&database.Migration{
		Version:     "003_create_users_table",
		Description: "Create users table",
		Up: func(tx *sql.Tx) error {
			query := `
				CREATE TABLE users (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name VARCHAR(255) NOT NULL,
					email VARCHAR(255) NOT NULL UNIQUE,
					password VARCHAR(255) NOT NULL,
					role VARCHAR(50) DEFAULT 'user',
					avatar VARCHAR(255),
					bio TEXT,
					active BOOLEAN DEFAULT true,
					email_verified_at TIMESTAMP,
					remember_token VARCHAR(100),
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					deleted_at TIMESTAMP
				)
			`
			// Adjust for different databases
			if DriverName == "mysql" {
				query = `
					CREATE TABLE users (
						id INT AUTO_INCREMENT PRIMARY KEY,
						name VARCHAR(255) NOT NULL,
						email VARCHAR(255) NOT NULL UNIQUE,
						password VARCHAR(255) NOT NULL,
						role VARCHAR(50) DEFAULT 'user',
						avatar VARCHAR(255),
						bio TEXT,
						active BOOLEAN DEFAULT true,
						email_verified_at TIMESTAMP NULL,
						remember_token VARCHAR(100),
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
						deleted_at TIMESTAMP NULL,
						INDEX idx_email (email),
						INDEX idx_role (role)
					)
				`
			} else if DriverName == "postgres" || DriverName == "postgresql" {
				query = `
					CREATE TABLE users (
						id SERIAL PRIMARY KEY,
						name VARCHAR(255) NOT NULL,
						email VARCHAR(255) NOT NULL UNIQUE,
						password VARCHAR(255) NOT NULL,
						role VARCHAR(50) DEFAULT 'user',
						avatar VARCHAR(255),
						bio TEXT,
						active BOOLEAN DEFAULT true,
						email_verified_at TIMESTAMP,
						remember_token VARCHAR(100),
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						deleted_at TIMESTAMP
					);
					
					CREATE INDEX idx_users_email ON users(email);
					CREATE INDEX idx_users_role ON users(role);
				`
			}

			_, err := tx.Exec(query)
			return err
		},
		Down: func(tx *sql.Tx) error {
			_, err := tx.Exec("DROP TABLE IF EXISTS users")
			return err
		},
	})
}