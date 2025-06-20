package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/repository"
	"go-barcode-webapp/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"gorm.io/gorm"
)

type WorkflowHandler struct {
	templateRepo   *repository.JobTemplateRepository
	jobRepo        *repository.JobRepository
	customerRepo   *repository.CustomerRepository
	packageRepo    *repository.EquipmentPackageRepository
	deviceRepo     *repository.DeviceRepository
	db             *gorm.DB
	barcodeService *services.BarcodeService
}

func NewWorkflowHandler(templateRepo *repository.JobTemplateRepository, jobRepo *repository.JobRepository, customerRepo *repository.CustomerRepository, packageRepo *repository.EquipmentPackageRepository, deviceRepo *repository.DeviceRepository, db *gorm.DB, barcodeService *services.BarcodeService) *WorkflowHandler {
	return &WorkflowHandler{
		templateRepo:   templateRepo,
		jobRepo:        jobRepo,
		customerRepo:   customerRepo,
		packageRepo:    packageRepo,
		deviceRepo:     deviceRepo,
		db:             db,
		barcodeService: barcodeService,
	}
}

// ================================================================
// JOB TEMPLATES - WEB HANDLERS
// ================================================================

// ListJobTemplates displays all job templates
func (h *WorkflowHandler) ListJobTemplates(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	params := &models.FilterParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error(), "user": user})
		return
	}

	templates, err := h.templateRepo.List(params)
	if err != nil {
		log.Printf("ListJobTemplates: Error fetching templates: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load job templates", "user": user})
		return
	}

	// Get total count for pagination
	totalCount, err := h.templateRepo.Count(params)
	if err != nil {
		log.Printf("ListJobTemplates: Error getting count: %v", err)
		totalCount = 0
	}

	c.HTML(http.StatusOK, "job_templates_list.html", gin.H{
		"title":      "Job Templates",
		"templates":  templates,
		"params":     params,
		"totalCount": totalCount,
		"user":       user,
	})
}

// NewJobTemplateForm displays the form for creating a new job template
func (h *WorkflowHandler) NewJobTemplateForm(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	c.HTML(http.StatusOK, "job_template_form.html", gin.H{
		"title":    "New Job Template",
		"template": &models.JobTemplate{},
		"isEdit":   false,
		"user":     user,
	})
}

