package seeders

import (
	"github.com/btassone/obtura/pkg/database"
)

var (
	// seeders holds all registered seeders
	seeders []database.Seeder
)

// RegisterSeeder adds a seeder to the registry
func RegisterSeeder(seeder *database.Seeder) {
	seeders = append(seeders, *seeder)
}

// GetSeeders returns all registered seeders
func GetSeeders() []database.Seeder {
	return seeders
}