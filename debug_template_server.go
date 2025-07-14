package main

import (
	"net/http"
	"strconv"

	"go-barcode-webapp/internal/config"
	"go-barcode-webapp/internal/handlers"
	"go-barcode-webapp/internal/repository"
	"go-barcode-webapp/internal/models"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		panic(err)
	}

	// Initialize database
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Initialize repositories
	jobRepo := repository.NewJobRepository(db)

	// Setup Gin router
	r := gin.Default()

	// Load templates
	r.LoadHTMLGlob("web/templates/*")

	r.GET("/test/job/:id", func(c *gin.Context) {
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

		// Group devices by product and calculate pricing (same logic as handler)
		productGroups := make(map[string]*handlers.ProductGroup)
		totalDevices := len(jobDevices)
		totalValue := 0.0

		for _, jd := range jobDevices {
			if jd.Device.Product == nil {
				continue
			}

			productName := jd.Device.Product.Name
			if _, exists := productGroups[productName]; !exists {
				productGroups[productName] = &handlers.ProductGroup{
					Product: jd.Device.Product,
					Devices: []models.JobDevice{},
				}
			}

			// Calculate effective price
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
		}

		// Create a mock user
		user := gin.H{
			"Username": "admin",
			"UserID":   1,
		}

		c.HTML(http.StatusOK, "job_detail.html", gin.H{
			"title":         "Job Details",
			"job":           job,
			"jobDevices":    jobDevices,
			"productGroups": productGroups,
			"totalDevices":  totalDevices,
			"totalValue":    totalValue,
			"user":          user,
		})
	})

	r.Run(":8082")
}