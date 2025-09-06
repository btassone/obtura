package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/btassone/obtura/internal/server"

	// Import migrations and seeders to register them
	_ "github.com/btassone/obtura/internal/database/migrations"
	_ "github.com/btassone/obtura/internal/database/seeders"
)

func main() {
	// Handle database commands first (before flag parsing)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			runMigrate()
			return
		case "seed":
			runSeed()
			return
		case "rollback":
			runRollback()
			return
		}
	}

	var (
		port = flag.String("port", "8080", "Server port")
		mode = flag.String("mode", "dev", "Run mode (dev/prod)")
	)
	flag.Parse()

	if len(os.Args) > 1 && os.Args[1] == "serve" {
		srv, err := server.New(*port, *mode)
		if err != nil {
			log.Fatalf("Failed to initialize server: %v", err)
		}
		
		// Handle graceful shutdown
		defer func() {
			if err := srv.Close(); err != nil {
				log.Printf("Error closing server: %v", err)
			}
		}()
		
		if *mode == "dev" {
			log.Printf("Starting Obtura development server on port %s", *port)
			log.Printf("Access the application at http://localhost:3000 (proxied by Air)")
			log.Printf("Using SQLite database for development")
		} else {
			log.Printf("Starting Obtura server on port %s in %s mode", *port, *mode)
		}
		
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
		return
	}

	fmt.Println("Obtura - A modular web framework")
	fmt.Println("\nUsage:")
	fmt.Println("  obtura serve      Start the web server")
	fmt.Println("  obtura migrate    Run database migrations")
	fmt.Println("  obtura rollback   Rollback database migrations")
	fmt.Println("  obtura seed       Run database seeders")
	fmt.Println("  obtura generate   Generate components (coming soon)")
	fmt.Println("  obtura build      Build for production (coming soon)")
}
