package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"go-barcode-webapp/internal/config"
	"go-barcode-webapp/internal/handlers"
	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/repository"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.json", "Configuration file path")
	username := flag.String("username", "", "Username (required)")
	email := flag.String("email", "", "Email address (required)")
	password := flag.String("password", "", "Password (required)")
	firstName := flag.String("firstname", "", "First name")
	lastName := flag.String("lastname", "", "Last name")
	flag.Parse()

	// Validate required fields
	if *username == "" || *email == "" || *password == "" {
		fmt.Println("Usage: go run create_user.go -username=<username> -email=<email> -password=<password> [-firstname=<name>] [-lastname=<name>] [-config=<config.json>]")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  go run create_user.go -username=admin -email=admin@company.com -password=secure123 -firstname=Admin -lastname=User")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Printf("Failed to load config, using defaults: %v", err)
		cfg = &config.Config{}
		cfg.Database.Host = "localhost"
		cfg.Database.Port = 3306
		cfg.Database.Database = "jobscanner"
		cfg.Database.Username = "root"
		cfg.Database.Password = ""
	}

	// Initialize database
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Auto-migrate User table
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to auto-migrate users table: %v", err)
	}

	// Create auth handler for user creation
	authHandler := handlers.NewAuthHandler(db.DB)

	// Create user
	err = authHandler.CreateUser(*username, *email, *password, *firstName, *lastName)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("✅ User created successfully!\n")
	fmt.Printf("Username: %s\n", *username)
	fmt.Printf("Email: %s\n", *email)
	fmt.Printf("First Name: %s\n", *firstName)
	fmt.Printf("Last Name: %s\n", *lastName)
	fmt.Println("\nYou can now login to the web interface with these credentials.")
}