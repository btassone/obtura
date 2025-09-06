package seeders

import (
	"time"

	"github.com/btassone/obtura/pkg/database"
)

func init() {
	RegisterSeeder(&database.Seeder{
		Name:        "001_default_settings",
		Description: "Seed default application settings",
		Run: func(db *database.DB) error {
			settings := []struct {
				Key         string
				Value       string
				Type        string
				GroupName   string
				Description string
			}{
				{
					Key:         "site_title",
					Value:       "Obtura Framework",
					Type:        "string",
					GroupName:   "general",
					Description: "The title of your website",
				},
				{
					Key:         "site_description",
					Value:       "A modular web framework inspired by Laravel",
					Type:        "text",
					GroupName:   "general",
					Description: "A brief description of your website",
				},
				{
					Key:         "site_url",
					Value:       "http://localhost:3000",
					Type:        "string",
					GroupName:   "general",
					Description: "The URL of your website",
				},
				{
					Key:         "admin_email",
					Value:       "admin@example.com",
					Type:        "email",
					GroupName:   "general",
					Description: "Administrator email address",
				},
				{
					Key:         "timezone",
					Value:       "UTC",
					Type:        "string",
					GroupName:   "general",
					Description: "Default timezone for the application",
				},
				{
					Key:         "date_format",
					Value:       "2006-01-02",
					Type:        "string",
					GroupName:   "general",
					Description: "Default date format",
				},
				{
					Key:         "time_format",
					Value:       "15:04:05",
					Type:        "string",
					GroupName:   "general",
					Description: "Default time format",
				},
				{
					Key:         "posts_per_page",
					Value:       "10",
					Type:        "number",
					GroupName:   "reading",
					Description: "Number of posts to show per page",
				},
				{
					Key:         "enable_comments",
					Value:       "true",
					Type:        "boolean",
					GroupName:   "discussion",
					Description: "Allow comments on posts",
				},
				{
					Key:         "require_comment_moderation",
					Value:       "true",
					Type:        "boolean",
					GroupName:   "discussion",
					Description: "Comments must be approved before appearing",
				},
				{
					Key:         "upload_max_size",
					Value:       "10485760",
					Type:        "number",
					GroupName:   "media",
					Description: "Maximum upload file size in bytes (10MB default)",
				},
				{
					Key:         "allowed_file_types",
					Value:       "jpg,jpeg,png,gif,pdf,doc,docx,mp4,webm",
					Type:        "string",
					GroupName:   "media",
					Description: "Comma-separated list of allowed file extensions",
				},
				{
					Key:         "active_theme",
					Value:       "default",
					Type:        "string",
					GroupName:   "appearance",
					Description: "Currently active theme",
				},
				{
					Key:         "maintenance_mode",
					Value:       "false",
					Type:        "boolean",
					GroupName:   "general",
					Description: "Put site in maintenance mode",
				},
			}

			for _, setting := range settings {
				// Check if setting already exists
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM settings WHERE key = ?", setting.Key).Scan(&count)
				if err != nil {
					return err
				}

				if count == 0 {
					// Insert new setting
					_, err = db.Exec(
						`INSERT INTO settings (key, value, type, group_name, description, created_at, updated_at) 
						 VALUES (?, ?, ?, ?, ?, ?, ?)`,
						setting.Key,
						setting.Value,
						setting.Type,
						setting.GroupName,
						setting.Description,
						time.Now(),
						time.Now(),
					)
					if err != nil {
						return err
					}
				}
			}

			return nil
		},
	})
}