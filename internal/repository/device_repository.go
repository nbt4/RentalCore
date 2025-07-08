package repository

import (
	"log"
	"runtime/debug"
	"sort"
	"go-barcode-webapp/internal/models"
	"gorm.io/gorm"
)

type DeviceRepository struct {
	db *Database
}

func NewDeviceRepository(db *Database) *DeviceRepository {
	return &DeviceRepository{db: db}
}

func (r *DeviceRepository) Create(device *models.Device) error {
	// DEBUG: Log device creation attempts to track automatic creation
	log.Printf("🚨 DEVICE CREATION ATTEMPT:")
	log.Printf("   DeviceID: '%s'", device.DeviceID)
	if device.ProductID != nil {
		log.Printf("   ProductID: %d", *device.ProductID)
	} else {
		log.Printf("   ProductID: NULL")
	}
	if device.SerialNumber != nil {
		log.Printf("   SerialNumber: '%s'", *device.SerialNumber)
	} else {
		log.Printf("   SerialNumber: NULL")
	}
	log.Printf("   Status: '%s'", device.Status)
	
	// Print stack trace to see what's calling this
	log.Printf("   📍 Stack trace:")
	log.Printf("%s", debug.Stack())
	
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
	var devices []models.Device

	query := r.db.Model(&models.Device{}).Preload("Product")

	if params.SearchTerm != "" {
		searchPattern := "%" + params.SearchTerm + "%"
		query = query.Where("deviceID LIKE ? OR serialnumber LIKE ?", searchPattern, searchPattern)
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	query = query.Order("deviceID DESC")

	err := query.Find(&devices).Error
	if err != nil {
		return nil, err
	}

	// Convert to DeviceWithJobInfo
	var result []models.DeviceWithJobInfo
	for _, device := range devices {
		// Check if device is assigned to any job
		var jobDevice models.JobDevice
		isAssigned := false
		var jobID *uint
		
		err := r.db.Where("deviceID = ?", device.DeviceID).First(&jobDevice).Error
		if err == nil {
			isAssigned = true
			jobID = &jobDevice.JobID
		} else if err != gorm.ErrRecordNotFound {
			// If there's an error other than "not found", we should handle it
			// For now, we'll just continue and assume not assigned
		}

		result = append(result, models.DeviceWithJobInfo{
			Device:     device,
			JobID:      jobID,
			IsAssigned: isAssigned,
		})
	}

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

	if params.SearchTerm != "" {
		searchPattern := "%" + params.SearchTerm + "%"
		query = query.Where("deviceID LIKE ? OR serialnumber LIKE ?", searchPattern, searchPattern)
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
	
	query := r.db.Model(&models.Device{}).
		Preload("Product").
		Preload("Product.Category").
		Preload("Product.Subcategory").
		Preload("Product.Brand").
		Preload("Product.Manufacturer")

	if params.SearchTerm != "" {
		searchPattern := "%" + params.SearchTerm + "%"
		query = query.Where("deviceID LIKE ? OR serialnumber LIKE ?", searchPattern, searchPattern)
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	query = query.Order("deviceID DESC")

	err := query.Find(&devices).Error
	if err != nil {
		return nil, err
	}

	// Convert to DeviceWithJobInfo and group by categories
	deviceMap := make(map[uint]map[string][]models.DeviceWithJobInfo) // categoryID -> subcategoryID -> devices
	categoryMap := make(map[uint]*models.Category)
	subcategoryMap := make(map[string]*models.Subcategory)
	var uncategorized []models.DeviceWithJobInfo

	for _, device := range devices {
		// Check if device is assigned to any job
		var jobDevice models.JobDevice
		isAssigned := false
		var jobID *uint
		
		err := r.db.Where("deviceID = ?", device.DeviceID).First(&jobDevice).Error
		if err == nil {
			isAssigned = true
			jobID = &jobDevice.JobID
		} else if err != gorm.ErrRecordNotFound {
			// If there's an error other than "not found", continue and assume not assigned
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