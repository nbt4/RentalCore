package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"go-barcode-webapp/internal/models"

	"gorm.io/gorm"
)

type JobTemplateRepository struct {
	db *Database
}

func NewJobTemplateRepository(db *Database) *JobTemplateRepository {
	return &JobTemplateRepository{db: db}
}

// List returns all job templates with optional filtering
func (r *JobTemplateRepository) List(params *models.FilterParams) ([]models.JobTemplate, error) {
	var templates []models.JobTemplate
	
	query := r.db.DB.Model(&models.JobTemplate{}).
		Preload("JobCategory").
		Preload("Creator")
	
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
	}
	
	err := query.Find(&templates).Error
	return templates, err
}

// GetByID returns a job template by ID
func (r *JobTemplateRepository) GetByID(id uint) (*models.JobTemplate, error) {
	var template models.JobTemplate
	err := r.db.DB.Preload("JobCategory").
		Preload("Creator").
		First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// Create creates a new job template
func (r *JobTemplateRepository) Create(template *models.JobTemplate) error {
	// Set timestamps
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	
	return r.db.DB.Create(template).Error
}

// Update updates an existing job template
func (r *JobTemplateRepository) Update(template *models.JobTemplate) error {
	// Update timestamp
	template.UpdatedAt = time.Now()
	
	return r.db.DB.Save(template).Error
}

// Delete soft deletes a job template
func (r *JobTemplateRepository) Delete(id uint) error {
	return r.db.DB.Model(&models.JobTemplate{}).
		Where("templateID = ?", id).
		Update("is_active", false).Error
}

// IncrementUsageCount increments the usage count for a template
func (r *JobTemplateRepository) IncrementUsageCount(id uint) error {
	return r.db.DB.Model(&models.JobTemplate{}).
		Where("templateID = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

// GetMostUsed returns the most used templates
func (r *JobTemplateRepository) GetMostUsed(limit int) ([]models.JobTemplate, error) {
	var templates []models.JobTemplate
	err := r.db.DB.Preload("JobCategory").
		Preload("Creator").
		Where("is_active = ?", true).
		Order("usage_count DESC").
		Limit(limit).
		Find(&templates).Error
	return templates, err
}

// GetByCategory returns templates for a specific job category
func (r *JobTemplateRepository) GetByCategory(categoryID uint) ([]models.JobTemplate, error) {
	var templates []models.JobTemplate
	err := r.db.DB.Preload("JobCategory").
		Preload("Creator").
		Where("jobcategoryID = ? AND is_active = ?", categoryID, true).
		Order("name ASC").
		Find(&templates).Error
	return templates, err
}

// ValidateEquipmentList validates the equipment list JSON structure
func (r *JobTemplateRepository) ValidateEquipmentList(equipmentList []byte) error {
	if len(equipmentList) == 0 {
		return nil // Empty is valid
	}
	
	var equipment []map[string]interface{}
	if err := json.Unmarshal(equipmentList, &equipment); err != nil {
		return fmt.Errorf("invalid equipment list format: %w", err)
	}
	
	// Validate each equipment item has required fields
	for i, item := range equipment {
		if _, ok := item["deviceID"]; !ok {
			return fmt.Errorf("equipment item %d missing deviceID", i)
		}
		if _, ok := item["quantity"]; !ok {
			return fmt.Errorf("equipment item %d missing quantity", i)
		}
	}
	
	return nil
}

// ValidatePricingTemplate validates the pricing template JSON structure
func (r *JobTemplateRepository) ValidatePricingTemplate(pricingTemplate []byte) error {
	if len(pricingTemplate) == 0 {
		return nil // Empty is valid
	}
	
	var pricing map[string]interface{}
	if err := json.Unmarshal(pricingTemplate, &pricing); err != nil {
		return fmt.Errorf("invalid pricing template format: %w", err)
	}
	
	// Validate pricing structure
	if basePrice, ok := pricing["basePrice"]; ok {
		if _, isFloat := basePrice.(float64); !isFloat {
			return fmt.Errorf("basePrice must be a number")
		}
	}
	
	return nil
}

// Count returns the total number of templates
func (r *JobTemplateRepository) Count(params *models.FilterParams) (int64, error) {
	var count int64
	query := r.db.DB.Model(&models.JobTemplate{})
	
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
	
	err := query.Count(&count).Error
	return count, err
}