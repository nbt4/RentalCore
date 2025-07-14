package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"go-barcode-webapp/internal/config"
	"go-barcode-webapp/internal/repository"
	"go-barcode-webapp/internal/handlers"
	"go-barcode-webapp/internal/models"

	"github.com/gin-gonic/gin"
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

	// Test specific job
	jobID := uint(1004) // Change this to test different job IDs
	
	fmt.Printf("=== Testing Job ID: %d ===\n", jobID)
	
	// Get job details
	job, err := jobRepo.GetByID(jobID)
	if err != nil {
		log.Fatalf("Failed to get job: %v", err)
	}
	
	fmt.Printf("Job found: %+v\n", job)
	
	// Get job devices
	jobDevices, err := jobRepo.GetJobDevices(jobID)
	if err != nil {
		log.Fatalf("Failed to get job devices: %v", err)
	}
	
	fmt.Printf("Raw job devices count: %d\n", len(jobDevices))
	
	// Group devices by product and calculate pricing (same logic as handler)
	productGroups := make(map[string]*handlers.ProductGroup)
	totalDevices := len(jobDevices)
	totalValue := 0.0

	fmt.Printf("Processing %d jobDevices\n", len(jobDevices))
	
	for i, jd := range jobDevices {
		fmt.Printf("Processing device %d: ID=%s\n", i+1, jd.DeviceID)
		
		if jd.Device.Product == nil {
			fmt.Printf("Device %s has no product, skipping\n", jd.DeviceID)
			continue
		}

		productName := jd.Device.Product.Name
		fmt.Printf("Device %s has product: %s\n", jd.DeviceID, productName)
		
		if _, exists := productGroups[productName]; !exists {
			productGroups[productName] = &handlers.ProductGroup{
				Product: jd.Device.Product,
				Devices: []models.JobDevice{},
			}
			fmt.Printf("Created new product group for: %s\n", productName)
		}

		// Calculate effective price (custom price if set, otherwise default product price)
		var effectivePrice float64
		if jd.CustomPrice != nil && *jd.CustomPrice > 0 {
			effectivePrice = *jd.CustomPrice
		} else if jd.Device.Product.ItemCostPerDay != nil {
			effectivePrice = *jd.Device.Product.ItemCostPerDay
		}

		// Create a copy of the job device with calculated price for display
		jdCopy := jd
		jdCopy.CustomPrice = &effectivePrice

		productGroups[productName].Devices = append(productGroups[productName].Devices, jdCopy)
		productGroups[productName].Count = len(productGroups[productName].Devices)
		productGroups[productName].TotalValue += effectivePrice
		totalValue += effectivePrice
		
		fmt.Printf("Added device to group %s, now has %d devices\n", productName, productGroups[productName].Count)
	}
	
	fmt.Printf("Final product groups: %d\n", len(productGroups))
	for name, group := range productGroups {
		fmt.Printf("Group %s: %d devices\n", name, group.Count)
		for i, device := range group.Devices {
			fmt.Printf("  Device %d: %s (Serial: %s)\n", i+1, device.DeviceID, device.Device.SerialNumber)
		}
	}
	
	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total devices: %d\n", totalDevices)
	fmt.Printf("Total value: %.2f\n", totalValue)
	fmt.Printf("Product groups: %d\n", len(productGroups))
	
	// Create a simple web server to test the endpoint
	r := gin.Default()
	
	r.GET("/debug/job/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
			return
		}
		
		job, err := jobRepo.GetByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}
		
		jobDevices, err := jobRepo.GetJobDevices(uint(id))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"job": job,
			"devices": jobDevices,
			"device_count": len(jobDevices),
		})
	})
	
	fmt.Printf("\nStarting debug server on :8081\n")
	fmt.Printf("Visit: http://localhost:8081/debug/job/%d\n", jobID)
	
	r.Run(":8081")
}