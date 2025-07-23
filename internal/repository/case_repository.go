package repository

import (
	"log"
	"go-barcode-webapp/internal/models"
	"gorm.io/gorm"
)

type CaseRepository struct {
	db *Database
}

func NewCaseRepository(db *Database) *CaseRepository {
	return &CaseRepository{db: db}
}

// GetAll returns all cases
func (r *CaseRepository) GetAll() ([]models.Case, error) {
	var cases []models.Case
	err := r.db.DB.Find(&cases).Error
	if err != nil {
		return cases, err
	}
	
	// Add device counts using simple COUNT queries
	for i := range cases {
		var deviceCount int64
		if err := r.db.DB.Table("devicescases").Where("caseID = ?", cases[i].CaseID).Count(&deviceCount).Error; err != nil {
			deviceCount = 0
		}
		cases[i].DeviceCount = int(deviceCount)
		// Don't load full device data for list view
		cases[i].Devices = []models.DeviceCase{}
	}
	
	return cases, err
}

// GetByID returns a case by ID
func (r *CaseRepository) GetByID(id uint) (*models.Case, error) {
	var case_ models.Case
	err := r.db.DB.Preload("Devices.Device.Product").First(&case_, id).Error
	if err != nil {
		return nil, err
	}
	
	// Add device count
	var deviceCount int64
	if err := r.db.DB.Table("devicescases").Where("caseID = ?", case_.CaseID).Count(&deviceCount).Error; err != nil {
		deviceCount = 0
	}
	case_.DeviceCount = int(deviceCount)
	
	return &case_, nil
}

// Create creates a new case
func (r *CaseRepository) Create(case_ *models.Case) error {
	return r.db.DB.Create(case_).Error
}

// Update updates an existing case
func (r *CaseRepository) Update(case_ *models.Case) error {
	return r.db.DB.Save(case_).Error
}

// Delete deletes a case by ID
func (r *CaseRepository) Delete(id uint) error {
	// First remove all devices from the case
	err := r.db.DB.Where("case_id = ?", id).Delete(&models.DeviceCase{}).Error
	if err != nil {
		return err
	}
	
	// Then delete the case
	return r.db.DB.Delete(&models.Case{}, id).Error
}

// GetDevicesInCase returns all devices assigned to a case
func (r *CaseRepository) GetDevicesInCase(caseID uint) ([]models.DeviceCase, error) {
	var deviceCases []models.DeviceCase
	err := r.db.DB.Preload("Device").
		Preload("Device.Product").
		Preload("Device.Product.Category").
		Preload("Device.Product.Subcategory").
		Where("caseID = ?", caseID).
		Find(&deviceCases).Error
	return deviceCases, err
}

