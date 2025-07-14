package main

import (
	"fmt"
	"log"

	"go-barcode-webapp/internal/config"
	"go-barcode-webapp/internal/repository"
	"go-barcode-webapp/internal/models"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	jobRepo := repository.NewJobRepository(db)

	// List all jobs
	fmt.Printf("=== Listing all jobs ===\n")
	
	params := &models.FilterParams{
		Limit: 10,
	}
	
	jobs, err := jobRepo.List(params)
	if err != nil {
		log.Fatalf("Failed to list jobs: %v", err)
	}
	
	fmt.Printf("Found %d jobs:\n", len(jobs))
	for _, job := range jobs {
		fmt.Printf("Job ID: %d, Customer: %s, Status: %s, Device Count: %d\n", 
			job.JobID, job.CustomerName, job.StatusName, job.DeviceCount)
	}
	
	// Get first job if exists
	if len(jobs) > 0 {
		firstJob := jobs[0]
		fmt.Printf("\n=== Testing Job ID: %d ===\n", firstJob.JobID)
		
		// Get job devices
		jobDevices, err := jobRepo.GetJobDevices(firstJob.JobID)
		if err != nil {
			log.Fatalf("Failed to get job devices: %v", err)
		}
		
		fmt.Printf("Job devices count: %d\n", len(jobDevices))
		
		for i, jd := range jobDevices {
			fmt.Printf("Device %d: ID=%s, Product=%s\n", i+1, jd.DeviceID, 
				func() string {
					if jd.Device.Product != nil {
						return jd.Device.Product.Name
					}
					return "nil"
				}())
		}
	}
}