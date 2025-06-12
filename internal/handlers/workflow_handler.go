package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/repository"

	"github.com/gin-gonic/gin"
)

type WorkflowHandler struct {
	templateRepo *repository.JobTemplateRepository
	jobRepo      *repository.JobRepository
	customerRepo *repository.CustomerRepository
	packageRepo  *repository.EquipmentPackageRepository
	deviceRepo   *repository.DeviceRepository
}

func NewWorkflowHandler(templateRepo *repository.JobTemplateRepository, jobRepo *repository.JobRepository, customerRepo *repository.CustomerRepository, packageRepo *repository.EquipmentPackageRepository, deviceRepo *repository.DeviceRepository) *WorkflowHandler {
	return &WorkflowHandler{
		templateRepo: templateRepo,
		jobRepo:      jobRepo,
		customerRepo: customerRepo,
		packageRepo:  packageRepo,
		deviceRepo:   deviceRepo,
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
	
	c.HTML(http.StatusOK, "job_template_form_simple.html", gin.H{
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
		c.HTML(http.StatusUnauthorized, "job_template_form_simple.html", gin.H{
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
		c.HTML(http.StatusBadRequest, "job_template_form_simple.html", gin.H{
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
			c.HTML(http.StatusBadRequest, "job_template_form_simple.html", gin.H{
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
			c.HTML(http.StatusBadRequest, "job_template_form_simple.html", gin.H{
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
		c.HTML(http.StatusInternalServerError, "job_template_form_simple.html", gin.H{
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

	c.HTML(http.StatusOK, "equipment_package_form_simple.html", gin.H{
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
		c.HTML(http.StatusUnauthorized, "equipment_package_form_simple.html", gin.H{
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
		c.HTML(http.StatusBadRequest, "equipment_package_form_simple.html", gin.H{
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
		c.HTML(http.StatusInternalServerError, "equipment_package_form_simple.html", gin.H{
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
	// TODO: Implement bulk QR code generation
	log.Printf("BulkGenerateQRCodes: Not yet implemented")
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Bulk QR code generation not yet implemented",
	})
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