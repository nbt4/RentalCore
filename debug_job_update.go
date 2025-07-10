package main

import (
	"fmt"
	"log"
	"go-barcode-webapp/internal/config"
	"go-barcode-webapp/internal/repository"
)

func main() {
	// Load configuration
	cfg := &config.Config{}
	cfg.Database.Host = "mysql"
	cfg.Database.Port = 3306
	cfg.Database.Database = "TS-Lager"
	cfg.Database.Username = "tsweb"
	cfg.Database.Password = "TS123!2024#"
	cfg.Database.PoolSize = 5

	// Initialize database
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create job repository
	jobRepo := repository.NewJobRepository(db)

	// Test getting a job (use ID 1 as example)
	fmt.Println("=== Testing Job Update ===")
	
	job, err := jobRepo.GetByID(1)
	if err != nil {
		log.Fatalf("Failed to get job: %v", err)
	}

	fmt.Printf("Original job description: '%s'\n", func() string {
		if job.Description == nil {
			return "<nil>"
		}
		return *job.Description
	}())

	// Update description
	newDescription := "Test update from debug tool"
	job.Description = &newDescription

	fmt.Printf("Updating to description: '%s'\n", newDescription)

	// Save the job
	err = jobRepo.Update(job)
	if err != nil {
		log.Fatalf("Failed to update job: %v", err)
	}

	// Verify the update
	updatedJob, err := jobRepo.GetByID(1)
	if err != nil {
		log.Fatalf("Failed to get updated job: %v", err)
	}

	fmt.Printf("After update, description is: '%s'\n", func() string {
		if updatedJob.Description == nil {
			return "<nil>"
		}
		return *updatedJob.Description
	}())

	if updatedJob.Description != nil && *updatedJob.Description == newDescription {
		fmt.Println("✅ Update successful!")
	} else {
		fmt.Println("❌ Update failed!")
	}
}