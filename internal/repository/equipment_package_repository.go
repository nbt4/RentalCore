package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"go-barcode-webapp/internal/models"

	"gorm.io/gorm"
)

type EquipmentPackageRepository struct {
	db *Database
}

func NewEquipmentPackageRepository(db *Database) *EquipmentPackageRepository {
	return &EquipmentPackageRepository{db: db}
}

// List returns all equipment packages with optional filtering
func (r *EquipmentPackageRepository) List(params *models.FilterParams) ([]models.EquipmentPackage, error) {
	var packages []models.EquipmentPackage
	
	query := r.db.DB.Model(&models.EquipmentPackage{})
	
	// Apply filters
	if params != nil {
		if params.SearchTerm != "" {
			query = query.Where("name LIKE ? OR description LIKE ?", 
				"%"+params.SearchTerm+"%", "%"+params.SearchTerm+"%")
		}
		
		if params.Category != "" {
			isActive := params.Category == "active"
			query = query.Where("is_active = ?", isActive)
		}
		
		// Add pagination
		if params.Limit > 0 {
			query = query.Limit(params.Limit)
		}
		if params.Offset > 0 {
			query = query.Offset(params.Offset)
		}
		
		// Default sorting by created_at DESC
		query = query.Order("created_at DESC")
	} else {
		query = query.Order("created_at DESC")
	}
	
	if err := query.Find(&packages).Error; err != nil {
		return nil, fmt.Errorf("failed to list equipment packages: %v", err)
	}
	
	return packages, nil
}

// GetByID returns a specific equipment package by ID
func (r *EquipmentPackageRepository) GetByID(id uint) (*models.EquipmentPackage, error) {
	var pkg models.EquipmentPackage
	
	if err := r.db.DB.First(&pkg, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("equipment package not found")
		}
		return nil, fmt.Errorf("failed to get equipment package: %v", err)
	}
	
	return &pkg, nil
}

// Create creates a new equipment package
func (r *EquipmentPackageRepository) Create(pkg *models.EquipmentPackage) error {
	// Set created_at timestamp
	now := time.Now()
	pkg.CreatedAt = now
	pkg.UpdatedAt = now
	
	// Ensure package items is valid JSON
	if pkg.PackageItems == nil {
		pkg.PackageItems = json.RawMessage("[]")
	}
	
	if err := r.db.DB.Create(pkg).Error; err != nil {
		return fmt.Errorf("failed to create equipment package: %v", err)
	}
	
	return nil
}

// Update updates an existing equipment package
func (r *EquipmentPackageRepository) Update(pkg *models.EquipmentPackage) error {
	// Set updated_at timestamp
	pkg.UpdatedAt = time.Now()
	
	// Ensure package items is valid JSON
	if pkg.PackageItems == nil {
		pkg.PackageItems = json.RawMessage("[]")
	}
	
	if err := r.db.DB.Save(pkg).Error; err != nil {
		return fmt.Errorf("failed to update equipment package: %v", err)
	}
	
	return nil
}

// Delete deletes an equipment package by ID
func (r *EquipmentPackageRepository) Delete(id uint) error {
	if err := r.db.DB.Delete(&models.EquipmentPackage{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete equipment package: %v", err)
	}
	
	return nil
}

// GetTotalCount returns the total count of equipment packages
func (r *EquipmentPackageRepository) GetTotalCount(params *models.FilterParams) (int64, error) {
	var count int64
	
	query := r.db.DB.Model(&models.EquipmentPackage{})
	
	// Apply same filters as List for consistent counting
	if params != nil {
		if params.SearchTerm != "" {
			query = query.Where("name LIKE ? OR description LIKE ?", 
				"%"+params.SearchTerm+"%", "%"+params.SearchTerm+"%")
		}
		
		if params.Category != "" {
			isActive := params.Category == "active"
			query = query.Where("is_active = ?", isActive)
		}
	}
	
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count equipment packages: %v", err)
	}
	
	return count, nil
}

// GetActivePackages returns only active equipment packages
func (r *EquipmentPackageRepository) GetActivePackages() ([]models.EquipmentPackage, error) {
	var packages []models.EquipmentPackage
	
	if err := r.db.DB.Where("is_active = ?", true).
		Order("name ASC").
		Find(&packages).Error; err != nil {
		return nil, fmt.Errorf("failed to get active equipment packages: %v", err)
	}
	
	return packages, nil
}

