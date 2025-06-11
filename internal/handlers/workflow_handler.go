package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go-barcode-webapp/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WorkflowHandler struct {
	db *gorm.DB
}

func NewWorkflowHandler(db *gorm.DB) *WorkflowHandler {
	return &WorkflowHandler{db: db}
}

// ================================================================
// JOB TEMPLATES
// ================================================================

// ListJobTemplates displays all job templates
func (h *WorkflowHandler) ListJobTemplates(c *gin.Context) {
	var templates []models.JobTemplate
	
	result := h.db.Preload("JobCategory").Preload("Creator").
		Order("usage_count DESC, name ASC").
		Find(&templates)
	
	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title": "Error",
			"error": "Failed to load job templates",
		})
		return
	}

	user, _ := GetCurrentUser(c)
	c.HTML(http.StatusOK, "job_templates_list.html", gin.H{
		"title":     "Job Templates",
		"user":      user,
		"templates": templates,
	})
}

// NewJobTemplateForm shows the form to create a new job template
func (h *WorkflowHandler) NewJobTemplateForm(c *gin.Context) {
	var jobCategories []models.JobCategory
	h.db.Find(&jobCategories)

	user, _ := GetCurrentUser(c)
	c.HTML(http.StatusOK, "job_template_form.html", gin.H{
		"title":         "New Job Template",
		"user":          user,
		"jobCategories": jobCategories,
		"isEdit":        false,
	})
}

