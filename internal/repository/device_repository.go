package repository

import (
	"sort"
	"time"
	"log"
	"go-barcode-webapp/internal/models"
)

type DeviceRepository struct {
	db *Database
}

func NewDeviceRepository(db *Database) *DeviceRepository {
	return &DeviceRepository{db: db}
}

func (r *DeviceRepository) Create(device *models.Device) error {
	return r.db.Create(device).Error
}

func (r *DeviceRepository) GetByID(deviceID string) (*models.Device, error) {
	var device models.Device
	err := r.db.Where("deviceID = ?", deviceID).Preload("Product").First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *DeviceRepository) GetBySerialNo(serialNo string) (*models.Device, error) {
	var device models.Device
	err := r.db.Where("serialnumber = ?", serialNo).Preload("Product").First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *DeviceRepository) Update(device *models.Device) error {
	return r.db.Save(device).Error
}

func (r *DeviceRepository) Delete(deviceID string) error {
	return r.db.Where("deviceID = ?", deviceID).Delete(&models.Device{}).Error
}

func (r *DeviceRepository) List(params *models.FilterParams) ([]models.DeviceWithJobInfo, error) {
	startTime := time.Now()
	log.Printf("⏱️  DeviceRepository.List() started")

	var devices []models.Device

	// Set default pagination if not provided
	limit := params.Limit
	if limit <= 0 {
		limit = 20 // Reduce default to 20 items per page for better performance
	}
	
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	// Simple query without complex joins for better performance
	query := r.db.Model(&models.Device{})
	
	// Only preload Product if we actually need it
	if params.SearchTerm != "" {
		searchPattern := "%" + params.SearchTerm + "%"
		query = query.Preload("Product").
			Joins("LEFT JOIN products ON products.productID = devices.productID").
			Where("devices.deviceID LIKE ? OR devices.serialnumber LIKE ? OR products.name LIKE ?", searchPattern, searchPattern, searchPattern)
	} else {
		// For normal list view, we don't need full product details
		query = query.Preload("Product")
	}

	query = query.Limit(limit).Offset(offset).Order("deviceID DESC")

	queryStart := time.Now()
	err := query.Find(&devices).Error
	queryTime := time.Since(queryStart)
	log.Printf("⏱️  Device query took: %v", queryTime)
	
	if err != nil {
		log.Printf("❌ Device query error: %v", err)
		return nil, err
	}

	// Skip job assignment check for better performance - we can add it back later if needed
	var result []models.DeviceWithJobInfo
	for _, device := range devices {
		result = append(result, models.DeviceWithJobInfo{
			Device:     device,
			JobID:      nil,
			IsAssigned: false,
		})
	}

	totalTime := time.Since(startTime)
	log.Printf("⏱️  DeviceRepository.List() completed in %v (found %d devices)", totalTime, len(result))

	return result, nil
}

