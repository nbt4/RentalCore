package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/repository"
	"go-barcode-webapp/internal/services"

	"github.com/gin-gonic/gin"
)

type InvoiceHandler struct {
	invoiceRepo  *repository.InvoiceRepository
	customerRepo *repository.CustomerRepository
	jobRepo      *repository.JobRepository
	deviceRepo   *repository.DeviceRepository
	packageRepo  *repository.EquipmentPackageRepository
	pdfService   *services.PDFService
	emailService *services.EmailService
}

func NewInvoiceHandler(
	invoiceRepo *repository.InvoiceRepository,
	customerRepo *repository.CustomerRepository,
	jobRepo *repository.JobRepository,
	deviceRepo *repository.DeviceRepository,
	packageRepo *repository.EquipmentPackageRepository,
) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceRepo:  invoiceRepo,
		customerRepo: customerRepo,
		jobRepo:      jobRepo,
		deviceRepo:   deviceRepo,
		packageRepo:  packageRepo,
		pdfService:   services.NewPDFService(),
		emailService: services.NewEmailService(),
	}
}

// ================================================================
// WEB HANDLERS
// ================================================================

// ListInvoices displays all invoices
func (h *InvoiceHandler) ListInvoices(c *gin.Context) {
	user, _ := GetCurrentUser(c)

	// Parse filter parameters
	var filter models.InvoiceFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		log.Printf("ListInvoices: Filter binding error: %v", err)
	}

	// Set default pagination
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	// Get invoices
	invoices, totalCount, err := h.invoiceRepo.GetInvoices(&filter)
	if err != nil {
		log.Printf("ListInvoices: Error fetching invoices: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load invoices",
			"user":  user,
		})
		return
	}

	// Get customers for filter dropdown
	customers, err := h.customerRepo.GetAll()
	if err != nil {
		log.Printf("ListInvoices: Error fetching customers: %v", err)
		customers = []models.Customer{} // Use empty slice on error
	}

	// Get jobs for filter dropdown
	jobs, err := h.jobRepo.GetAll(&models.FilterParams{Limit: 100})
	if err != nil {
		log.Printf("ListInvoices: Error fetching jobs: %v", err)
		jobs = []models.Job{} // Use empty slice on error
	}

	// Calculate pagination
	totalPages := int((totalCount + int64(filter.PageSize) - 1) / int64(filter.PageSize))

	c.HTML(http.StatusOK, "invoices_list.html", gin.H{
		"title":      "Invoices",
		"invoices":   invoices,
		"customers":  customers,
		"jobs":       jobs,
		"filter":     filter,
		"totalCount": totalCount,
		"totalPages": totalPages,
		"user":       user,
	})
}

// NewInvoiceForm displays the form for creating a new invoice
func (h *InvoiceHandler) NewInvoiceForm(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Get customers for dropdown
	customers, err := h.customerRepo.GetAll()
	if err != nil {
		log.Printf("NewInvoiceForm: Error fetching customers: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load customers",
			"user":  user,
		})
		return
	}

	// Get jobs for dropdown
	jobs, err := h.jobRepo.GetAll(&models.FilterParams{Limit: 100})
	if err != nil {
		log.Printf("NewInvoiceForm: Error fetching jobs: %v", err)
		jobs = []models.Job{}
	}

	// Get devices for line items
	devices, err := h.deviceRepo.GetAll(&models.FilterParams{Limit: 200})
	if err != nil {
		log.Printf("NewInvoiceForm: Error fetching devices: %v", err)
		devices = []models.Device{}
	}

	// Get equipment packages
	packages, err := h.packageRepo.GetActivePackages()
	if err != nil {
		log.Printf("NewInvoiceForm: Error fetching packages: %v", err)
		packages = []models.EquipmentPackage{}
	}

	// Get invoice templates
	templates, err := h.invoiceRepo.GetInvoiceTemplates()
	if err != nil {
		log.Printf("NewInvoiceForm: Error fetching templates: %v", err)
		templates = []models.InvoiceTemplate{}
	}

	// Get invoice settings
	settings, err := h.invoiceRepo.GetAllInvoiceSettings()
	if err != nil {
		log.Printf("NewInvoiceForm: Error fetching settings: %v", err)
		settings = &models.InvoiceSettings{
			DefaultTaxRate:      19.0,
			DefaultPaymentTerms: 30,
			CurrencySymbol:      "€",
		}
	}

	// Pre-fill dates
	now := time.Now()
	dueDate := now.AddDate(0, 0, settings.DefaultPaymentTerms)

	c.HTML(http.StatusOK, "invoice_form.html", gin.H{
		"title":         "New Invoice",
		"invoice":       &models.Invoice{},
		"customers":     customers,
		"jobs":          jobs,
		"devices":       devices,
		"packages":      packages,
		"templates":     templates,
		"settings":      settings,
		"defaultIssueDate": now.Format("2006-01-02"),
		"defaultDueDate":   dueDate.Format("2006-01-02"),
		"isEdit":        false,
		"user":          user,
	})
}

