package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/btassone/obtura/internal/database"
)

func runMigrate() {
	dbManager, err := database.NewManager()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbManager.Close()

	// Get migration status
	status, err := dbManager.MigrationStatus()
	if err != nil {
		log.Fatalf("Failed to get migration status: %v", err)
	}

	// Show pending migrations
	var pending []string
	for _, s := range status {
		if !s.Applied {
			pending = append(pending, s.Version)
		}
	}

	if len(pending) == 0 {
		fmt.Println("No migrations to run.")
		return
	}

	fmt.Printf("Found %d pending migration(s):\n", len(pending))
	for _, v := range pending {
		fmt.Printf("  - %s\n", v)
	}

	// Run migrations
	if err := dbManager.Migrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}

func runRollback() {
	// Parse rollback steps
	rollbackCmd := flag.NewFlagSet("rollback", flag.ExitOnError)
	steps := rollbackCmd.Int("steps", 1, "Number of migrations to rollback")
	rollbackCmd.Parse(os.Args[2:])

	dbManager, err := database.NewManager()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbManager.Close()

	// Run rollback
	if err := dbManager.Rollback(*steps); err != nil {
		log.Fatalf("Rollback failed: %v", err)
	}
}

func runSeed() {
	// Parse seeder names
	seedCmd := flag.NewFlagSet("seed", flag.ExitOnError)
	specific := seedCmd.String("only", "", "Comma-separated list of specific seeders to run")
	seedCmd.Parse(os.Args[2:])

	dbManager, err := database.NewManager()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbManager.Close()

	// Run seeders
	if *specific != "" {
		seeders := strings.Split(*specific, ",")
		if err := dbManager.Seed(seeders...); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
	} else {
		// Show available seeders
		seeders := dbManager.ListSeeders()
		fmt.Printf("Available seeders:\n")
		for _, s := range seeders {
			fmt.Printf("  - %s: %s\n", s.Name, s.Description)
		}
		fmt.Println()

		// Run all seeders
		if err := dbManager.Seed(); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
	}
}