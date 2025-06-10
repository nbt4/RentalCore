package repository

import (
	"go-barcode-webapp/internal/models"
)

type CaseRepository struct {
	db *Database
}

func NewCaseRepository(db *Database) *CaseRepository {
	return &CaseRepository{db: db}
}

func (r *CaseRepository) Create(case_ *models.Case) error {
	return r.db.Create(case_).Error
}

func (r *CaseRepository) GetByID(caseID uint) (*models.Case, error) {
	var case_ models.Case
	err := r.db.Where("caseID = ?", caseID).Preload("Devices").Preload("Devices.Device").First(&case_).Error
	if err != nil {
		return nil, err
	}
	return &case_, nil
}

func (r *CaseRepository) Update(case_ *models.Case) error {
	return r.db.Save(case_).Error
}

func (r *CaseRepository) Delete(caseID uint) error {
	// First delete all device mappings
	if err := r.db.Where("caseID = ?", caseID).Delete(&models.DeviceCase{}).Error; err != nil {
		return err
	}
	// Then delete the case
	return r.db.Where("caseID = ?", caseID).Delete(&models.Case{}).Error
}

func (r *CaseRepository) List(params *models.FilterParams) ([]models.Case, error) {
	var cases []models.Case

	query := r.db.Model(&models.Case{}).Preload("Devices").Preload("Devices.Device")

	if params.SearchTerm != "" {
		searchPattern := "%" + params.SearchTerm + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	query = query.Order("caseID DESC")

	err := query.Find(&cases).Error
	return cases, err
}

func (r *CaseRepository) GetDevicesInCase(caseID uint) ([]models.DeviceCase, error) {
	var deviceCases []models.DeviceCase
	err := r.db.Where("caseID = ?", caseID).Preload("Device").Preload("Device.Product").Find(&deviceCases).Error
	return deviceCases, err
}

func (r *CaseRepository) AddDeviceToCase(caseID uint, deviceID string) error {
	deviceCase := models.DeviceCase{
		CaseID:   caseID,
		DeviceID: deviceID,
	}
	return r.db.Create(&deviceCase).Error
}

func (r *CaseRepository) RemoveDeviceFromCase(caseID uint, deviceID string) error {
	return r.db.Where("caseID = ? AND deviceID = ?", caseID, deviceID).Delete(&models.DeviceCase{}).Error
}

func (r *CaseRepository) GetDeviceCase(deviceID string) (*models.DeviceCase, error) {
	var deviceCase models.DeviceCase
	err := r.db.Where("deviceID = ?", deviceID).Preload("Case").First(&deviceCase).Error
	if err != nil {
		return nil, err
	}
	return &deviceCase, nil
}

func (r *CaseRepository) IsDeviceInAnyCase(deviceID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.DeviceCase{}).Where("deviceID = ?", deviceID).Count(&count).Error
	return count > 0, err
}