// CreateInvoice creates a new invoice
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var request models.InvoiceCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("CreateInvoice: Validation error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input data",
			"details": err.Error(),
		})
		return
	}

	// Create invoice from request
	invoice := models.Invoice{
		CustomerID:      request.CustomerID,
		JobID:           request.JobID,
		TemplateID:      request.TemplateID,
		Status:          "draft",
		IssueDate:       request.IssueDate,
		DueDate:         request.DueDate,
		PaymentTerms:    request.PaymentTerms,
		TaxRate:         request.TaxRate,
		DiscountAmount:  request.DiscountAmount,
		Notes:           request.Notes,
		TermsConditions: request.TermsConditions,
		CreatedBy:       &user.UserID,
	}

	// Create line items
	for i, item := range request.LineItems {
		lineItem := models.InvoiceLineItem{
			ItemType:        item.ItemType,
			DeviceID:        item.DeviceID,
			PackageID:       item.PackageID,
			Description:     item.Description,
			Quantity:        item.Quantity,
			UnitPrice:       item.UnitPrice,
			RentalStartDate: item.RentalStartDate,
			RentalEndDate:   item.RentalEndDate,
			RentalDays:      item.RentalDays,
			SortOrder:       func() *uint { order := uint(i); return &order }(),
		}
		lineItem.CalculateTotal()
		invoice.LineItems = append(invoice.LineItems, lineItem)
	}

	// Create the invoice
	if err := h.invoiceRepo.CreateInvoice(&invoice); err != nil {
		log.Printf("CreateInvoice: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invoice"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Invoice created successfully",
		"invoiceId": invoice.InvoiceID,
		"invoiceNumber": invoice.InvoiceNumber,
	})
}

// GetInvoice displays a specific invoice
func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	user, _ := GetCurrentUser(c)

	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Invalid invoice ID",
			"user":  user,
		})
		return
	}

	invoice, err := h.invoiceRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Invoice not found",
			"user":  user,
		})
		return
	}

	// Get company settings for display
	company, err := h.invoiceRepo.GetCompanySettings()
	if err != nil {
		log.Printf("GetInvoice: Error fetching company settings: %v", err)
		company = &models.CompanySettings{CompanyName: "RentalCore Company"}
	}

	// Get invoice settings
	settings, err := h.invoiceRepo.GetAllInvoiceSettings()
	if err != nil {
		log.Printf("GetInvoice: Error fetching settings: %v", err)
		settings = &models.InvoiceSettings{CurrencySymbol: "€"}
	}

	c.HTML(http.StatusOK, "invoice_detail.html", gin.H{
		"title":    "Invoice " + invoice.InvoiceNumber,
		"invoice":  invoice,
		"company":  company,
		"settings": settings,
		"user":     user,
	})
}

