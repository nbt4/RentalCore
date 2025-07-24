package repository

import (
	"log"
	"go-barcode-webapp/internal/models"
)

type CableRepository struct {
	db *Database
}

func NewCableRepository(db *Database) *CableRepository {
	return &CableRepository{db: db}
}

func (r *CableRepository) Create(cable *models.Cable) error {
	return r.db.Create(cable).Error
}

func (r *CableRepository) GetByID(id int) (*models.Cable, error) {
	var cable models.Cable
	err := r.db.Preload("Connector1Info").Preload("Connector2Info").Preload("TypeInfo").First(&cable, id).Error
	if err != nil {
		return nil, err
	}
	return &cable, nil
}

func (r *CableRepository) Update(cable *models.Cable) error {
	return r.db.Save(cable).Error
}

func (r *CableRepository) Delete(id int) error {
	return r.db.Delete(&models.Cable{}, id).Error
}

func (r *CableRepository) List(params *models.FilterParams) ([]models.Cable, error) {
	var cables []models.Cable

	query := r.db.Model(&models.Cable{}).
		Preload("Connector1Info").
		Preload("Connector2Info").
		Preload("TypeInfo")

	if params.SearchTerm != "" {
		searchPattern := "%" + params.SearchTerm + "%"
		query = query.Where("name LIKE ?", searchPattern)
	}

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	query = query.Order("name ASC")

	err := query.Find(&cables).Error
	return cables, err
}

func (r *CableRepository) GetTotalCount() (int, error) {
	var count int64
	err := r.db.Model(&models.Cable{}).Count(&count).Error
	return int(count), err
}

// Get all cable types for forms
func (r *CableRepository) GetAllCableTypes() ([]models.CableType, error) {
	var types []models.CableType
	err := r.db.Order("name ASC").Find(&types).Error
	if err != nil {
		log.Printf("❌ GetAllCableTypes error: %v", err)
		return nil, err
	}
	return types, nil
}

// Get all cable connectors for forms
func (r *CableRepository) GetAllCableConnectors() ([]models.CableConnector, error) {
	var connectors []models.CableConnector
	err := r.db.Order("name ASC").Find(&connectors).Error
	if err != nil {
		log.Printf("❌ GetAllCableConnectors error: %v", err)
		return nil, err
	}
	return connectors, nil
}