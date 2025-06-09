package main

import (
	"fmt"
	"log"
	"go-barcode-webapp/internal/config"
	"go-barcode-webapp/internal/repository"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.production.direct.json")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create job repository
	jobRepo := repository.NewJobRepository(db)

	// Test specific job retrieval for job 1004
	fmt.Printf("🔍 Testing job 1004 retrieval:\n")
	
	job, err := jobRepo.GetByID(1004)
	if err != nil {
		fmt.Printf("   ❌ ERROR getting job 1004: %v\n", err)
		return
	}
	
	fmt.Printf("   ✅ Job found: %d\n", job.JobID)
	fmt.Printf("   Customer: %s (ID: %d)\n", job.Customer.GetDisplayName(), job.CustomerID)
	if job.Description != nil {
		fmt.Printf("   Description: %s\n", *job.Description)
	}
	fmt.Printf("   Status: %d\n", job.StatusID)

	// Test getting job devices
	devices, err := jobRepo.GetJobDevices(1004)
	if err != nil {
		fmt.Printf("   ❌ ERROR getting job devices: %v\n", err)
		return
	}
	
	fmt.Printf("   Assigned devices: %d\n", len(devices))
	for i, device := range devices {
		fmt.Printf("     %d. Device: %s\n", i+1, device.DeviceID)
	}
}