// GetWithDevices returns a package with its associated devices
func (r *EquipmentPackageRepository) GetWithDevices(id uint) (*models.EquipmentPackage, error) {
	var pkg models.EquipmentPackage
	
	if err := r.db.DB.Preload("PackageDevices").Preload("PackageDevices.Device").
		First(&pkg, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("equipment package not found")
		}
		return nil, fmt.Errorf("failed to get equipment package: %v", err)
	}
	
	return &pkg, nil
}

// CreateWithDevices creates a package and associates devices with it
func (r *EquipmentPackageRepository) CreateWithDevices(pkg *models.EquipmentPackage, deviceMappings []models.PackageDevice) error {
	return r.db.DB.Transaction(func(tx *gorm.DB) error {
		// Create the package first
		now := time.Now()
		pkg.CreatedAt = now
		pkg.UpdatedAt = now
		
		if pkg.PackageItems == nil {
			pkg.PackageItems = json.RawMessage("[]")
		}
		
		if err := tx.Create(pkg).Error; err != nil {
			return fmt.Errorf("failed to create equipment package: %v", err)
		}
		
		// Create device associations
		for i := range deviceMappings {
			deviceMappings[i].PackageID = pkg.PackageID
			deviceMappings[i].CreatedAt = now
			deviceMappings[i].UpdatedAt = now
		}
		
		if len(deviceMappings) > 0 {
			if err := tx.Create(&deviceMappings).Error; err != nil {
				return fmt.Errorf("failed to create package device associations: %v", err)
			}
		}
		
		return nil
	})
}

// UpdateDeviceAssociations updates the devices associated with a package
func (r *EquipmentPackageRepository) UpdateDeviceAssociations(packageID uint, deviceMappings []models.PackageDevice) error {
	return r.db.DB.Transaction(func(tx *gorm.DB) error {
		// Delete existing associations
		if err := tx.Where("package_id = ?", packageID).Delete(&models.PackageDevice{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing device associations: %v", err)
		}
		
		// Create new associations
		now := time.Now()
		for i := range deviceMappings {
			deviceMappings[i].PackageID = packageID
			deviceMappings[i].CreatedAt = now
			deviceMappings[i].UpdatedAt = now
		}
		
		if len(deviceMappings) > 0 {
			if err := tx.Create(&deviceMappings).Error; err != nil {
				return fmt.Errorf("failed to create new device associations: %v", err)
			}
		}
		
		return nil
	})
}

// GetAvailableDevices returns devices that can be added to packages
func (r *EquipmentPackageRepository) GetAvailableDevices() ([]models.Device, error) {
	var devices []models.Device
	
	// Get devices with common available status values
	if err := r.db.DB.Where("status IN (?)", []string{"free", "available", "ready"}).
		Preload("Product").
		Order("deviceID ASC").
		Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("failed to get available devices: %v", err)
	}
	
	return devices, nil
}

// GetPackagesByCategory returns packages filtered by category
func (r *EquipmentPackageRepository) GetPackagesByCategory(category string) ([]models.EquipmentPackage, error) {
	var packages []models.EquipmentPackage
	
	query := r.db.DB.Where("is_active = ?", true)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	
	if err := query.Order("name ASC").Find(&packages).Error; err != nil {
		return nil, fmt.Errorf("failed to get packages by category: %v", err)
	}
	
	return packages, nil
}

// GetPopularPackages returns most used packages
func (r *EquipmentPackageRepository) GetPopularPackages(limit int) ([]models.EquipmentPackage, error) {
	var packages []models.EquipmentPackage
	
	if err := r.db.DB.Where("is_active = ? AND usage_count > 0", true).
		Order("usage_count DESC, name ASC").
		Limit(limit).
		Find(&packages).Error; err != nil {
		return nil, fmt.Errorf("failed to get popular packages: %v", err)
	}
	
	return packages, nil
}

