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
			query = query.Where("name ILIKE ? OR description ILIKE ?", 
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
			query = query.Where("name ILIKE ? OR description ILIKE ?", 
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
	
	// First, let's check if there are ANY devices at all
	var totalCount int64
	if err := r.db.DB.Model(&models.Device{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count devices: %v", err)
	}
	
	fmt.Printf("DEBUG: Total devices in database: %d\n", totalCount)
	
	if totalCount == 0 {
		fmt.Printf("DEBUG: No devices found in database at all\n")
		return []models.Device{}, nil
	}
	
	// Check what status values exist
	var statusCounts []struct {
		Status string
		Count  int64
	}
	if err := r.db.DB.Model(&models.Device{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		fmt.Printf("DEBUG: Error checking status counts: %v\n", err)
	} else {
		fmt.Printf("DEBUG: Device status counts:\n")
		for _, sc := range statusCounts {
			fmt.Printf("  Status '%s': %d devices\n", sc.Status, sc.Count)
		}
	}
	
	// Try to get devices with common status values
	if err := r.db.DB.Where("status IN (?)", []string{"free", "available", "ready", ""}).
		Preload("Product").
		Order("deviceID ASC").
		Find(&devices).Error; err != nil {
		fmt.Printf("DEBUG: Error querying devices with status filter: %v\n", err)
		return nil, fmt.Errorf("failed to get available devices: %v", err)
	}
	
	fmt.Printf("DEBUG: Found %d devices with filtered status\n", len(devices))
	
	// If still no devices found, get ALL devices (for debugging)
	if len(devices) == 0 {
		fmt.Printf("DEBUG: No devices found with status filter, getting all devices...\n")
		if err := r.db.DB.Preload("Product").
			Order("deviceID ASC").
			Limit(10). // Limit to first 10 for debugging
			Find(&devices).Error; err != nil {
			fmt.Printf("DEBUG: Error getting all devices: %v\n", err)
			return nil, fmt.Errorf("failed to get any devices: %v", err)
		}
		fmt.Printf("DEBUG: Found %d total devices (ignoring status)\n", len(devices))
	}
	
	// Show sample devices
	for i, device := range devices {
		if i >= 3 { break } // Only show first 3
		productName := "No Product"
		if device.Product != nil {
			productName = device.Product.Name
		}
		fmt.Printf("DEBUG: Device %d: ID='%s', Status='%s', Product='%s'\n", i+1, device.DeviceID, device.Status, productName)
	}
	
	return devices, nil
}