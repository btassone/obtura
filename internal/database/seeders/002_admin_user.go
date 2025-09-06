package seeders

import (
	"fmt"
	"time"

	"github.com/btassone/obtura/pkg/database"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	RegisterSeeder(&database.Seeder{
		Name:        "002_admin_user",
		Description: "Create default admin user",
		Run: func(db *database.DB) error {
			// Check if admin user already exists
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", "admin@example.com").Scan(&count)
			if err != nil {
				return err
			}

			if count > 0 {
				fmt.Println("Admin user already exists, skipping...")
				return nil
			}

			// Use a default password for development
			defaultPassword := "admin123"
			
			// Hash the password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
			if err != nil {
				return err
			}

			// Insert admin user
			_, err = db.Exec(
				`INSERT INTO users (name, email, password, role, active, email_verified_at, created_at, updated_at) 
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				"Admin User",
				"admin@example.com",
				string(hashedPassword),
				"admin",
				true,
				time.Now(),
				time.Now(),
				time.Now(),
			)
			if err != nil {
				return err
			}

			fmt.Printf("\n=================================\n")
			fmt.Printf("Admin user created successfully!\n")
			fmt.Printf("Email: admin@example.com\n")
			fmt.Printf("Password: %s\n", defaultPassword)
			fmt.Printf("Please change this password immediately!\n")
			fmt.Printf("=================================\n\n")

			return nil
		},
	})
}