// CreateJobTemplate creates a new job template
func (h *WorkflowHandler) CreateJobTemplate(c *gin.Context) {
	var request struct {
		Name                string                 `json:"name" binding:"required"`
		Description         string                 `json:"description"`
		JobCategoryID       uint                   `json:"jobCategoryID" binding:"required"`
		DefaultDurationDays int                    `json:"defaultDurationDays"`
		EquipmentList       []map[string]interface{} `json:"equipmentList"`
		DefaultNotes        string                 `json:"defaultNotes"`
		PricingTemplate     map[string]interface{} `json:"pricingTemplate"`
		RequiredDocuments   []string               `json:"requiredDocuments"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Convert to JSON
	equipmentJSON, _ := json.Marshal(request.EquipmentList)
	pricingJSON, _ := json.Marshal(request.PricingTemplate)
	documentsJSON, _ := json.Marshal(request.RequiredDocuments)

	template := models.JobTemplate{
		Name:                request.Name,
		Description:         request.Description,
		JobCategoryID:       &request.JobCategoryID,
		DefaultDurationDays: request.DefaultDurationDays,
		EquipmentList:       equipmentJSON,
		DefaultNotes:        request.DefaultNotes,
		PricingTemplate:     pricingJSON,
		RequiredDocuments:   documentsJSON,
		CreatedBy:           &currentUser.UserID,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		IsActive:            true,
		UsageCount:          0,
	}

	if err := h.db.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job template"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Job template created successfully",
		"templateID": template.TemplateID,
	})
}

// GetJobTemplate retrieves a specific job template
func (h *WorkflowHandler) GetJobTemplate(c *gin.Context) {
	templateID := c.Param("id")

	var template models.JobTemplate
	result := h.db.Preload("JobCategory").Preload("Creator").
		First(&template, templateID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"title": "Template Not Found",
				"error": "Job template not found",
			})
		} else {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"title": "Error",
				"error": "Failed to load job template",
			})
		}
		return
	}

	user, _ := GetCurrentUser(c)
	c.HTML(http.StatusOK, "job_template_detail.html", gin.H{
		"title":    "Job Template: " + template.Name,
		"user":     user,
		"template": template,
	})
}

// EditJobTemplateForm shows the form to edit a job template
func (h *WorkflowHandler) EditJobTemplateForm(c *gin.Context) {
	templateID := c.Param("id")

	var template models.JobTemplate
	result := h.db.Preload("JobCategory").First(&template, templateID)

	if result.Error != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title": "Template Not Found",
			"error": "Job template not found",
		})
		return
	}

	var jobCategories []models.JobCategory
	h.db.Find(&jobCategories)

	user, _ := GetCurrentUser(c)
	c.HTML(http.StatusOK, "job_template_form.html", gin.H{
		"title":         "Edit Job Template",
		"user":          user,
		"template":      template,
		"jobCategories": jobCategories,
		"isEdit":        true,
	})
}

// UpdateJobTemplate updates an existing job template
func (h *WorkflowHandler) UpdateJobTemplate(c *gin.Context) {
	templateID := c.Param("id")

	var template models.JobTemplate
	if err := h.db.First(&template, templateID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job template not found"})
		return
	}

	var request struct {
		Name                string                 `json:"name" binding:"required"`
		Description         string                 `json:"description"`
		JobCategoryID       uint                   `json:"jobCategoryID" binding:"required"`
		DefaultDurationDays int                    `json:"defaultDurationDays"`
		EquipmentList       []map[string]interface{} `json:"equipmentList"`
		DefaultNotes        string                 `json:"defaultNotes"`
		PricingTemplate     map[string]interface{} `json:"pricingTemplate"`
		RequiredDocuments   []string               `json:"requiredDocuments"`
		IsActive            bool                   `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to JSON
	equipmentJSON, _ := json.Marshal(request.EquipmentList)
	pricingJSON, _ := json.Marshal(request.PricingTemplate)
	documentsJSON, _ := json.Marshal(request.RequiredDocuments)

	// Update template
	template.Name = request.Name
	template.Description = request.Description
	template.JobCategoryID = &request.JobCategoryID
	template.DefaultDurationDays = request.DefaultDurationDays
	template.EquipmentList = equipmentJSON
	template.DefaultNotes = request.DefaultNotes
	template.PricingTemplate = pricingJSON
	template.RequiredDocuments = documentsJSON
	template.IsActive = request.IsActive
	template.UpdatedAt = time.Now()

	if err := h.db.Save(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job template updated successfully"})
}

// DeleteJobTemplate deletes a job template
func (h *WorkflowHandler) DeleteJobTemplate(c *gin.Context) {
	templateID := c.Param("id")

	result := h.db.Delete(&models.JobTemplate{}, templateID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete job template"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job template not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job template deleted successfully"})
}

// CreateJobFromTemplate creates a new job based on a template
func (h *WorkflowHandler) CreateJobFromTemplate(c *gin.Context) {
	templateID := c.Param("id")

	var template models.JobTemplate
	if err := h.db.First(&template, templateID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job template not found"})
		return
	}

	var request struct {
		CustomerID  uint       `json:"customerID" binding:"required"`
		StartDate   string     `json:"startDate"`
		EndDate     string     `json:"endDate"`
		Description string     `json:"description"`
		Notes       string     `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse dates
	var startDate, endDate *time.Time
	if request.StartDate != "" {
		if parsed, err := time.Parse("2006-01-02", request.StartDate); err == nil {
			startDate = &parsed
		}
	}
	if request.EndDate != "" {
		if parsed, err := time.Parse("2006-01-02", request.EndDate); err == nil {
			endDate = &parsed
		}
	}

	// Use template description if no custom description provided
	description := request.Description
	if description == "" {
		description = template.DefaultNotes
	}

	// Create job from template
	job := models.Job{
		CustomerID:    request.CustomerID,
		StatusID:      1, // Default to first status
		JobCategoryID: template.JobCategoryID,
		Description:   &description,
		StartDate:     startDate,
		EndDate:       endDate,
	}

	if err := h.db.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job from template"})
		return
	}

	// Increment template usage count
	h.db.Model(&template).Update("usage_count", gorm.Expr("usage_count + 1"))

	c.JSON(http.StatusCreated, gin.H{
		"message": "Job created from template successfully",
		"jobID":   job.JobID,
	})
}

// ================================================================
// EQUIPMENT PACKAGES
// ================================================================

// ListEquipmentPackages displays all equipment packages
func (h *WorkflowHandler) ListEquipmentPackages(c *gin.Context) {
	var packages []models.EquipmentPackage
	
	result := h.db.Preload("Creator").
		Order("usage_count DESC, name ASC").
		Find(&packages)
	
	if result.Error != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title": "Error",
			"error": "Failed to load equipment packages",
		})
		return
	}

	user, _ := GetCurrentUser(c)
	c.HTML(http.StatusOK, "equipment_packages_list.html", gin.H{
		"title":    "Equipment Packages",
		"user":     user,
		"packages": packages,
	})
}

// NewEquipmentPackageForm shows the form to create a new equipment package
func (h *WorkflowHandler) NewEquipmentPackageForm(c *gin.Context) {
	// Get available products for package creation
	var products []models.Product
	h.db.Find(&products)

	user, _ := GetCurrentUser(c)
	c.HTML(http.StatusOK, "equipment_package_form.html", gin.H{
		"title":    "New Equipment Package",
		"user":     user,
		"products": products,
		"isEdit":   false,
	})
}

// CreateEquipmentPackage creates a new equipment package
func (h *WorkflowHandler) CreateEquipmentPackage(c *gin.Context) {
	var request struct {
		Name            string                   `json:"name" binding:"required"`
		Description     string                   `json:"description"`
		PackageItems    []map[string]interface{} `json:"packageItems" binding:"required"`
		PackagePrice    float64                  `json:"packagePrice"`
		DiscountPercent float64                  `json:"discountPercent"`
		MinRentalDays   int                      `json:"minRentalDays"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Convert package items to JSON
	itemsJSON, _ := json.Marshal(request.PackageItems)

	pkg := models.EquipmentPackage{
		Name:            request.Name,
		Description:     request.Description,
		PackageItems:    itemsJSON,
		PackagePrice:    &request.PackagePrice,
		DiscountPercent: request.DiscountPercent,
		MinRentalDays:   request.MinRentalDays,
		IsActive:        true,
		CreatedBy:       &currentUser.UserID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		UsageCount:      0,
	}

	if err := h.db.Create(&pkg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create equipment package"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Equipment package created successfully",
		"packageID": pkg.PackageID,
	})
}

// GetEquipmentPackage retrieves a specific equipment package
func (h *WorkflowHandler) GetEquipmentPackage(c *gin.Context) {
	packageID := c.Param("id")

	var pkg models.EquipmentPackage
	result := h.db.Preload("Creator").First(&pkg, packageID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"title": "Package Not Found",
				"error": "Equipment package not found",
			})
		} else {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"title": "Error",
				"error": "Failed to load equipment package",
			})
		}
		return
	}

	user, _ := GetCurrentUser(c)
	c.HTML(http.StatusOK, "equipment_package_detail.html", gin.H{
		"title":   "Equipment Package: " + pkg.Name,
		"user":    user,
		"package": pkg,
	})
}

