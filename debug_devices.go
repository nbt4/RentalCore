package main

import (
	"fmt"
	"log"
	"encoding/json"
	"os"

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

type Device struct {
	DeviceID     string  `json:"deviceID" gorm:"primaryKey;column:deviceID"`
	ProductID    *uint   `json:"productID" gorm:"column:productID"`
	SerialNumber *string `json:"serialnumber" gorm:"column:serialnumber"`
	Status       string  `json:"status" gorm:"column:status;default:free"`
}

func (Device) TableName() string {
	return "devices"
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

	fmt.Println("=== DATABASE INVESTIGATION ===")
	
	// 1. Check if devices table exists and show structure
	fmt.Println("\n1. Checking devices table structure:")
	var tableExists bool
	err = db.Raw("SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = ? AND table_name = 'devices'", 
		config.Database.Database).Scan(&tableExists).Error
	if err != nil {
		log.Printf("Error checking table existence: %v", err)
	} else {
		fmt.Printf("Devices table exists: %v\n", tableExists)
	}

	if tableExists {
		// Show table structure
		var columns []struct {
			Field   string `json:"Field"`
			Type    string `json:"Type"`
			Null    string `json:"Null"`
			Key     string `json:"Key"`
			Default *string `json:"Default"`
			Extra   string `json:"Extra"`
		}
		
		err = db.Raw("DESCRIBE devices").Scan(&columns).Error
		if err != nil {
			log.Printf("Error describing table: %v", err)
		} else {
			fmt.Println("Table structure:")
			for _, col := range columns {
				defaultStr := "NULL"
				if col.Default != nil {
					defaultStr = *col.Default
				}
				fmt.Printf("  %s | %s | %s | Key: %s | Default: %s\n", 
					col.Field, col.Type, col.Null, col.Key, defaultStr)
			}
		}

		// 2. Count total devices
		fmt.Println("\n2. Total device count:")
		var totalCount int64
		err = db.Model(&Device{}).Count(&totalCount).Error
		if err != nil {
			log.Printf("Error counting devices: %v", err)
		} else {
			fmt.Printf("Total devices in database: %d\n", totalCount)
		}

		// 3. Show all distinct status values
		fmt.Println("\n3. All distinct status values:")
		var statuses []string
		err = db.Model(&Device{}).Distinct("status").Pluck("status", &statuses).Error
		if err != nil {
			log.Printf("Error getting status values: %v", err)
		} else {
			fmt.Printf("Distinct status values: %v\n", statuses)
		}

		// 4. Count devices by status
		fmt.Println("\n4. Device count by status:")
		var statusCounts []struct {
			Status string `json:"status"`
			Count  int64  `json:"count"`
		}
		err = db.Model(&Device{}).Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts).Error
		if err != nil {
			log.Printf("Error getting status counts: %v", err)
		} else {
			for _, sc := range statusCounts {
				fmt.Printf("  Status '%s': %d devices\n", sc.Status, sc.Count)
			}
		}

		// 5. Show sample devices with all details
		fmt.Println("\n5. Sample devices (first 5):")
		var sampleDevices []Device
		err = db.Limit(5).Find(&sampleDevices).Error
		if err != nil {
			log.Printf("Error getting sample devices: %v", err)
		} else {
			for i, device := range sampleDevices {
				serialNum := "NULL"
				if device.SerialNumber != nil {
					serialNum = *device.SerialNumber
				}
				prodID := "NULL"
				if device.ProductID != nil {
					prodID = fmt.Sprintf("%d", *device.ProductID)
				}
				fmt.Printf("  Device %d: ID='%s', Status='%s', ProductID=%s, SerialNumber='%s'\n", 
					i+1, device.DeviceID, device.Status, prodID, serialNum)
			}
		}

		// 6. Specifically check for devices with status = 'free'
		fmt.Println("\n6. Devices with status = 'free':")
		var freeDevices []Device
		err = db.Where("status = ?", "free").Find(&freeDevices).Error
		if err != nil {
			log.Printf("Error getting free devices: %v", err)
		} else {
			fmt.Printf("Found %d devices with status = 'free'\n", len(freeDevices))
			for i, device := range freeDevices {
				if i < 3 { // Show first 3
					fmt.Printf("  Free device %d: ID='%s', Status='%s'\n", 
						i+1, device.DeviceID, device.Status)
				}
			}
		}

		// 7. Check jobdevices table to see assignments
		fmt.Println("\n7. Check job assignments:")
		var jobDeviceCount int64
		err = db.Table("jobdevices").Count(&jobDeviceCount).Error
		if err != nil {
			log.Printf("Error counting job devices: %v", err)
		} else {
			fmt.Printf("Total job-device assignments: %d\n", jobDeviceCount)
		}

		// 8. Show available devices (the query from GetAvailableDevices)
		fmt.Println("\n8. Available devices (status='free' AND not in jobdevices):")
		var availableDevices []Device
		err = db.Where(`status = 'free' AND deviceID NOT IN (
			SELECT DISTINCT deviceID FROM jobdevices
		)`).Find(&availableDevices).Error
		if err != nil {
			log.Printf("Error getting available devices: %v", err)
		} else {
			fmt.Printf("Found %d available devices\n", len(availableDevices))
			for i, device := range availableDevices {
				if i < 5 { // Show first 5
					fmt.Printf("  Available device %d: ID='%s', Status='%s'\n", 
						i+1, device.DeviceID, device.Status)
				}
			}
		}
	}

	fmt.Println("\n=== INVESTIGATION COMPLETE ===")
}