// CreateJobTemplate creates a new job template
func (h *WorkflowHandler) CreateJobTemplate(c *gin.Context) {
	// Get current user
	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.HTML(http.StatusUnauthorized, "job_template_form.html", gin.H{
			"title":    "New Job Template",
			"template": &models.JobTemplate{},
			"isEdit":   false,
			"error":    "Authentication required. Please log in to continue.",
			"user":     currentUser,
		})
		return
	}

	// Parse form data
	name := c.PostForm("name")
	description := c.PostForm("description")
	jobCategoryIDStr := c.PostForm("jobCategoryID")
	defaultDurationDaysStr := c.PostForm("defaultDurationDays")
	isActiveStr := c.PostForm("isActive")
	defaultNotes := c.PostForm("defaultNotes")
	equipmentList := c.PostForm("equipmentList")
	pricingTemplate := c.PostForm("pricingTemplate")
	requiredDocuments := c.PostForm("requiredDocuments")

	// Validate required fields
	if name == "" {
		c.HTML(http.StatusBadRequest, "job_template_form.html", gin.H{
			"title":    "New Job Template",
			"template": &models.JobTemplate{},
			"isEdit":   false,
			"error":    "Template name is required",
			"user":     currentUser,
		})
		return
	}

	// Parse optional fields
	var jobCategoryID *uint
	if jobCategoryIDStr != "" {
		if id, err := strconv.ParseUint(jobCategoryIDStr, 10, 32); err == nil {
			catID := uint(id)
			jobCategoryID = &catID
		}
	}

	var defaultDurationDays *int
	if defaultDurationDaysStr != "" {
		if days, err := strconv.Atoi(defaultDurationDaysStr); err == nil && days > 0 {
			defaultDurationDays = &days
		}
	}

	isActive := isActiveStr != "false" // Default to true

	// Build template
	template := models.JobTemplate{
		Name:                name,
		Description:         description,
		JobCategoryID:       jobCategoryID,
		DefaultDurationDays: defaultDurationDays,
		DefaultNotes:        defaultNotes,
		IsActive:            isActive,
		CreatedBy:           &currentUser.UserID,
	}

	// Handle JSON fields
	if equipmentList != "" && equipmentList != "[]" {
		template.EquipmentList = []byte(equipmentList)
		if err := h.templateRepo.ValidateEquipmentList(template.EquipmentList); err != nil {
			c.HTML(http.StatusBadRequest, "job_template_form.html", gin.H{
				"title":    "New Job Template",
				"template": &template,
				"isEdit":   false,
				"error":    "Invalid equipment list: " + err.Error(),
				"user":     currentUser,
			})
			return
		}
	}

	if pricingTemplate != "" && pricingTemplate != "{}" {
		template.PricingTemplate = []byte(pricingTemplate)
		if err := h.templateRepo.ValidatePricingTemplate(template.PricingTemplate); err != nil {
			c.HTML(http.StatusBadRequest, "job_template_form.html", gin.H{
				"title":    "New Job Template",
				"template": &template,
				"isEdit":   false,
				"error":    "Invalid pricing template: " + err.Error(),
				"user":     currentUser,
			})
			return
		}
	}

	if requiredDocuments != "" && requiredDocuments != "[]" {
		template.RequiredDocuments = []byte(requiredDocuments)
	}

	// Create template
	if err := h.templateRepo.Create(&template); err != nil {
		log.Printf("CreateJobTemplate: Database error: %v", err)
		c.HTML(http.StatusInternalServerError, "job_template_form.html", gin.H{
			"title":    "New Job Template",
			"template": &template,
			"isEdit":   false,
			"error":    "Failed to create job template: " + err.Error(),
			"user":     currentUser,
		})
		return
	}

	// Redirect to template detail page
	c.Redirect(http.StatusFound, "/workflow/templates/"+strconv.Itoa(int(template.TemplateID)))
}

// GetJobTemplate returns a specific job template
func (h *WorkflowHandler) GetJobTemplate(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid template ID", "user": user})
		return
	}

	template, err := h.templateRepo.GetByID(uint(templateID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Job template not found", "user": user})
		return
	}

	c.HTML(http.StatusOK, "job_template_detail.html", gin.H{
		"title":    "Job Template Details",
		"template": template,
		"user":     user,
	})
}

// EditJobTemplateForm displays the form for editing a job template
func (h *WorkflowHandler) EditJobTemplateForm(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid template ID", "user": user})
		return
	}

	template, err := h.templateRepo.GetByID(uint(templateID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Job template not found", "user": user})
		return
	}

	c.HTML(http.StatusOK, "job_template_form.html", gin.H{
		"title":    "Edit Job Template",
		"template": template,
		"isEdit":   true,
		"user":     user,
	})
}

// UpdateJobTemplate updates an existing job template
func (h *WorkflowHandler) UpdateJobTemplate(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	// Get existing template
	existingTemplate, err := h.templateRepo.GetByID(uint(templateID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job template not found"})
		return
	}

	// Check authorization - only creator or admin can edit
	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	if existingTemplate.CreatedBy != nil && *existingTemplate.CreatedBy != currentUser.UserID {
		// TODO: Add admin role check here
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit templates you created"})
		return
	}

	// Bind updated data
	var updatedTemplate models.JobTemplate
	if err := c.ShouldBindJSON(&updatedTemplate); err != nil {
		log.Printf("UpdateJobTemplate: Validation error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input data",
			"details": err.Error(),
		})
		return
	}

	// Preserve original fields
	updatedTemplate.TemplateID = existingTemplate.TemplateID
	updatedTemplate.CreatedBy = existingTemplate.CreatedBy
	updatedTemplate.CreatedAt = existingTemplate.CreatedAt
	updatedTemplate.UsageCount = existingTemplate.UsageCount

	// Validate JSON fields
	if err := h.templateRepo.ValidateEquipmentList(updatedTemplate.EquipmentList); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid equipment list",
			"details": err.Error(),
		})
		return
	}

	if err := h.templateRepo.ValidatePricingTemplate(updatedTemplate.PricingTemplate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid pricing template",
			"details": err.Error(),
		})
		return
	}

	// Update template
	if err := h.templateRepo.Update(&updatedTemplate); err != nil {
		log.Printf("UpdateJobTemplate: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Job template updated successfully",
		"template": updatedTemplate,
	})
}