func (r *DeviceRepository) ListWithCategories(params *models.FilterParams) ([]models.Device, error) {
	var devices []models.Device

	query := r.db.Model(&models.Device{}).
		Preload("Product").
		Preload("Product.Category").
		Preload("Product.Subcategory").
		Preload("Product.Brand").
		Preload("Product.Manufacturer")

	// Join products table for search and category filtering
	query = query.Joins("JOIN products ON products.productID = devices.productID")

	if params.SearchTerm != "" {
		searchPattern := "%" + params.SearchTerm + "%"
		query = query.Where("devices.deviceID LIKE ? OR devices.serialnumber LIKE ? OR products.name LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	// Category filter
	if params.Category != "" {
		query = query.Joins("JOIN categories ON categories.categoryID = products.categoryID").
			Where("categories.name = ?", params.Category)
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	query = query.Order("deviceID DESC")

	err := query.Find(&devices).Error
	return devices, err
}

func (r *DeviceRepository) GetByProductID(productID uint) ([]models.Device, error) {
	var devices []models.Device
	err := r.db.Where("productID = ?", productID).
		Preload("Product").
		Order("deviceID ASC").
		Find(&devices).Error
	return devices, err
}

func (r *DeviceRepository) GetAvailableDevices() ([]models.Device, error) {
	var devices []models.Device
	
	// Get devices that are available and not currently assigned to any job
	err := r.db.Where(`status = 'free' AND deviceID NOT IN (
		SELECT DISTINCT deviceID FROM jobdevices
	)`).Find(&devices).Error
	
	return devices, err
}

func (r *DeviceRepository) GetDevicesByCategory(category string) ([]models.Device, error) {
	var devices []models.Device
	err := r.db.Where("category = ? AND available = true", category).
		Find(&devices).Error
	return devices, err
}

func (r *DeviceRepository) CheckDeviceAvailability(deviceID uint) (bool, error) {
	var count int64
	err := r.db.Table("job_devices").
		Where("device_id = ? AND removed_at IS NULL", deviceID).
		Count(&count).Error
	
	return count == 0, err
}

func (r *DeviceRepository) GetDeviceJobHistory(deviceID uint) ([]models.JobDevice, error) {
	var jobDevices []models.JobDevice
	err := r.db.Where("device_id = ?", deviceID).
		Preload("Job").
		Preload("Job.Customer").
		Find(&jobDevices).Error
	
	return jobDevices, err
}

func (r *DeviceRepository) GetAvailableDevicesForCaseManagement() ([]models.Device, error) {
	var devices []models.Device
	
	// Get all devices with product information, regardless of status or case assignment
	err := r.db.Preload("Product").
		Preload("Product.Category").
		Preload("Product.Subcategory").
		Preload("Product.Brand").
		Preload("Product.Manufacturer").
		Find(&devices).Error
	
	return devices, err
}

func (r *DeviceRepository) GetDevicesGroupedByCategory(params *models.FilterParams) (*models.CategorizedDevicesResponse, error) {
	var devices []models.Device
	
	// Set default pagination if not provided
	limit := params.Limit
	if limit <= 0 {
		limit = 100 // Default to 100 items for categorized view
	}
	
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	query := r.db.Model(&models.Device{}).
		Preload("Product").
		Preload("Product.Category").
		Preload("Product.Subcategory").
		Preload("Product.Brand").
		Preload("Product.Manufacturer")

	if params.SearchTerm != "" {
		searchPattern := "%" + params.SearchTerm + "%"
		query = query.Joins("LEFT JOIN products ON products.productID = devices.productID").
			Where("devices.deviceID LIKE ? OR devices.serialnumber LIKE ? OR products.name LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	query = query.Limit(limit).Offset(offset).Order("deviceID DESC")

	err := query.Find(&devices).Error
	if err != nil {
		return nil, err
	}

	// --- Performance Optimization ---
	// 1. Get all device IDs from the current list
	deviceIDs := make([]string, len(devices))
	for i, d := range devices {
		deviceIDs[i] = d.DeviceID
	}

	// 2. Fetch all relevant job assignments in a single query
	var jobDevices []models.JobDevice
	if len(deviceIDs) > 0 {
		r.db.Where("deviceID IN ?", deviceIDs).Find(&jobDevices)
	}

	// 3. Create a map for quick lookup of job assignments
	jobDeviceMap := make(map[string]*models.JobDevice)
	for i, jd := range jobDevices {
		jobDeviceMap[jd.DeviceID] = &jobDevices[i]
	}
	// --- End of Performance Optimization ---

	// Convert to DeviceWithJobInfo and group by categories
	deviceMap := make(map[uint]map[string][]models.DeviceWithJobInfo) // categoryID -> subcategoryID -> devices
	categoryMap := make(map[uint]*models.Category)
	subcategoryMap := make(map[string]*models.Subcategory)
	var uncategorized []models.DeviceWithJobInfo

	for _, device := range devices {
		// Check if device is assigned using the optimized map
		jobDevice, isAssigned := jobDeviceMap[device.DeviceID]
		var jobID *uint
		if isAssigned {
			jobID = &jobDevice.JobID
		}

		deviceWithJob := models.DeviceWithJobInfo{
			Device:     device,
			JobID:      jobID,
			IsAssigned: isAssigned,
		}

		// Group by category and subcategory
		if device.Product != nil && device.Product.Category != nil {
			categoryID := device.Product.Category.CategoryID
			var subcategoryID string
			if device.Product.Subcategory != nil {
				subcategoryID = device.Product.Subcategory.SubcategoryID
			}

			// Store category and subcategory references
			categoryMap[categoryID] = device.Product.Category
			if device.Product.Subcategory != nil {
				subcategoryMap[subcategoryID] = device.Product.Subcategory
			}

			// Initialize maps if needed
			if deviceMap[categoryID] == nil {
				deviceMap[categoryID] = make(map[string][]models.DeviceWithJobInfo)
			}
			if deviceMap[categoryID][subcategoryID] == nil {
				deviceMap[categoryID][subcategoryID] = []models.DeviceWithJobInfo{}
			}

			deviceMap[categoryID][subcategoryID] = append(deviceMap[categoryID][subcategoryID], deviceWithJob)
		} else {
			// Add to uncategorized if product, category, or subcategory is missing
			uncategorized = append(uncategorized, deviceWithJob)
		}
	}

	// Build response structure
	var categories []models.DeviceCategoryGroup
	totalDevices := len(devices)

	for categoryID, subcategoryDevices := range deviceMap {
		category := categoryMap[categoryID]
		var subcategories []models.DeviceSubcategoryGroup
		categoryDeviceCount := 0

		for subcategoryID, devices := range subcategoryDevices {
			var subcategory *models.Subcategory
			if subcategoryID != "" {
				subcategory = subcategoryMap[subcategoryID]
			} else {
				subcategory = &models.Subcategory{Name: "General"}
			}
			subcategories = append(subcategories, models.DeviceSubcategoryGroup{
				Subcategory: subcategory,
				Devices:     devices,
				DeviceCount: len(devices),
			})
			categoryDeviceCount += len(devices)
		}

		// Sort subcategories by name
		sort.Slice(subcategories, func(i, j int) bool {
			if subcategories[i].Subcategory == nil {
				return false
			}
			if subcategories[j].Subcategory == nil {
				return true
			}
			return subcategories[i].Subcategory.Name < subcategories[j].Subcategory.Name
		})

		categories = append(categories, models.DeviceCategoryGroup{
			Category:      category,
			Subcategories: subcategories,
			DeviceCount:   categoryDeviceCount,
		})
	}

	// Sort categories by name
	sort.Slice(categories, func(i, j int) bool {
		if categories[i].Category == nil {
			return false
		}
		if categories[j].Category == nil {
			return true
		}
		return categories[i].Category.Name < categories[j].Category.Name
	})

	return &models.CategorizedDevicesResponse{
		Categories:    categories,
		Uncategorized: uncategorized,
		TotalDevices:  totalDevices,
	}, nil
}

func (r *DeviceRepository) GetDeviceStats(deviceID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Get total number of jobs this device has been assigned to
	var totalJobs int64
	err := r.db.Model(&models.JobDevice{}).
		Where("deviceID = ?", deviceID).
		Count(&totalJobs).Error
	if err != nil {
		log.Printf("Error counting jobs for device %s: %v", deviceID, err)
		totalJobs = 0
	}
	
	// Get total earnings from jobs (simplified calculation)
	var totalEarnings float64
	err = r.db.Raw(`
		SELECT COALESCE(SUM(DATEDIFF(COALESCE(j.endDate, NOW()), j.startDate) * COALESCE(p.itemcostperday, 0)), 0) as total_earnings
		FROM jobdevices jd
		JOIN jobs j ON jd.jobID = j.jobID
		JOIN devices d ON jd.deviceID = d.deviceID
		LEFT JOIN products p ON d.productID = p.productID
		WHERE jd.deviceID = ?
	`, deviceID).Scan(&totalEarnings).Error
	if err != nil {
		log.Printf("Error calculating earnings for device %s: %v", deviceID, err)
		totalEarnings = 0.0
	}
	
	// Get total days rented
	var totalDaysRented int64
	err = r.db.Raw(`
		SELECT COALESCE(SUM(DATEDIFF(COALESCE(j.endDate, NOW()), j.startDate)), 0) as total_days
		FROM jobdevices jd
		JOIN jobs j ON jd.jobID = j.jobID
		WHERE jd.deviceID = ?
	`, deviceID).Scan(&totalDaysRented).Error
	if err != nil {
		log.Printf("Error calculating days rented for device %s: %v", deviceID, err)
		totalDaysRented = 0
	}
	
	// Calculate average rental duration
	var averageRentalDuration float64
	if totalJobs > 0 {
		averageRentalDuration = float64(totalDaysRented) / float64(totalJobs)
	}
	
	// Get device product details for price per day
	var device models.Device
	err = r.db.Where("deviceID = ?", deviceID).Preload("Product").First(&device).Error
	if err != nil {
		log.Printf("Error getting device details for %s: %v", deviceID, err)
	}
	
	var pricePerDay float64
	var weight float64
	if device.Product != nil {
		if device.Product.ItemCostPerDay != nil {
			pricePerDay = *device.Product.ItemCostPerDay
		}
		if device.Product.Weight != nil {
			weight = *device.Product.Weight
		}
	}
	
	stats["totalJobs"] = totalJobs
	stats["totalEarnings"] = totalEarnings
	stats["totalDaysRented"] = totalDaysRented
	stats["averageRentalDuration"] = averageRentalDuration
	stats["pricePerDay"] = pricePerDay
	stats["weight"] = weight
	
	return stats, nil
}