// AddDeviceToCase assigns a device to a case
func (r *CaseRepository) AddDeviceToCase(caseID uint, deviceID string) error {
	// Check if device is already in the case
	var existingDeviceCase models.DeviceCase
	err := r.db.DB.Where("caseID = ? AND deviceID = ?", caseID, deviceID).First(&existingDeviceCase).Error
	if err == nil {
		// Device already in case
		return gorm.ErrDuplicatedKey
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	// Create new device-case relationship
	deviceCase := models.DeviceCase{
		CaseID:   caseID,
		DeviceID: deviceID,
	}
	
	return r.db.DB.Create(&deviceCase).Error
}

// RemoveDeviceFromCase removes a device from a case
func (r *CaseRepository) RemoveDeviceFromCase(caseID uint, deviceID string) error {
	return r.db.DB.Where("caseID = ? AND deviceID = ?", caseID, deviceID).
		Delete(&models.DeviceCase{}).Error
}

// GetAvailableDevices returns devices that are not assigned to any case
func (r *CaseRepository) GetAvailableDevices() ([]models.Device, error) {
	var devices []models.Device
	err := r.db.DB.Preload("Product").
		Where("deviceID NOT IN (SELECT deviceID FROM devicescases)").
		Find(&devices).Error
	return devices, err
}

// GetAvailableDevicesForCase returns devices that are either not assigned to any case
// or are assigned to the specified case (for editing purposes)
func (r *CaseRepository) GetAvailableDevicesForCase(caseID uint) ([]models.Device, error) {
	var devices []models.Device
	err := r.db.DB.Preload("Product").
		Where("deviceID NOT IN (SELECT deviceID FROM devicescases WHERE caseID != ?)", caseID).
		Find(&devices).Error
	return devices, err
}

// GetCasesByCustomer returns all cases for a specific customer (cases don't have customer relationships)
func (r *CaseRepository) GetCasesByCustomer(customerID uint) ([]models.Case, error) {
	var cases []models.Case
	// Cases don't have customer relationships - this method may need to be reconsidered
	err := r.db.DB.Preload("Devices").Find(&cases).Error
	return cases, err
}

// GetDeviceCount returns the number of devices in a case
func (r *CaseRepository) GetDeviceCount(caseID uint) (int64, error) {
	var count int64
	err := r.db.DB.Model(&models.DeviceCase{}).
		Where("case_id = ?", caseID).
		Count(&count).Error
	return count, err
}

// List returns cases with optional filtering
func (r *CaseRepository) List(filter *models.FilterParams) ([]models.Case, error) {
	log.Printf("🔥🔥🔥 CASE LIST CALLED!")
	
	// Direct SQL with COUNT - no bullshit
	sqlQuery := `
		SELECT 
			c.caseID, 
			c.name, 
			c.description, 
			c.width, 
			c.height, 
			c.depth, 
			c.weight, 
			c.status,
			COALESCE(COUNT(dc.deviceID), 0) as device_count
		FROM cases c 
		LEFT JOIN devicescases dc ON c.caseID = dc.caseID`
	
	var args []interface{}
	if filter != nil && filter.SearchTerm != "" {
		sqlQuery += " WHERE c.name LIKE ? OR c.description LIKE ?"
		searchTerm := "%" + filter.SearchTerm + "%"
		args = append(args, searchTerm, searchTerm)
	}
	
	sqlQuery += " GROUP BY c.caseID ORDER BY c.caseID"
	
	log.Printf("🔥🔥🔥 SQL: %s", sqlQuery)
	
	type CaseResult struct {
		CaseID      uint     `json:"caseID"`
		Name        string   `json:"name"`
		Description *string  `json:"description"`
		Width       *float64 `json:"width"`
		Height      *float64 `json:"height"`
		Depth       *float64 `json:"depth"`
		Weight      *float64 `json:"weight"`
		Status      string   `json:"status"`
		DeviceCount int      `json:"device_count"`
	}
	
	var results []CaseResult
	err := r.db.DB.Raw(sqlQuery, args...).Scan(&results).Error
	if err != nil {
		log.Printf("🔥🔥🔥 SQL ERROR: %v", err)
		return nil, err
	}
	
	log.Printf("🔥🔥🔥 FOUND %d CASES", len(results))
	
	var cases []models.Case
	for _, result := range results {
		log.Printf("🔥🔥🔥 CASE %d ('%s') = %d DEVICES", result.CaseID, result.Name, result.DeviceCount)
		
		case_ := models.Case{
			CaseID:      result.CaseID,
			Name:        result.Name,
			Description: result.Description,
			Width:       result.Width,
			Height:      result.Height,
			Depth:       result.Depth,
			Weight:      result.Weight,
			Status:      result.Status,
			DeviceCount: result.DeviceCount,
			Devices:     []models.DeviceCase{},
		}
		
		cases = append(cases, case_)
	}
	
	log.Printf("🔥🔥🔥 RETURNING %d CASES", len(cases))
	return cases, nil
}

// IsDeviceInAnyCase checks if a device is assigned to any case
func (r *CaseRepository) IsDeviceInAnyCase(deviceID string) (bool, error) {
	var count int64
	err := r.db.DB.Model(&models.DeviceCase{}).
		Where("deviceID = ?", deviceID).
		Count(&count).Error
	return count > 0, err
}