// EditInvoiceForm displays the form for editing an invoice
func (h *InvoiceHandler) EditInvoiceForm(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Invalid invoice ID",
			"user":  user,
		})
		return
	}

	invoice, err := h.invoiceRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Invoice not found",
			"user":  user,
		})
		return
	}

	// Only allow editing of draft invoices
	if invoice.Status != "draft" {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"error": "Only draft invoices can be edited",
			"user":  user,
		})
		return
	}

	// Get necessary data for form
	customers, _ := h.customerRepo.GetAll()
	jobs, _ := h.jobRepo.GetAll(&models.FilterParams{Limit: 100})
	devices, _ := h.deviceRepo.GetAll(&models.FilterParams{Limit: 200})
	packages, _ := h.packageRepo.GetActivePackages()
	templates, _ := h.invoiceRepo.GetInvoiceTemplates()
	settings, _ := h.invoiceRepo.GetAllInvoiceSettings()

	c.HTML(http.StatusOK, "invoice_form.html", gin.H{
		"title":     "Edit Invoice",
		"invoice":   invoice,
		"customers": customers,
		"jobs":      jobs,
		"devices":   devices,
		"packages":  packages,
		"templates": templates,
		"settings":  settings,
		"isEdit":    true,
		"user":      user,
	})
}

// UpdateInvoice updates an existing invoice
func (h *InvoiceHandler) UpdateInvoice(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	// Get existing invoice
	existingInvoice, err := h.invoiceRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	// Only allow editing of draft invoices
	if existingInvoice.Status != "draft" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only draft invoices can be edited"})
		return
	}

	var request models.InvoiceCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input data",
			"details": err.Error(),
		})
		return
	}

	// Update invoice fields
	existingInvoice.CustomerID = request.CustomerID
	existingInvoice.JobID = request.JobID
	existingInvoice.TemplateID = request.TemplateID
	existingInvoice.IssueDate = request.IssueDate
	existingInvoice.DueDate = request.DueDate
	existingInvoice.PaymentTerms = request.PaymentTerms
	existingInvoice.TaxRate = request.TaxRate
	existingInvoice.DiscountAmount = request.DiscountAmount
	existingInvoice.Notes = request.Notes
	existingInvoice.TermsConditions = request.TermsConditions

	// Replace line items
	existingInvoice.LineItems = []models.InvoiceLineItem{}
	for i, item := range request.LineItems {
		lineItem := models.InvoiceLineItem{
			InvoiceID:       invoiceID,
			ItemType:        item.ItemType,
			DeviceID:        item.DeviceID,
			PackageID:       item.PackageID,
			Description:     item.Description,
			Quantity:        item.Quantity,
			UnitPrice:       item.UnitPrice,
			RentalStartDate: item.RentalStartDate,
			RentalEndDate:   item.RentalEndDate,
			RentalDays:      item.RentalDays,
			SortOrder:       func() *uint { order := uint(i); return &order }(),
		}
		lineItem.CalculateTotal()
		existingInvoice.LineItems = append(existingInvoice.LineItems, lineItem)
	}

	// Update the invoice
	if err := h.invoiceRepo.UpdateInvoice(existingInvoice); err != nil {
		log.Printf("UpdateInvoice: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invoice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invoice updated successfully",
		"invoiceId": existingInvoice.InvoiceID,
	})
}

// DeleteInvoice deletes an invoice
func (h *InvoiceHandler) DeleteInvoice(c *gin.Context) {
	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	if err := h.invoiceRepo.DeleteInvoice(invoiceID); err != nil {
		log.Printf("DeleteInvoice: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete invoice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invoice deleted successfully"})
}