// DeleteJobTemplate soft deletes a job template
func (h *WorkflowHandler) DeleteJobTemplate(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	// Get existing template for authorization check
	existingTemplate, err := h.templateRepo.GetByID(uint(templateID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job template not found"})
		return
	}

	// Check authorization
	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	if existingTemplate.CreatedBy != nil && *existingTemplate.CreatedBy != currentUser.UserID {
		// TODO: Add admin role check here
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete templates you created"})
		return
	}

	// Soft delete
	if err := h.templateRepo.Delete(uint(templateID)); err != nil {
		log.Printf("DeleteJobTemplate: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete job template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job template deleted successfully"})
}

// CreateJobFromTemplate creates a new job using a template
func (h *WorkflowHandler) CreateJobFromTemplate(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var request struct {
		CustomerID  uint       `json:"customerID" binding:"required"`
		StartDate   *time.Time `json:"startDate"`
		EndDate     *time.Time `json:"endDate"`
		Description string     `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input data",
			"details": err.Error(),
		})
		return
	}

	// Validate customer exists
	_, err = h.customerRepo.GetByID(request.CustomerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	// Get template
	template, err := h.templateRepo.GetByID(uint(templateID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job template not found"})
		return
	}

	// Calculate dates if not provided
	if request.StartDate == nil {
		now := time.Now()
		request.StartDate = &now
	}

	if request.EndDate == nil && template.DefaultDurationDays != nil {
		endDate := request.StartDate.AddDate(0, 0, *template.DefaultDurationDays)
		request.EndDate = &endDate
	}

	// Get default status ID
	defaultStatusID := uint(1) // TODO: Make this configurable
	if getDefaultStatusID() != nil {
		defaultStatusID = *getDefaultStatusID()
	}

	// Create job from template
	job := models.Job{
		CustomerID:  request.CustomerID,
		StartDate:   request.StartDate,
		EndDate:     request.EndDate,
		StatusID:    defaultStatusID,
		Description: &request.Description,
		TemplateID:  &template.TemplateID,
	}

	// Set description from template if not provided
	if request.Description == "" && template.DefaultNotes != "" {
		job.Description = &template.DefaultNotes
	}

	// Create the job
	if err := h.jobRepo.Create(&job); err != nil {
		log.Printf("CreateJobFromTemplate: Failed to create job: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job from template"})
		return
	}

	// Increment template usage count
	if err := h.templateRepo.IncrementUsageCount(template.TemplateID); err != nil {
		log.Printf("CreateJobFromTemplate: Failed to increment usage count: %v", err)
		// Don't fail the request, just log the error
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Job created successfully from template",
		"jobID":   job.JobID,
	})
}

// ================================================================
// API HANDLERS
// ================================================================

// ListJobTemplatesAPI returns job templates as JSON
func (h *WorkflowHandler) ListJobTemplatesAPI(c *gin.Context) {
	params := &models.FilterParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	templates, err := h.templateRepo.List(params)
	if err != nil {
		log.Printf("ListJobTemplatesAPI: Error fetching templates: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load job templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// GetMostUsedTemplatesAPI returns most used templates
func (h *WorkflowHandler) GetMostUsedTemplatesAPI(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	templates, err := h.templateRepo.GetMostUsed(limit)
	if err != nil {
		log.Printf("GetMostUsedTemplatesAPI: Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// GetTemplatesByCategoryAPI returns templates for a specific category
func (h *WorkflowHandler) GetTemplatesByCategoryAPI(c *gin.Context) {
	categoryIDStr := c.Param("categoryId")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	templates, err := h.templateRepo.GetByCategory(uint(categoryID))
	if err != nil {
		log.Printf("GetTemplatesByCategoryAPI: Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// ================================================================
// HELPER FUNCTIONS
// ================================================================

// getDefaultStatusID returns the default status ID for new jobs
// TODO: This should be configurable or fetched from database
func getDefaultStatusID() *uint {
	// For now, return nil to let the job creation handle the default
	// This prevents the hardcoded status ID issue
	return nil
}

// ================================================================
// EQUIPMENT PACKAGES - PLACEHOLDER METHODS
// ================================================================

// ListEquipmentPackages displays all equipment packages
func (h *WorkflowHandler) ListEquipmentPackages(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	params := &models.FilterParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error(), "user": user})
		return
	}

	packages, err := h.packageRepo.List(params)
	if err != nil {
		log.Printf("ListEquipmentPackages: Error fetching packages: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load equipment packages", "user": user})
		return
	}

	// Get total count for pagination
	totalCount, err := h.packageRepo.GetTotalCount(params)
	if err != nil {
		log.Printf("ListEquipmentPackages: Error getting total count: %v", err)
		totalCount = 0
	}

	c.HTML(http.StatusOK, "equipment_packages_list.html", gin.H{
		"title":      "Equipment Packages",
		"packages":   packages,
		"totalCount": totalCount,
		"user":       user,
	})
}

// NewEquipmentPackageForm displays the form for creating a new equipment package
func (h *WorkflowHandler) NewEquipmentPackageForm(c *gin.Context) {
	// Get current user for base template
	currentUser, exists := GetCurrentUser(c)
	if !exists {
		log.Printf("NewEquipmentPackageForm: User not authenticated")
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Get available devices for the dropdown
	availableDevices, err := h.packageRepo.GetAvailableDevices()
	if err != nil {
		log.Printf("NewEquipmentPackageForm: Error fetching available devices: %v", err)
		availableDevices = []models.Device{} // Use empty slice if error
	}
	
	log.Printf("NewEquipmentPackageForm: Found %d available devices", len(availableDevices))
	if len(availableDevices) > 0 {
		log.Printf("NewEquipmentPackageForm: Sample device: ID=%s, Product=%v", 
			availableDevices[0].DeviceID, 
			func() string { if availableDevices[0].Product != nil { return availableDevices[0].Product.Name } else { return "nil" } }())
	}

	c.HTML(http.StatusOK, "equipment_package_form.html", gin.H{
		"title":            "New Equipment Package",
		"package":          nil,
		"isEdit":           false,
		"user":             currentUser,
		"availableDevices": availableDevices,
	})
}

// CreateEquipmentPackage creates a new equipment package
func (h *WorkflowHandler) CreateEquipmentPackage(c *gin.Context) {
	// Get current user
	currentUser, exists := GetCurrentUser(c)
	if !exists {
		availableDevices, _ := h.packageRepo.GetAvailableDevices()
		c.HTML(http.StatusUnauthorized, "equipment_package_form.html", gin.H{
			"title":            "New Equipment Package",
			"package":          &models.EquipmentPackage{},
			"isEdit":           false,
			"error":            "Authentication required. Please log in to continue.",
			"user":             currentUser,
			"availableDevices": availableDevices,
		})
		return
	}

	// Parse form data
	var pkg models.EquipmentPackage
	pkg.Name = c.PostForm("name")
	pkg.Description = c.PostForm("description")
	
	// Parse optional fields
	if packagePrice := c.PostForm("packagePrice"); packagePrice != "" {
		if price, err := strconv.ParseFloat(packagePrice, 64); err == nil {
			pkg.PackagePrice = &price
		}
	}
	
	if discountPercent := c.PostForm("discountPercent"); discountPercent != "" {
		if discount, err := strconv.ParseFloat(discountPercent, 64); err == nil {
			pkg.DiscountPercent = discount
		}
	}
	
	if minRentalDays := c.PostForm("minRentalDays"); minRentalDays != "" {
		if days, err := strconv.Atoi(minRentalDays); err == nil {
			pkg.MinRentalDays = days
		}
	}
	
	pkg.IsActive = c.PostForm("isActive") == "on"
	
	// Handle package items
	packageItemsStr := c.PostForm("packageItems")
	if packageItemsStr == "" {
		packageItemsStr = "[]"
	}
	pkg.PackageItems = json.RawMessage(packageItemsStr)

	// Parse device selections
	var deviceMappings []models.PackageDevice
	
	// Parse devices array from form
	i := 0
	for {
		deviceIDKey := "devices[" + strconv.Itoa(i) + "][deviceID]"
		deviceID := c.PostForm(deviceIDKey)
		
		if deviceID == "" {
			break // No more devices
		}
		
		// Parse device data
		quantityKey := "devices[" + strconv.Itoa(i) + "][quantity]"
		customPriceKey := "devices[" + strconv.Itoa(i) + "][customPrice]"
		isRequiredKey := "devices[" + strconv.Itoa(i) + "][isRequired]"
		notesKey := "devices[" + strconv.Itoa(i) + "][notes]"
		
		quantity, _ := strconv.ParseUint(c.PostForm(quantityKey), 10, 32)
		if quantity == 0 {
			quantity = 1
		}
		
		var customPrice *float64
		if customPriceStr := c.PostForm(customPriceKey); customPriceStr != "" {
			if price, err := strconv.ParseFloat(customPriceStr, 64); err == nil {
				customPrice = &price
			}
		}
		
		isRequired := c.PostForm(isRequiredKey) == "true"
		notes := c.PostForm(notesKey)
		
		deviceMapping := models.PackageDevice{
			DeviceID:    deviceID,
			Quantity:    uint(quantity),
			CustomPrice: customPrice,
			IsRequired:  isRequired,
			Notes:       notes,
		}
		
		deviceMappings = append(deviceMappings, deviceMapping)
		i++
	}

	// Validate required fields
	if pkg.Name == "" {
		availableDevices, _ := h.packageRepo.GetAvailableDevices()
		c.HTML(http.StatusBadRequest, "equipment_package_form.html", gin.H{
			"title":            "New Equipment Package",
			"package":          &pkg,
			"isEdit":           false,
			"error":            "Package name is required",
			"user":             currentUser,
			"availableDevices": availableDevices,
		})
		return
	}

	// Set creator
	pkg.CreatedBy = &currentUser.UserID
	
	// Save to database with device associations
	if err := h.packageRepo.CreateWithDevices(&pkg, deviceMappings); err != nil {
		log.Printf("CreateEquipmentPackage: Database error: %v", err)
		availableDevices, _ := h.packageRepo.GetAvailableDevices()
		c.HTML(http.StatusInternalServerError, "equipment_package_form.html", gin.H{
			"title":            "New Equipment Package",
			"package":          &pkg,
			"isEdit":           false,
			"error":            "Failed to create equipment package: " + err.Error(),
			"user":             currentUser,
			"availableDevices": availableDevices,
		})
		return
	}
	
	log.Printf("CreateEquipmentPackage: Successfully created package '%s' (ID: %d) with %d devices by user %s", 
		pkg.Name, pkg.PackageID, len(deviceMappings), currentUser.Username)
	
	// Redirect to packages list on success
	c.Redirect(http.StatusSeeOther, "/workflow/packages")
}

// GetEquipmentPackage returns a specific equipment package
func (h *WorkflowHandler) GetEquipmentPackage(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	// TODO: Implement equipment package details
	packageID := c.Param("id")
	log.Printf("GetEquipmentPackage: Package ID %s not yet implemented", packageID)
	c.HTML(http.StatusOK, "equipment_package_detail.html", gin.H{
		"title":   "Equipment Package Details",
		"package": map[string]interface{}{"id": packageID},
		"user":    user,
	})
}

// DebugPackageForm shows debug info for package form
func (h *WorkflowHandler) DebugPackageForm(c *gin.Context) {
	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Get available devices for debugging
	availableDevices, err := h.packageRepo.GetAvailableDevices()
	if err != nil {
		log.Printf("DebugPackageForm: Error fetching available devices: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "equipment_package_form_debug.html", gin.H{
		"title":            "Debug Package Form",
		"user":             currentUser,
		"availableDevices": availableDevices,
	})
}

// ================================================================
// BULK OPERATIONS - PLACEHOLDER METHODS
// ================================================================

// BulkOperationsForm displays the bulk operations interface
func (h *WorkflowHandler) BulkOperationsForm(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	// TODO: Implement bulk operations interface
	c.HTML(http.StatusOK, "bulk_operations.html", gin.H{
		"title": "Bulk Operations",
		"user":  user,
	})
}

// BulkUpdateDeviceStatus updates multiple device statuses
func (h *WorkflowHandler) BulkUpdateDeviceStatus(c *gin.Context) {
	// TODO: Implement bulk device status update
	log.Printf("BulkUpdateDeviceStatus: Not yet implemented")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Bulk device status update not yet implemented",
	})
}

// BulkAssignToJob assigns multiple devices to a job
func (h *WorkflowHandler) BulkAssignToJob(c *gin.Context) {
	// TODO: Implement bulk device assignment
	log.Printf("BulkAssignToJob: Not yet implemented")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Bulk device assignment not yet implemented",
	})
}

// BulkGenerateQRCodes generates QR codes for multiple devices
func (h *WorkflowHandler) BulkGenerateQRCodes(c *gin.Context) {
	// Parse request
	var request struct {
		DeviceIDs    []string `json:"deviceIds" form:"deviceIds"`
		Format       string   `json:"format" form:"format"`       // "pdf" or "zip"
		LabelFormat  string   `json:"labelFormat" form:"labelFormat"` // "simple" or "detailed"
		PrintReady   bool     `json:"printReady" form:"printReady"`
	}

	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate device IDs
	if len(request.DeviceIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No device IDs provided"})
		return
	}

	// Default values
	if request.Format == "" {
		request.Format = "pdf"
	}
	if request.LabelFormat == "" {
		request.LabelFormat = "simple"
	}

	log.Printf("Generating QR codes for %d devices, format: %s", len(request.DeviceIDs), request.Format)

	// Fetch device information
	devices := make([]models.Device, 0, len(request.DeviceIDs))
	for _, deviceID := range request.DeviceIDs {
		var device models.Device
		if err := h.db.Preload("Product").Preload("Product.Brand").Where("deviceID = ?", deviceID).First(&device).Error; err != nil {
			log.Printf("Warning: Device %s not found in database, will generate QR anyway", deviceID)
			// Create a minimal device record for QR generation
			device = models.Device{
				DeviceID: deviceID,
				Status:   "unknown",
				Product:  nil, // Will be handled in template
			}
		}
		devices = append(devices, device)
	}

	if request.Format == "zip" {
		// Generate PNG files and create ZIP
		zipBytes, err := h.generateDeviceLabelsZIP(devices, request.LabelFormat, request.PrintReady)
		if err != nil {
			log.Printf("Error generating device labels ZIP: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate device labels ZIP"})
			return
		}

		// Set headers for ZIP download
		filename := fmt.Sprintf("device_labels_%s.zip", time.Now().Format("20060102_150405"))
		c.Header("Content-Type", "application/zip")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Header("Content-Length", fmt.Sprintf("%d", len(zipBytes)))

		// Return ZIP
		c.Data(http.StatusOK, "application/zip", zipBytes)
	} else {
		// Generate PDF with multiple labels per page
		pdfBytes, err := h.generateDeviceLabelsPDF(devices, request.LabelFormat, request.PrintReady)
		if err != nil {
			log.Printf("Error generating device labels PDF: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate device labels PDF"})
			return
		}

		// Set headers for PDF download
		filename := fmt.Sprintf("device_labels_%s.pdf", time.Now().Format("20060102_150405"))
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

		// Return PDF
		c.Data(http.StatusOK, "application/pdf", pdfBytes)
	}
}

// generateDeviceLabelsPDF creates a PDF with multiple device labels per page
func (h *WorkflowHandler) generateDeviceLabelsPDF(devices []models.Device, labelFormat string, printReady bool) ([]byte, error) {
	// Create PDF document - A4 Portrait for multiple labels
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	
	// Load logo if exists
	logoPath := "logo.png"
	logoExists := false
	if _, err := os.Stat(logoPath); err == nil {
		logoExists = true
	}
	
	// Label dimensions - 3x7 grid on A4 (21 labels per page)
	labelWidth := 60.0
	labelHeight := 35.0
	labelsPerRow := 3
	labelsPerCol := 7
	labelsPerPage := labelsPerRow * labelsPerCol
	
	// Process devices in batches per page
	for pageStart := 0; pageStart < len(devices); pageStart += labelsPerPage {
		pdf.AddPage()
		
		// Draw labels for this page
		for i := 0; i < labelsPerPage && pageStart+i < len(devices); i++ {
			device := devices[pageStart+i]
			
			// Calculate position for this label
			row := i / labelsPerRow
			col := i % labelsPerRow
			
			offsetX := 10.0 + float64(col)*labelWidth
			offsetY := 10.0 + float64(row)*labelHeight
			
			h.drawSingleLabel(pdf, device, offsetX, offsetY, labelWidth, labelHeight, logoExists, logoPath)
		}
	}
	
	// Output PDF to bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}
	
	return buf.Bytes(), nil
}

// drawSingleLabel draws a single device label at the specified position
func (h *WorkflowHandler) drawSingleLabel(pdf *gofpdf.Fpdf, device models.Device, offsetX, offsetY, width, height float64, logoExists bool, logoPath string) {
	// Get product name
	productName := "Unknown Product"
	if device.Product != nil {
		productName = device.Product.Name
	}
	
	// Draw border around label (optional)
	pdf.SetDrawColor(200, 200, 200)
	pdf.Rect(offsetX, offsetY, width, height, "D")
	
	// 1. Logo at right side, vertically centered (if exists)
	if logoExists {
		logoX := offsetX + width - 20
		logoY := offsetY + (height-8)/2
		pdf.Image(logoPath, logoX, logoY, 15, 8, false, "", 0, "")
	}
	
	// Remove the title - start barcode higher up
	
	// 3. Main barcode in center area (moved up since no title)
	barcodeX := offsetX + 2
	barcodeY := offsetY + 4
	barcodeWidth := width - 25 // Leave space for logo
	barcodeHeight := 8.0
	
	// Generate realistic Code128 barcode pattern
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetFillColor(0, 0, 0)
	
	// Use device ID for barcode data
	deviceData := device.DeviceID
	totalBars := len(deviceData) * 8 + 20
	barWidth := barcodeWidth / float64(totalBars)
	
	x := barcodeX
	
	// Start pattern
	for i := 0; i < 3; i++ {
		pdf.Rect(x, barcodeY, barWidth, barcodeHeight, "F")
		x += barWidth * 2
	}
	
	// Data encoding
	for i, char := range deviceData {
		charVal := int(char) + i
		for j := 0; j < 6; j++ {
			if (charVal+j)%3 != 0 {
				pdf.Rect(x, barcodeY, barWidth, barcodeHeight, "F")
			}
			x += barWidth
		}
		x += barWidth
	}
	
	// End pattern
	for i := 0; i < 3; i++ {
		pdf.Rect(x, barcodeY, barWidth, barcodeHeight, "F")
		x += barWidth * 2
	}
	
	// 4. Human readable text under barcode
	pdf.SetXY(barcodeX, barcodeY + barcodeHeight + 1)
	pdf.SetFont("Arial", "", 5)
	pdf.CellFormat(barcodeWidth, 2, device.DeviceID, "", 0, "C", false, 0, "")
	
	// 5. Device information at bottom
	pdf.SetXY(offsetX+2, offsetY+height-10)
	pdf.SetFont("Arial", "B", 7)
	pdf.Cell(0, 3, device.DeviceID)
	
	pdf.SetXY(offsetX+2, offsetY+height-7)
	pdf.SetFont("Arial", "", 6)
	// Truncate product name if too long
	if len(productName) > 25 {
		productName = productName[:22] + "..."
	}
	pdf.Cell(0, 3, productName)
}

// generateDeviceLabelsZIP creates complete label PNG files for each device and packages them in a ZIP
func (h *WorkflowHandler) generateDeviceLabelsZIP(devices []models.Device, labelFormat string, printReady bool) ([]byte, error) {
	// Create ZIP file in memory
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	defer zipWriter.Close()
	
	// Load logo image if exists
	var logoImg image.Image
	logoPath := "logo.png"
	if logoFile, err := os.Open(logoPath); err == nil {
		logoImg, _, _ = image.Decode(logoFile)
		logoFile.Close()
	}
	
	// Create complete label PNG for each device
	for _, device := range devices {
		// Create PNG image for this device
		pngBytes, err := h.createLabelPNG(device, logoImg)
		if err != nil {
			log.Printf("Error generating PNG for device %s: %v", device.DeviceID, err)
			continue
		}
		
		// Create PNG filename
		filename := fmt.Sprintf("label_%s.png", device.DeviceID)
		
		zipFile, err := zipWriter.Create(filename)
		if err != nil {
			log.Printf("Error creating zip file for device %s: %v", device.DeviceID, err)
			continue
		}
		
		_, err = zipFile.Write(pngBytes)
		if err != nil {
			log.Printf("Error writing to zip file for device %s: %v", device.DeviceID, err)
			continue
		}
	}
	
	err := zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close ZIP writer: %v", err)
	}
	
	return buf.Bytes(), nil
}

// createLabelPNG creates a complete label as PNG image
func (h *WorkflowHandler) createLabelPNG(device models.Device, logoImg image.Image) ([]byte, error) {
	// Label dimensions in pixels (300 DPI equivalent for 100x60mm)
	width := 1200
	height := 700
	
	// Create a new RGBA image
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Fill with white background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)
	
	// Draw border
	borderColor := color.RGBA{200, 200, 200, 255}
	h.drawRect(img, 10, 10, width-20, height-20, borderColor)
	
	// Generate barcode image
	barcodeBytes, err := h.barcodeService.GenerateDeviceBarcode(device.DeviceID)
	if err == nil {
		if barcodeImg, _, err := image.Decode(bytes.NewReader(barcodeBytes)); err == nil {
			// Scale and position barcode
			barcodeRect := image.Rect(50, 150, 800, 350)
			xdraw.BiLinear.Scale(img, barcodeRect, barcodeImg, barcodeImg.Bounds(), draw.Over, nil)
		}
	}
	
	// Draw logo if available
	if logoImg != nil {
		logoRect := image.Rect(900, 200, 1100, 300)
		xdraw.BiLinear.Scale(img, logoRect, logoImg, logoImg.Bounds(), draw.Over, nil)
	}
	
	// Get product name
	productName := "Unknown Product"
	if device.Product != nil {
		productName = device.Product.Name
		if len(productName) > 30 {
			productName = productName[:27] + "..."
		}
	}
	
	// Draw text
	textColor := color.RGBA{0, 0, 0, 255}
	
	// Device ID (large, bold)
	h.drawText(img, device.DeviceID, 50, 450, 48, textColor)
	
	// Product name (smaller)
	h.drawText(img, productName, 50, 520, 32, textColor)
	
	// Device ID under barcode (small)
	h.drawText(img, device.DeviceID, 350, 380, 24, textColor)
	
	// Convert to PNG bytes
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %v", err)
	}
	
	return buf.Bytes(), nil
}

// drawRect draws a rectangle outline
func (h *WorkflowHandler) drawRect(img *image.RGBA, x, y, width, height int, c color.RGBA) {
	for i := 0; i < width; i++ {
		img.Set(x+i, y, c)
		img.Set(x+i, y+height-1, c)
	}
	for i := 0; i < height; i++ {
		img.Set(x, y+i, c)
		img.Set(x+width-1, y+i, c)
	}
}

// drawText draws text on the image
func (h *WorkflowHandler) drawText(img *image.RGBA, text string, x, y, size int, c color.RGBA) {
	point := fixed.Point26_6{
		X: fixed.Int26_6(x * 64),
		Y: fixed.Int26_6(y * 64),
	}
	
	d := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{c},
		Face: basicfont.Face7x13, // Simple built-in font
		Dot:  point,
	}
	
	d.DrawString(text)
}


// ================================================================
// WORKFLOW STATISTICS - PLACEHOLDER METHOD
// ================================================================

// GetWorkflowStats returns workflow statistics
func (h *WorkflowHandler) GetWorkflowStats(c *gin.Context) {
	// TODO: Implement comprehensive workflow statistics
	stats := map[string]interface{}{
		"totalTemplates":     0, // TODO: Get from repository
		"totalPackages":      0, // TODO: Implement packages
		"templatesThisMonth": 0, // TODO: Calculate from database
		"mostUsedTemplate":   nil, // TODO: Get from repository
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}