// ================================================================
// BULK OPERATIONS
// ================================================================

// BulkOperationsForm shows the bulk operations interface
func (h *WorkflowHandler) BulkOperationsForm(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	c.HTML(http.StatusOK, "bulk_operations.html", gin.H{
		"title": "Bulk Operations",
		"user":  user,
	})
}

// BulkUpdateDeviceStatus updates status for multiple devices
func (h *WorkflowHandler) BulkUpdateDeviceStatus(c *gin.Context) {
	var request struct {
		DeviceIDs []string `json:"deviceIDs" binding:"required"`
		NewStatus string   `json:"newStatus" binding:"required"`
		Notes     string   `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate device IDs exist
	var count int64
	h.db.Model(&models.Device{}).Where("deviceID IN ?", request.DeviceIDs).Count(&count)
	
	if count != int64(len(request.DeviceIDs)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Some device IDs are invalid"})
		return
	}

	// Update device statuses
	result := h.db.Model(&models.Device{}).
		Where("deviceID IN ?", request.DeviceIDs).
		Update("status", request.NewStatus)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update device statuses"})
		return
	}

	// Log the bulk operation
	for _, deviceID := range request.DeviceIDs {
		log := models.EquipmentUsageLog{
			DeviceID:  deviceID,
			Action:    request.NewStatus,
			Timestamp: time.Now(),
			Notes:     request.Notes,
			CreatedAt: time.Now(),
		}
		h.db.Create(&log)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Device statuses updated successfully",
		"devicesUpdated": len(request.DeviceIDs),
	})
}

// BulkAssignToJob assigns multiple devices to a job
func (h *WorkflowHandler) BulkAssignToJob(c *gin.Context) {
	var request struct {
		DeviceIDs []string `json:"deviceIDs" binding:"required"`
		JobID     uint     `json:"jobID" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate job exists
	var job models.Job
	if err := h.db.First(&job, request.JobID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	// Validate devices exist and are available
	var availableDevices []models.Device
	h.db.Where("deviceID IN ? AND status = 'free'", request.DeviceIDs).Find(&availableDevices)

	if len(availableDevices) != len(request.DeviceIDs) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Some devices are not available for assignment"})
		return
	}

	// Create job-device assignments
	successCount := 0
	for _, deviceID := range request.DeviceIDs {
		// Check if already assigned
		var existing models.JobDevice
		result := h.db.Where("jobID = ? AND deviceID = ?", request.JobID, deviceID).First(&existing)
		
		if result.Error == gorm.ErrRecordNotFound {
			// Create new assignment
			assignment := models.JobDevice{
				JobID:    request.JobID,
				DeviceID: deviceID,
			}
			if err := h.db.Create(&assignment).Error; err == nil {
				// Update device status
				h.db.Model(&models.Device{}).Where("deviceID = ?", deviceID).Update("status", "assigned")
				successCount++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Devices assigned to job successfully",
		"devicesAssigned": successCount,
	})
}

// BulkGenerateQRCodes generates QR codes for multiple devices
func (h *WorkflowHandler) BulkGenerateQRCodes(c *gin.Context) {
	var request struct {
		DeviceIDs []string `json:"deviceIDs" binding:"required"`
		Format    string   `json:"format"` // pdf, zip
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate devices exist
	var devices []models.Device
	result := h.db.Where("deviceID IN ?", request.DeviceIDs).Find(&devices)
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve devices"})
		return
	}

	if len(devices) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No devices found"})
		return
	}

	// For now, return the device data - QR generation would be implemented separately
	c.JSON(http.StatusOK, gin.H{
		"message": "QR codes generation initiated",
		"devices": devices,
		"count":   len(devices),
	})
}

// GetWorkflowStats returns workflow statistics
func (h *WorkflowHandler) GetWorkflowStats(c *gin.Context) {
	var stats struct {
		TotalTemplates    int64 `json:"totalTemplates"`
		ActiveTemplates   int64 `json:"activeTemplates"`
		TotalPackages     int64 `json:"totalPackages"`
		ActivePackages    int64 `json:"activePackages"`
		TemplateUsage     int64 `json:"templateUsage"`
		PackageUsage      int64 `json:"packageUsage"`
		AvailableDevices  int64 `json:"availableDevices"`
		AssignedDevices   int64 `json:"assignedDevices"`
	}

	// Get template stats
	h.db.Model(&models.JobTemplate{}).Count(&stats.TotalTemplates)
	h.db.Model(&models.JobTemplate{}).Where("is_active = ?", true).Count(&stats.ActiveTemplates)
	h.db.Model(&models.JobTemplate{}).Select("COALESCE(SUM(usage_count), 0)").Scan(&stats.TemplateUsage)

	// Get package stats
	h.db.Model(&models.EquipmentPackage{}).Count(&stats.TotalPackages)
	h.db.Model(&models.EquipmentPackage{}).Where("is_active = ?", true).Count(&stats.ActivePackages)
	h.db.Model(&models.EquipmentPackage{}).Select("COALESCE(SUM(usage_count), 0)").Scan(&stats.PackageUsage)

	// Get device stats
	h.db.Model(&models.Device{}).Where("status = ?", "free").Count(&stats.AvailableDevices)
	h.db.Model(&models.Device{}).Where("status = ?", "assigned").Count(&stats.AssignedDevices)

	c.JSON(http.StatusOK, stats)
}

// API endpoints for workflow management

// ListJobTemplatesAPI returns job templates as JSON
func (h *WorkflowHandler) ListJobTemplatesAPI(c *gin.Context) {
	var templates []models.JobTemplate
	
	result := h.db.Preload("JobCategory").Preload("Creator").
		Order("usage_count DESC, name ASC").
		Find(&templates)
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load job templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"count":     len(templates),
	})
}

// ListEquipmentPackagesAPI returns equipment packages as JSON
func (h *WorkflowHandler) ListEquipmentPackagesAPI(c *gin.Context) {
	var packages []models.EquipmentPackage
	
	result := h.db.Preload("Creator").
		Order("usage_count DESC, name ASC").
		Find(&packages)
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load equipment packages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"packages": packages,
		"count":    len(packages),
	})
}