package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go-barcode-webapp/internal/config"
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
	emailConfig *config.EmailConfig,
	pdfConfig *config.PDFConfig,
) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceRepo:  invoiceRepo,
		customerRepo: customerRepo,
		jobRepo:      jobRepo,
		deviceRepo:   deviceRepo,
		packageRepo:  packageRepo,
		pdfService:   services.NewPDFService(pdfConfig),
		emailService: services.NewEmailService(emailConfig),
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
	invoices, _, err := h.invoiceRepo.GetInvoices(&filter)
	if err != nil {
		log.Printf("ListInvoices: Error fetching invoices: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load invoices",
			"user":  user,
		})
		return
	}

	// Note: Removed unused customers, jobsWithDetails, and totalPages variables
	// that were causing build failures. These can be re-added when filtering
	// functionality is implemented.

	c.HTML(http.StatusOK, "invoices_list.html", gin.H{
		"title":    "Invoices",
		"invoices": invoices,
		"user":     user,
	})
}

// NewInvoiceForm displays the form for creating a new invoice
func (h *InvoiceHandler) NewInvoiceForm(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Note: Simplified data to avoid JavaScript context errors
	// Complex data structures removed temporarily
	now := time.Now()
	dueDate := now.AddDate(0, 0, 30) // Default 30 days

	c.HTML(http.StatusOK, "invoice_form.html", gin.H{
		"title":            "New Invoice",
		"invoice":          &models.Invoice{},
		"defaultIssueDate": now.Format("2006-01-02"),
		"defaultDueDate":   dueDate.Format("2006-01-02"),
		"isEdit":           false,
		"user":             user,
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
	customers, _ := h.customerRepo.List(&models.FilterParams{Limit: 100})
	jobsWithDetails, _ := h.jobRepo.List(&models.FilterParams{Limit: 100})
	devicesWithJobInfo, _ := h.deviceRepo.List(&models.FilterParams{Limit: 200})

	// Extract Device objects from DeviceWithJobInfo
	var devices []models.Device
	for _, dwj := range devicesWithJobInfo {
		devices = append(devices, dwj.Device)
	}
	packages, _ := h.packageRepo.GetActivePackages()
	templates, _ := h.invoiceRepo.GetInvoiceTemplates()
	settings, _ := h.invoiceRepo.GetAllInvoiceSettings()

	c.HTML(http.StatusOK, "invoice_form.html", gin.H{
		"title":     "Edit Invoice",
		"invoice":   invoice,
		"customers": customers,
		"jobs":      jobsWithDetails,
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
	_, exists := GetCurrentUser(c)
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
	_, exists := GetCurrentUser(c)
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
	if toEmail == "" && invoice.Customer != nil && invoice.Customer.Email != nil {
		toEmail = *invoice.Customer.Email
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
		SupportEmail: func() string { if company.Email != nil { return *company.Email } else { return "" } }(),
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

// ================================================================
// INVOICE TEMPLATE MANAGEMENT
// ================================================================

// ListInvoiceTemplates displays all invoice templates
func (h *InvoiceHandler) ListInvoiceTemplates(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	templates, err := h.invoiceRepo.GetInvoiceTemplates()
	if err != nil {
		log.Printf("ListInvoiceTemplates: Error fetching templates: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load templates",
			"user":  user,
		})
		return
	}

	c.HTML(http.StatusOK, "invoice_templates_list.html", gin.H{
		"title":     "Invoice Templates",
		"templates": templates,
		"user":      user,
	})
}

// NewInvoiceTemplateForm displays the form for creating a new invoice template
func (h *InvoiceHandler) NewInvoiceTemplateForm(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	c.HTML(http.StatusOK, "invoice_template_designer.html", gin.H{
		"title":    "New Invoice Template - Dummy-Friendly Designer",
		"template": &models.InvoiceTemplate{},
		"isEdit":   false,
		"user":     user,
	})
}

// CreateInvoiceTemplate creates a new invoice template
func (h *InvoiceHandler) CreateInvoiceTemplate(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var request struct {
		Name         string `json:"name" binding:"required"`
		Description  string `json:"description"`
		HTMLTemplate string `json:"htmlTemplate" binding:"required"`
		CSSStyles    string `json:"cssStyles"`
		IsDefault    bool   `json:"isDefault"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	// Create template
	template := models.InvoiceTemplate{
		Name:         request.Name,
		Description:  &request.Description,
		HTMLTemplate: request.HTMLTemplate,
		CSSStyles:    &request.CSSStyles,
		IsDefault:    request.IsDefault,
		IsActive:     true,
		CreatedBy:    &user.UserID,
	}

	if err := h.invoiceRepo.CreateInvoiceTemplate(&template); err != nil {
		log.Printf("CreateInvoiceTemplate: Error creating template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Template created successfully",
		"templateId": template.TemplateID,
	})
}

// GetInvoiceTemplate returns a specific invoice template
func (h *InvoiceHandler) GetInvoiceTemplate(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := h.invoiceRepo.GetInvoiceTemplateByID(uint(templateID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// EditInvoiceTemplateForm displays the form for editing an invoice template
func (h *InvoiceHandler) EditInvoiceTemplateForm(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Invalid template ID",
			"user":  user,
		})
		return
	}

	template, err := h.invoiceRepo.GetInvoiceTemplateByID(uint(templateID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Template not found",
			"user":  user,
		})
		return
	}

	c.HTML(http.StatusOK, "invoice_template_designer.html", gin.H{
		"title":    "Edit Invoice Template - Dummy-Friendly Designer",
		"template": template,
		"isEdit":   true,
		"user":     user,
	})
}

// UpdateInvoiceTemplate updates an existing invoice template
func (h *InvoiceHandler) UpdateInvoiceTemplate(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var request struct {
		Name         string `json:"name" binding:"required"`
		Description  string `json:"description"`
		HTMLTemplate string `json:"htmlTemplate" binding:"required"`
		CSSStyles    string `json:"cssStyles"`
		IsDefault    bool   `json:"isDefault"`
		IsActive     bool   `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	// Get existing template
	template, err := h.invoiceRepo.GetInvoiceTemplateByID(uint(templateID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	// Update template
	template.Name = request.Name
	template.Description = &request.Description
	template.HTMLTemplate = request.HTMLTemplate
	template.CSSStyles = &request.CSSStyles
	template.IsDefault = request.IsDefault
	template.IsActive = request.IsActive

	if err := h.invoiceRepo.UpdateInvoiceTemplate(template); err != nil {
		log.Printf("UpdateInvoiceTemplate: Error updating template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template updated successfully"})
}

// DeleteInvoiceTemplate deletes an invoice template
func (h *InvoiceHandler) DeleteInvoiceTemplate(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	if err := h.invoiceRepo.DeleteInvoiceTemplate(uint(templateID)); err != nil {
		log.Printf("DeleteInvoiceTemplate: Error deleting template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}