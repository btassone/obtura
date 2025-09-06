package database

import (
	"fmt"
	"sort"
)

// Seeder represents a database seeder
type Seeder struct {
	Name        string
	Description string
	Run         func(*DB) error
}

// SeederRunner handles running seeders
type SeederRunner struct {
	db      *DB
	seeders []Seeder
}

// NewSeederRunner creates a new seeder runner
func NewSeederRunner(db *DB) *SeederRunner {
	return &SeederRunner{
		db:      db,
		seeders: make([]Seeder, 0),
	}
}

// AddSeeder adds a seeder to the runner
func (r *SeederRunner) AddSeeder(seeder Seeder) {
	r.seeders = append(r.seeders, seeder)
}

// Run executes all seeders
func (r *SeederRunner) Run() error {
	// Sort seeders by name for consistent ordering
	sort.Slice(r.seeders, func(i, j int) bool {
		return r.seeders[i].Name < r.seeders[j].Name
	})

	for _, seeder := range r.seeders {
		fmt.Printf("Running seeder: %s\n", seeder.Description)
		
		if err := seeder.Run(r.db); err != nil {
			return fmt.Errorf("failed to run seeder %s: %w", seeder.Name, err)
		}
		
		fmt.Printf("Seeder %s completed successfully\n", seeder.Name)
	}

	return nil
}

// RunSpecific runs specific seeders by name
func (r *SeederRunner) RunSpecific(names ...string) error {
	seederMap := make(map[string]Seeder)
	for _, s := range r.seeders {
		seederMap[s.Name] = s
	}

	for _, name := range names {
		seeder, ok := seederMap[name]
		if !ok {
			return fmt.Errorf("seeder %s not found", name)
		}

		fmt.Printf("Running seeder: %s\n", seeder.Description)
		
		if err := seeder.Run(r.db); err != nil {
			return fmt.Errorf("failed to run seeder %s: %w", name, err)
		}
		
		fmt.Printf("Seeder %s completed successfully\n", name)
	}

	return nil
}

// List returns a list of all available seeders
func (r *SeederRunner) List() []SeederInfo {
	var info []SeederInfo
	for _, s := range r.seeders {
		info = append(info, SeederInfo{
			Name:        s.Name,
			Description: s.Description,
		})
	}
	return info
}

// SeederInfo provides information about a seeder
type SeederInfo struct {
	Name        string
	Description string
}