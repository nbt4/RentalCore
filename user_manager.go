package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"go-barcode-webapp/internal/config"
	"go-barcode-webapp/internal/handlers"
	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/repository"

	"golang.org/x/term"
)

func main() {
	var (
		configFile = flag.String("config", "config.json", "Configuration file path")
		username   = flag.String("username", "", "Username for the new user")
		email      = flag.String("email", "", "Email for the new user")
		firstName  = flag.String("firstname", "", "First name for the new user")
		lastName   = flag.String("lastname", "", "Last name for the new user")
		listUsers  = flag.Bool("list", false, "List all users")
		deleteUser = flag.String("delete", "", "Delete user by username")
	)
	flag.Parse()

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

	// Auto-migrate tables
	if err := db.GetDB().AutoMigrate(&models.User{}, &models.Session{}); err != nil {
		log.Fatalf("Failed to auto-migrate tables: %v", err)
	}

	authHandler := handlers.NewAuthHandler(db.GetDB())

	// Handle different operations
	switch {
	case *listUsers:
		listAllUsers(db)
	case *deleteUser != "":
		deleteUserByUsername(db, *deleteUser)
	case *username != "" || *email != "":
		createUser(authHandler, *username, *email, *firstName, *lastName)
	default:
		interactiveCreateUser(authHandler)
	}
}

func listAllUsers(db *repository.Database) {
	var users []models.User
	if err := db.GetDB().Find(&users).Error; err != nil {
		log.Fatalf("Failed to fetch users: %v", err)
	}

	if len(users) == 0 {
		fmt.Println("No users found.")
		return
	}

	fmt.Printf("%-10s %-20s %-30s %-20s %-20s %-10s %-20s\n", 
		"ID", "Username", "Email", "First Name", "Last Name", "Active", "Last Login")
	fmt.Println(strings.Repeat("-", 130))

	for _, user := range users {
		lastLogin := "Never"
		if user.LastLogin != nil {
			lastLogin = user.LastLogin.Format("2006-01-02 15:04")
		}
		
		fmt.Printf("%-10d %-20s %-30s %-20s %-20s %-10t %-20s\n",
			user.UserID, user.Username, user.Email, user.FirstName, user.LastName, user.IsActive, lastLogin)
	}
}

func deleteUserByUsername(db *repository.Database, username string) {
	var user models.User
	if err := db.GetDB().Where("username = ?", username).First(&user).Error; err != nil {
		log.Fatalf("User not found: %v", err)
	}

	fmt.Printf("Are you sure you want to delete user '%s' (%s)? [y/N]: ", user.Username, user.Email)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if response == "y" || response == "yes" {
		// Delete user's sessions first
		db.GetDB().Where("user_id = ?", user.UserID).Delete(&models.Session{})
		
		// Delete user
		if err := db.GetDB().Delete(&user).Error; err != nil {
			log.Fatalf("Failed to delete user: %v", err)
		}
		fmt.Printf("User '%s' deleted successfully.\n", username)
	} else {
		fmt.Println("User deletion cancelled.")
	}
}

func createUser(authHandler *handlers.AuthHandler, username, email, firstName, lastName string) {
	// Get missing information interactively
	scanner := bufio.NewScanner(os.Stdin)

	if username == "" {
		fmt.Print("Username: ")
		scanner.Scan()
		username = strings.TrimSpace(scanner.Text())
	}

	if email == "" {
		fmt.Print("Email: ")
		scanner.Scan()
		email = strings.TrimSpace(scanner.Text())
	}

	if firstName == "" {
		fmt.Print("First Name: ")
		scanner.Scan()
		firstName = strings.TrimSpace(scanner.Text())
	}

	if lastName == "" {
		fmt.Print("Last Name: ")
		scanner.Scan()
		lastName = strings.TrimSpace(scanner.Text())
	}

	// Get password securely
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}
	password := string(passwordBytes)
	fmt.Println() // New line after password input

	if password == "" {
		log.Fatal("Password cannot be empty")
	}

	// Validate required fields
	if username == "" || email == "" {
		log.Fatal("Username and email are required")
	}

	// Create user
	err = authHandler.CreateUser(username, email, password, firstName, lastName)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("User '%s' created successfully!\n", username)
}

func interactiveCreateUser(authHandler *handlers.AuthHandler) {
	fmt.Println("=== TS Jobscanner User Manager ===")
	fmt.Println("Creating a new user...")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Username: ")
	scanner.Scan()
	username := strings.TrimSpace(scanner.Text())

	fmt.Print("Email: ")
	scanner.Scan()
	email := strings.TrimSpace(scanner.Text())

	fmt.Print("First Name: ")
	scanner.Scan()
	firstName := strings.TrimSpace(scanner.Text())

	fmt.Print("Last Name: ")
	scanner.Scan()
	lastName := strings.TrimSpace(scanner.Text())

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}
	password := string(passwordBytes)
	fmt.Println() // New line after password input

	// Validate required fields
	if username == "" || email == "" || password == "" {
		log.Fatal("Username, email, and password are required")
	}

	// Create user
	err = authHandler.CreateUser(username, email, password, firstName, lastName)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("User '%s' created successfully!\n", username)
}