// IncrementUsageCount increments the usage count for a package
func (r *EquipmentPackageRepository) IncrementUsageCount(packageID uint) error {
	now := time.Now()
	
	if err := r.db.DB.Model(&models.EquipmentPackage{}).
		Where("package_id = ?", packageID).
		Updates(map[string]interface{}{
			"usage_count": gorm.Expr("usage_count + 1"),
			"last_used_at": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to increment usage count: %v", err)
	}
	
	return nil
}

// UpdateRevenue updates the total revenue for a package
func (r *EquipmentPackageRepository) UpdateRevenue(packageID uint, revenue float64) error {
	if err := r.db.DB.Model(&models.EquipmentPackage{}).
		Where("package_id = ?", packageID).
		Update("total_revenue", gorm.Expr("total_revenue + ?", revenue)).Error; err != nil {
		return fmt.Errorf("failed to update package revenue: %v", err)
	}
	
	return nil
}

// GetPackageStats returns statistics for a package
func (r *EquipmentPackageRepository) GetPackageStats(packageID uint) (map[string]interface{}, error) {
	var stats struct {
		DeviceCount     int64   `json:"deviceCount"`
		RequiredDevices int64   `json:"requiredDevices"`
		TotalQuantity   int64   `json:"totalQuantity"`
		CalculatedPrice float64 `json:"calculatedPrice"`
	}
	
	// Get device statistics
	if err := r.db.DB.Model(&models.PackageDevice{}).
		Select(`
			COUNT(*) as device_count,
			SUM(CASE WHEN is_required THEN 1 ELSE 0 END) as required_devices,
			SUM(quantity) as total_quantity
		`).
		Where("package_id = ?", packageID).
		Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get package device stats: %v", err)
	}
	
	// Calculate estimated price
	var priceData []struct {
		CustomPrice *float64
		ProductPrice *float64
		Quantity uint
	}
	
	if err := r.db.DB.Model(&models.PackageDevice{}).
		Select("package_devices.custom_price, products.item_cost_per_day as product_price, package_devices.quantity").
		Joins("LEFT JOIN devices ON package_devices.device_id = devices.device_id").
		Joins("LEFT JOIN products ON devices.product_id = products.product_id").
		Where("package_devices.package_id = ?", packageID).
		Scan(&priceData).Error; err != nil {
		return nil, fmt.Errorf("failed to get price data: %v", err)
	}
	
	for _, pd := range priceData {
		price := 0.0
		if pd.CustomPrice != nil {
			price = *pd.CustomPrice
		} else if pd.ProductPrice != nil {
			price = *pd.ProductPrice
		}
		stats.CalculatedPrice += price * float64(pd.Quantity)
	}
	
	return map[string]interface{}{
		"deviceCount":     stats.DeviceCount,
		"requiredDevices": stats.RequiredDevices,
		"totalQuantity":   stats.TotalQuantity,
		"calculatedPrice": stats.CalculatedPrice,
	}, nil
}

// ValidatePackageDevices validates that all devices in a package are still available
func (r *EquipmentPackageRepository) ValidatePackageDevices(packageID uint) (bool, []string, error) {
	var invalidDevices []string
	
	var packageDevices []models.PackageDevice
	if err := r.db.DB.Where("package_id = ?", packageID).Find(&packageDevices).Error; err != nil {
		return false, nil, fmt.Errorf("failed to get package devices: %v", err)
	}
	
	for _, pd := range packageDevices {
		var device models.Device
		if err := r.db.DB.First(&device, "device_id = ?", pd.DeviceID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				invalidDevices = append(invalidDevices, pd.DeviceID+" (not found)")
			} else {
				return false, nil, fmt.Errorf("failed to check device %s: %v", pd.DeviceID, err)
			}
		} else if device.Status != "free" && device.Status != "available" && device.Status != "ready" {
			invalidDevices = append(invalidDevices, pd.DeviceID+" (status: "+device.Status+")")
		}
	}
	
	return len(invalidDevices) == 0, invalidDevices, nil
}

// Search searches packages by name, description, category, or tags
func (r *EquipmentPackageRepository) Search(query string, params *models.FilterParams) ([]models.EquipmentPackage, error) {
	var packages []models.EquipmentPackage
	
	dbQuery := r.db.DB.Model(&models.EquipmentPackage{}).
		Where("name LIKE ? OR description LIKE ? OR category LIKE ? OR tags LIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	
	// Apply additional filters
	if params != nil {
		if params.Category != "" {
			if params.Category == "active" {
				dbQuery = dbQuery.Where("is_active = ?", true)
			} else if params.Category == "inactive" {
				dbQuery = dbQuery.Where("is_active = ?", false)
			} else {
				dbQuery = dbQuery.Where("category = ?", params.Category)
			}
		}
		
		// Pagination
		if params.Limit > 0 {
			dbQuery = dbQuery.Limit(params.Limit)
		}
		if params.Offset > 0 {
			dbQuery = dbQuery.Offset(params.Offset)
		}
	}
	
	if err := dbQuery.Order("name ASC").Find(&packages).Error; err != nil {
		return nil, fmt.Errorf("failed to search packages: %v", err)
	}
	
	return packages, nil
}