// UpdateInvoiceStatus updates the status of an invoice
func (h *InvoiceHandler) UpdateInvoiceStatus(c *gin.Context) {
	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	var request struct {
		Status string `json:"status" binding:"required,oneof=draft sent paid overdue cancelled"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid status",
			"details": err.Error(),
		})
		return
	}

	if err := h.invoiceRepo.UpdateInvoiceStatus(invoiceID, request.Status); err != nil {
		log.Printf("UpdateInvoiceStatus: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invoice status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invoice status updated successfully",
		"status":  request.Status,
	})
}

// ================================================================
// COMPANY SETTINGS
// ================================================================

// CompanySettingsForm displays the company settings form
func (h *InvoiceHandler) CompanySettingsForm(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	company, err := h.invoiceRepo.GetCompanySettings()
	if err != nil {
		log.Printf("CompanySettingsForm: Error fetching settings: %v", err)
		company = &models.CompanySettings{CompanyName: "RentalCore Company"}
	}

	c.HTML(http.StatusOK, "company_settings_form.html", gin.H{
		"title":   "Company Settings",
		"company": company,
		"user":    user,
	})
}

// UpdateCompanySettings updates company settings
func (h *InvoiceHandler) UpdateCompanySettings(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var company models.CompanySettings
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input data",
			"details": err.Error(),
		})
		return
	}

	// Preserve ID if updating existing settings
	existingCompany, err := h.invoiceRepo.GetCompanySettings()
	if err == nil {
		company.ID = existingCompany.ID
	}

	if err := h.invoiceRepo.UpdateCompanySettings(&company); err != nil {
		log.Printf("UpdateCompanySettings: Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Company settings updated successfully"})
}

// ================================================================
// INVOICE SETTINGS
// ================================================================

// InvoiceSettingsForm displays the invoice settings form
func (h *InvoiceHandler) InvoiceSettingsForm(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	settings, err := h.invoiceRepo.GetAllInvoiceSettings()
	if err != nil {
		log.Printf("InvoiceSettingsForm: Error fetching settings: %v", err)
		settings = &models.InvoiceSettings{}
	}

	c.HTML(http.StatusOK, "invoice_settings_form.html", gin.H{
		"title":    "Invoice Settings",
		"settings": settings,
		"user":     user,
	})
}

// UpdateInvoiceSettings updates invoice settings
func (h *InvoiceHandler) UpdateInvoiceSettings(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var settings models.InvoiceSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input data",
			"details": err.Error(),
		})
		return
	}

	// Update each setting
	settingsMap := map[string]string{
		"invoice_number_prefix":      settings.InvoiceNumberPrefix,
		"invoice_number_format":      settings.InvoiceNumberFormat,
		"default_payment_terms":      strconv.Itoa(settings.DefaultPaymentTerms),
		"default_tax_rate":           strconv.FormatFloat(settings.DefaultTaxRate, 'f', 2, 64),
		"auto_calculate_rental_days": strconv.FormatBool(settings.AutoCalculateRentalDays),
		"show_logo_on_invoice":       strconv.FormatBool(settings.ShowLogoOnInvoice),
		"currency_symbol":            settings.CurrencySymbol,
		"currency_code":              settings.CurrencyCode,
		"date_format":                settings.DateFormat,
	}

	for key, value := range settingsMap {
		if err := h.invoiceRepo.UpdateInvoiceSetting(key, value, &user.UserID); err != nil {
			log.Printf("UpdateInvoiceSettings: Error updating %s: %v", key, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update invoice settings",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invoice settings updated successfully"})
}

// ================================================================
// API HANDLERS
// ================================================================

// GetInvoicesAPI returns invoices as JSON
func (h *InvoiceHandler) GetInvoicesAPI(c *gin.Context) {
	var filter models.InvoiceFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoices, totalCount, err := h.invoiceRepo.GetInvoices(&filter)
	if err != nil {
		log.Printf("GetInvoicesAPI: Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load invoices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invoices":   invoices,
		"totalCount": totalCount,
		"filter":     filter,
	})
}

// GetInvoiceStatsAPI returns invoice statistics
func (h *InvoiceHandler) GetInvoiceStatsAPI(c *gin.Context) {
	stats, err := h.invoiceRepo.GetInvoiceStats()
	if err != nil {
		log.Printf("GetInvoiceStatsAPI: Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get invoice statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ================================================================
// PDF AND EMAIL HANDLERS
// ================================================================

// GenerateInvoicePDF generates and downloads a PDF for an invoice
func (h *InvoiceHandler) GenerateInvoicePDF(c *gin.Context) {
	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	// Get invoice
	invoice, err := h.invoiceRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	// Get company settings
	company, err := h.invoiceRepo.GetCompanySettings()
	if err != nil {
		log.Printf("GenerateInvoicePDF: Error fetching company settings: %v", err)
		company = &models.CompanySettings{CompanyName: "RentalCore Company"}
	}

	// Get invoice settings
	settings, err := h.invoiceRepo.GetAllInvoiceSettings()
	if err != nil {
		log.Printf("GenerateInvoicePDF: Error fetching settings: %v", err)
		settings = &models.InvoiceSettings{CurrencySymbol: "€"}
	}

	// Generate PDF
	pdfBytes, err := h.pdfService.GenerateInvoicePDF(invoice, company, settings)
	if err != nil {
		log.Printf("GenerateInvoicePDF: Error generating PDF: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}

	// Set headers for PDF download
	filename := fmt.Sprintf("Invoice_%s.pdf", invoice.InvoiceNumber)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

	// Send PDF
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// EmailInvoice sends an invoice via email
func (h *InvoiceHandler) EmailInvoice(c *gin.Context) {
	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	var request struct {
		ToEmail     string `json:"toEmail"`
		Subject     string `json:"subject"`
		Message     string `json:"message"`
		IncludePDF  bool   `json:"includePdf"`
		MarkAsSent  bool   `json:"markAsSent"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Get invoice
	invoice, err := h.invoiceRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	// Get company settings
	company, err := h.invoiceRepo.GetCompanySettings()
	if err != nil {
		log.Printf("EmailInvoice: Error fetching company settings: %v", err)
		company = &models.CompanySettings{CompanyName: "RentalCore Company"}
	}

	// Get invoice settings
	settings, err := h.invoiceRepo.GetAllInvoiceSettings()
	if err != nil {
		log.Printf("EmailInvoice: Error fetching settings: %v", err)
		settings = &models.InvoiceSettings{CurrencySymbol: "€"}
	}

	// Use customer email if not specified
	toEmail := request.ToEmail
	if toEmail == "" && invoice.Customer != nil {
		toEmail = invoice.Customer.Email
	}

	if toEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No email address specified"})
		return
	}

	// Prepare email data
	emailData := &services.EmailData{
		Invoice:      invoice,
		Company:      company,
		Customer:     invoice.Customer,
		Settings:     settings,
		InvoiceURL:   fmt.Sprintf("%s/invoices/%d", c.Request.Host, invoice.InvoiceID),
		SupportEmail: company.Email,
	}

	// Generate PDF if requested
	var pdfAttachment []byte
	if request.IncludePDF {
		pdfBytes, err := h.pdfService.GenerateInvoicePDF(invoice, company, settings)
		if err != nil {
			log.Printf("EmailInvoice: Error generating PDF: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF attachment"})
			return
		}
		pdfAttachment = pdfBytes
	}

	// Send email
	if err := h.emailService.SendInvoiceEmail(emailData, pdfAttachment); err != nil {
		log.Printf("EmailInvoice: Error sending email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send email",
			"details": err.Error(),
		})
		return
	}

	// Mark invoice as sent if requested
	if request.MarkAsSent && invoice.Status == "draft" {
		if err := h.invoiceRepo.UpdateInvoiceStatus(invoiceID, "sent"); err != nil {
			log.Printf("EmailInvoice: Error updating invoice status: %v", err)
			// Don't fail the request, email was sent successfully
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invoice sent successfully",
		"sentTo":  toEmail,
	})
}

