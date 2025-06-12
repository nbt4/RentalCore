package main

import (
	"fmt"
	"log"
	"encoding/json"
	"os"

	"go-barcode-webapp/internal/repository"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Database string `json:"database"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"database"`
}

func main() {
	// Load config
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal("Error opening config file:", err)
	}
	defer configFile.Close()

	var config Config
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		log.Fatal("Error decoding config:", err)
	}

	// Connect to database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Database.Username,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Database)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create repositories
	dbRepo := &repository.Database{DB: db}
	deviceRepo := repository.NewDeviceRepository(dbRepo)

	fmt.Println("=== TESTING TEMPLATE DATA ===")
	
	// Test GetAvailableDevicesForCaseManagement (the method used in the template)
	fmt.Println("\n1. Testing GetAvailableDevicesForCaseManagement():")
	devices, err := deviceRepo.GetAvailableDevicesForCaseManagement()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	fmt.Printf("Found %d devices\n", len(devices))
	
	if len(devices) == 0 {
		fmt.Println("❌ NO DEVICES FOUND - This is the problem!")
	} else {
		fmt.Println("✅ Devices found successfully!")
		
		// Show first 5 devices with details
		for i, device := range devices {
			if i >= 5 { break }
			
			productName := "NULL"
			categoryName := "NULL"
			subcategoryName := "NULL"
			
			if device.Product != nil {
				productName = device.Product.Name
				if device.Product.Category != nil {
					categoryName = device.Product.Category.Name
				}
				if device.Product.Subcategory != nil {
					subcategoryName = device.Product.Subcategory.Name
				}
			}
			
			serialNum := "NULL"
			if device.SerialNumber != nil {
				serialNum = *device.SerialNumber
			}
			
			fmt.Printf("  Device %d:\n", i+1)
			fmt.Printf("    DeviceID: %s\n", device.DeviceID)
			fmt.Printf("    Status: %s\n", device.Status)
			fmt.Printf("    Product: %s\n", productName)
			fmt.Printf("    Category: %s\n", categoryName)
			fmt.Printf("    Subcategory: %s\n", subcategoryName)
			fmt.Printf("    SerialNumber: %s\n", serialNum)
			fmt.Printf("    ProductID: %v\n", device.ProductID)
			fmt.Println()
		}
	}

	// Let's also check the regular GetAvailableDevices method
	fmt.Println("\n2. Testing GetAvailableDevices() (different method):")
	devices2, err := deviceRepo.GetAvailableDevices()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	
	fmt.Printf("Found %d devices with GetAvailableDevices()\n", len(devices2))
	
	if len(devices2) == 0 {
		fmt.Println("❌ NO DEVICES FOUND with GetAvailableDevices either!")
	} else {
		fmt.Println("✅ GetAvailableDevices() works fine")
		for i, device := range devices2 {
			if i >= 3 { break }
			fmt.Printf("  Device %d: ID=%s, Status=%s\n", i+1, device.DeviceID, device.Status)
		}
	}

	fmt.Println("\n=== TEMPLATE DATA TEST COMPLETE ===")
}