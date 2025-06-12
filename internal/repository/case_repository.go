package repository

import (
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
	err := r.db.DB.Preload("Devices").Find(&cases).Error
	return cases, err
}

// GetByID returns a case by ID
func (r *CaseRepository) GetByID(id uint) (*models.Case, error) {
	var case_ models.Case
	err := r.db.DB.Preload("Devices").First(&case_, id).Error
	if err != nil {
		return nil, err
	}
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
	query := r.db.DB.Preload("Devices")
	
	if filter != nil && filter.SearchTerm != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", 
			"%"+filter.SearchTerm+"%", "%"+filter.SearchTerm+"%")
	}
	
	var cases []models.Case
	err := query.Find(&cases).Error
	return cases, err
}

// IsDeviceInAnyCase checks if a device is assigned to any case
func (r *CaseRepository) IsDeviceInAnyCase(deviceID string) (bool, error) {
	var count int64
	err := r.db.DB.Model(&models.DeviceCase{}).
		Where("deviceID = ?", deviceID).
		Count(&count).Error
	return count > 0, err
}