// TestEmailSettings sends a test email to verify configuration
func (h *InvoiceHandler) TestEmailSettings(c *gin.Context) {
	var request struct {
		ToEmail string `json:"toEmail" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Send test email
	if err := h.emailService.SendTestEmail(request.ToEmail, nil); err != nil {
		log.Printf("TestEmailSettings: Error sending test email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send test email",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test email sent successfully",
		"sentTo":  request.ToEmail,
	})
}

// PreviewInvoicePDF generates and displays a PDF preview in browser
func (h *InvoiceHandler) PreviewInvoicePDF(c *gin.Context) {
	invoiceIDStr := c.Param("id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	// Get invoice
	invoice, err := h.invoiceRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	// Get company settings
	company, err := h.invoiceRepo.GetCompanySettings()
	if err != nil {
		log.Printf("PreviewInvoicePDF: Error fetching company settings: %v", err)
		company = &models.CompanySettings{CompanyName: "RentalCore Company"}
	}

	// Get invoice settings
	settings, err := h.invoiceRepo.GetAllInvoiceSettings()
	if err != nil {
		log.Printf("PreviewInvoicePDF: Error fetching settings: %v", err)
		settings = &models.InvoiceSettings{CurrencySymbol: "€"}
	}

	// Generate PDF
	pdfBytes, err := h.pdfService.GenerateInvoicePDF(invoice, company, settings)
	if err != nil {
		log.Printf("PreviewInvoicePDF: Error generating PDF: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}

	// Set headers for PDF preview (inline display)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline")
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

	// Send PDF
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}