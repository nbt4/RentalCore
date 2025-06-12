package repository

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-barcode-webapp/internal/models"

	"gorm.io/gorm"
)

type InvoiceRepository struct {
	db *Database
}

func NewInvoiceRepository(db *Database) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

// ================================================================
// INVOICE MANAGEMENT
// ================================================================

// GetInvoices returns a paginated list of invoices with filters
func (r *InvoiceRepository) GetInvoices(filter *models.InvoiceFilter) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var totalCount int64

	query := r.db.DB.Model(&models.Invoice{}).
		Preload("Customer").
		Preload("Job").
		Preload("Template").
		Preload("Creator").
		Preload("LineItems").
		Preload("Payments")

	// Apply filters
	if filter != nil {
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.CustomerID != nil {
			query = query.Where("customer_id = ?", *filter.CustomerID)
		}
		if filter.JobID != nil {
			query = query.Where("job_id = ?", *filter.JobID)
		}
		if filter.StartDate != nil {
			query = query.Where("issue_date >= ?", *filter.StartDate)
		}
		if filter.EndDate != nil {
			query = query.Where("issue_date <= ?", *filter.EndDate)
		}
		if filter.MinAmount != nil {
			query = query.Where("total_amount >= ?", *filter.MinAmount)
		}
		if filter.MaxAmount != nil {
			query = query.Where("total_amount <= ?", *filter.MaxAmount)
		}
		if filter.OverdueOnly {
			query = query.Where("due_date < ? AND status NOT IN ('paid', 'cancelled')", time.Now())
		}
		if filter.SearchTerm != "" {
			searchTerm := "%" + filter.SearchTerm + "%"
			query = query.Where("invoice_number ILIKE ? OR notes ILIKE ?", searchTerm, searchTerm)
		}
	}

	// Get total count
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %v", err)
	}

	// Apply pagination
	if filter != nil {
		if filter.PageSize > 0 {
			query = query.Limit(filter.PageSize)
		}
		if filter.Page > 0 {
			offset := (filter.Page - 1) * filter.PageSize
			query = query.Offset(offset)
		}
	}

	// Order by issue date desc
	query = query.Order("issue_date DESC, created_at DESC")

	if err := query.Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get invoices: %v", err)
	}

	return invoices, totalCount, nil
}

// GetInvoiceByID returns a specific invoice by ID
func (r *InvoiceRepository) GetInvoiceByID(id uint64) (*models.Invoice, error) {
	var invoice models.Invoice

	if err := r.db.DB.
		Preload("Customer").
		Preload("Job").
		Preload("Template").
		Preload("Creator").
		Preload("LineItems", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, line_item_id ASC")
		}).
		Preload("LineItems.Device").
		Preload("LineItems.Package").
		Preload("Payments", func(db *gorm.DB) *gorm.DB {
			return db.Order("payment_date DESC")
		}).
		First(&invoice, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %v", err)
	}

	return &invoice, nil
}

// GetInvoiceByNumber returns a specific invoice by invoice number
func (r *InvoiceRepository) GetInvoiceByNumber(invoiceNumber string) (*models.Invoice, error) {
	var invoice models.Invoice

	if err := r.db.DB.
		Preload("Customer").
		Preload("Job").
		Preload("Template").
		Preload("Creator").
		Preload("LineItems", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, line_item_id ASC")
		}).
		Preload("LineItems.Device").
		Preload("LineItems.Package").
		Preload("Payments", func(db *gorm.DB) *gorm.DB {
			return db.Order("payment_date DESC")
		}).
		Where("invoice_number = ?", invoiceNumber).
		First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %v", err)
	}

	return &invoice, nil
}

// CreateInvoice creates a new invoice with line items
func (r *InvoiceRepository) CreateInvoice(invoice *models.Invoice) error {
	return r.db.DB.Transaction(func(tx *gorm.DB) error {
		// Generate invoice number if not provided
		if invoice.InvoiceNumber == "" {
			invoiceNumber, err := r.generateInvoiceNumber()
			if err != nil {
				return fmt.Errorf("failed to generate invoice number: %v", err)
			}
			invoice.InvoiceNumber = invoiceNumber
		}

		// Set timestamps
		now := time.Now()
		invoice.CreatedAt = now
		invoice.UpdatedAt = now

		// Calculate totals
		invoice.CalculateTotals()

		// Create the invoice
		if err := tx.Create(invoice).Error; err != nil {
			return fmt.Errorf("failed to create invoice: %v", err)
		}

		// Create line items
		for i := range invoice.LineItems {
			invoice.LineItems[i].InvoiceID = invoice.InvoiceID
			invoice.LineItems[i].CreatedAt = now
			invoice.LineItems[i].UpdatedAt = now
			invoice.LineItems[i].CalculateTotal()
		}

		return nil
	})
}

// UpdateInvoice updates an existing invoice
func (r *InvoiceRepository) UpdateInvoice(invoice *models.Invoice) error {
	return r.db.DB.Transaction(func(tx *gorm.DB) error {
		// Update timestamp
		invoice.UpdatedAt = time.Now()

		// Calculate totals
		invoice.CalculateTotals()

		// Update the invoice
		if err := tx.Save(invoice).Error; err != nil {
			return fmt.Errorf("failed to update invoice: %v", err)
		}

		// Update line items if they exist
		for i := range invoice.LineItems {
			invoice.LineItems[i].UpdatedAt = invoice.UpdatedAt
			invoice.LineItems[i].CalculateTotal()
		}

		return nil
	})
}

// DeleteInvoice soft deletes an invoice (sets status to cancelled)
func (r *InvoiceRepository) DeleteInvoice(id uint64) error {
	result := r.db.DB.Model(&models.Invoice{}).
		Where("invoice_id = ?", id).
		Update("status", "cancelled")

	if result.Error != nil {
		return fmt.Errorf("failed to delete invoice: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// UpdateInvoiceStatus updates the status of an invoice
func (r *InvoiceRepository) UpdateInvoiceStatus(id uint64, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	// Set special timestamps based on status
	if status == "sent" {
		updates["sent_at"] = time.Now()
	} else if status == "paid" {
		updates["paid_at"] = time.Now()
	}

	result := r.db.DB.Model(&models.Invoice{}).
		Where("invoice_id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update invoice status: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// ================================================================
// INVOICE NUMBER GENERATION
// ================================================================

// generateInvoiceNumber generates a unique invoice number
func (r *InvoiceRepository) generateInvoiceNumber() (string, error) {
	// Get the current year and month
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")

	// Get the invoice number format setting
	format, err := r.GetInvoiceSetting("invoice_number_format")
	if err != nil || format == "" {
		format = "{prefix}{year}{month}{sequence:4}"
	}

	// Get the prefix
	prefix, err := r.GetInvoiceSetting("invoice_number_prefix")
	if err != nil || prefix == "" {
		prefix = "INV-"
	}

	// Get the next sequence number for this year/month
	var maxSequence int
	err = r.db.DB.Raw(`
		SELECT COALESCE(MAX(CAST(SUBSTRING(invoice_number FROM LENGTH(?) + LENGTH(?) + LENGTH(?) + 1 FOR 4) AS UNSIGNED)), 0)
		FROM invoices 
		WHERE invoice_number LIKE ?
	`, prefix, year, month, prefix+year+month+"%").Scan(&maxSequence).Error

	if err != nil {
		return "", fmt.Errorf("failed to get next sequence number: %v", err)
	}

	sequence := maxSequence + 1

	// Replace placeholders in format
	invoiceNumber := strings.ReplaceAll(format, "{prefix}", prefix)
	invoiceNumber = strings.ReplaceAll(invoiceNumber, "{year}", year)
	invoiceNumber = strings.ReplaceAll(invoiceNumber, "{month}", month)
	invoiceNumber = strings.ReplaceAll(invoiceNumber, "{sequence:4}", fmt.Sprintf("%04d", sequence))

	return invoiceNumber, nil
}

// ================================================================
// INVOICE PAYMENTS
// ================================================================

// AddPayment adds a payment to an invoice
func (r *InvoiceRepository) AddPayment(payment *models.InvoicePayment) error {
	return r.db.DB.Transaction(func(tx *gorm.DB) error {
		// Create the payment
		payment.CreatedAt = time.Now()
		if err := tx.Create(payment).Error; err != nil {
			return fmt.Errorf("failed to create payment: %v", err)
		}

		// Update invoice paid amount and status
		var invoice models.Invoice
		if err := tx.First(&invoice, payment.InvoiceID).Error; err != nil {
			return fmt.Errorf("failed to get invoice for payment update: %v", err)
		}

		// Calculate new paid amount
		var totalPaid float64
		if err := tx.Model(&models.InvoicePayment{}).
			Where("invoice_id = ?", payment.InvoiceID).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&totalPaid).Error; err != nil {
			return fmt.Errorf("failed to calculate total payments: %v", err)
		}

		// Update invoice
		updates := map[string]interface{}{
			"paid_amount": totalPaid,
			"balance_due": invoice.TotalAmount - totalPaid,
			"updated_at":  time.Now(),
		}

		// Update status if fully paid
		if totalPaid >= invoice.TotalAmount {
			updates["status"] = "paid"
			updates["paid_at"] = time.Now()
		}

		if err := tx.Model(&invoice).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update invoice with payment: %v", err)
		}

		return nil
	})
}

// ================================================================
// COMPANY SETTINGS
// ================================================================

// GetCompanySettings returns the company settings
func (r *InvoiceRepository) GetCompanySettings() (*models.CompanySettings, error) {
	var settings models.CompanySettings

	if err := r.db.DB.First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("company settings not found")
		}
		return nil, fmt.Errorf("failed to get company settings: %v", err)
	}

	return &settings, nil
}

// UpdateCompanySettings updates the company settings
func (r *InvoiceRepository) UpdateCompanySettings(settings *models.CompanySettings) error {
	settings.UpdatedAt = time.Now()

	if err := r.db.DB.Save(settings).Error; err != nil {
		return fmt.Errorf("failed to update company settings: %v", err)
	}

	return nil
}

// ================================================================
// INVOICE SETTINGS
// ================================================================

// GetInvoiceSetting returns a specific invoice setting
func (r *InvoiceRepository) GetInvoiceSetting(key string) (string, error) {
	var setting models.InvoiceSetting

	if err := r.db.DB.Where("setting_key = ?", key).First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil // Return empty string for missing settings
		}
		return "", fmt.Errorf("failed to get invoice setting: %v", err)
	}

	if setting.SettingValue != nil {
		return *setting.SettingValue, nil
	}

	return "", nil
}

// GetAllInvoiceSettings returns all invoice settings as a map
func (r *InvoiceRepository) GetAllInvoiceSettings() (*models.InvoiceSettings, error) {
	var dbSettings []models.InvoiceSetting

	if err := r.db.DB.Find(&dbSettings).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoice settings: %v", err)
	}

	// Convert to structured format
	settings := &models.InvoiceSettings{
		InvoiceNumberPrefix:     "INV-",
		InvoiceNumberFormat:     "{prefix}{year}{month}{sequence:4}",
		DefaultPaymentTerms:     30,
		DefaultTaxRate:          19.0,
		AutoCalculateRentalDays: true,
		ShowLogoOnInvoice:       true,
		CurrencySymbol:          "€",
		CurrencyCode:            "EUR",
		DateFormat:              "DD.MM.YYYY",
	}

	// Override with database values
	for _, setting := range dbSettings {
		if setting.SettingValue == nil {
			continue
		}

		switch setting.SettingKey {
		case "invoice_number_prefix":
			settings.InvoiceNumberPrefix = *setting.SettingValue
		case "invoice_number_format":
			settings.InvoiceNumberFormat = *setting.SettingValue
		case "default_payment_terms":
			if val, err := strconv.Atoi(*setting.SettingValue); err == nil {
				settings.DefaultPaymentTerms = val
			}
		case "default_tax_rate":
			if val, err := strconv.ParseFloat(*setting.SettingValue, 64); err == nil {
				settings.DefaultTaxRate = val
			}
		case "auto_calculate_rental_days":
			settings.AutoCalculateRentalDays = *setting.SettingValue == "true"
		case "show_logo_on_invoice":
			settings.ShowLogoOnInvoice = *setting.SettingValue == "true"
		case "currency_symbol":
			settings.CurrencySymbol = *setting.SettingValue
		case "currency_code":
			settings.CurrencyCode = *setting.SettingValue
		case "date_format":
			settings.DateFormat = *setting.SettingValue
		}
	}

	return settings, nil
}

// UpdateInvoiceSetting updates a specific invoice setting
func (r *InvoiceRepository) UpdateInvoiceSetting(key, value string, updatedBy *uint) error {
	setting := models.InvoiceSetting{
		SettingKey:   key,
		SettingValue: &value,
		UpdatedBy:    updatedBy,
		UpdatedAt:    time.Now(),
	}

	if err := r.db.DB.Save(&setting).Error; err != nil {
		return fmt.Errorf("failed to update invoice setting: %v", err)
	}

	return nil
}

// ================================================================
// INVOICE TEMPLATES
// ================================================================

// GetInvoiceTemplates returns all invoice templates
func (r *InvoiceRepository) GetInvoiceTemplates() ([]models.InvoiceTemplate, error) {
	var templates []models.InvoiceTemplate

	if err := r.db.DB.Where("is_active = ?", true).
		Order("is_default DESC, name ASC").
		Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoice templates: %v", err)
	}

	return templates, nil
}

// GetDefaultInvoiceTemplate returns the default invoice template
func (r *InvoiceRepository) GetDefaultInvoiceTemplate() (*models.InvoiceTemplate, error) {
	var template models.InvoiceTemplate

	if err := r.db.DB.Where("is_default = ? AND is_active = ?", true, true).
		First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no default invoice template found")
		}
		return nil, fmt.Errorf("failed to get default invoice template: %v", err)
	}

	return &template, nil
}

// ================================================================
// STATISTICS AND REPORTING
// ================================================================

// GetInvoiceStats returns invoice statistics
func (r *InvoiceRepository) GetInvoiceStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total invoices
	var totalInvoices int64
	r.db.DB.Model(&models.Invoice{}).Count(&totalInvoices)
	stats["total_invoices"] = totalInvoices

	// Total revenue
	var totalRevenue float64
	r.db.DB.Model(&models.Invoice{}).
		Where("status = ?", "paid").
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&totalRevenue)
	stats["total_revenue"] = totalRevenue

	// Outstanding amount
	var outstanding float64
	r.db.DB.Model(&models.Invoice{}).
		Where("status NOT IN ('paid', 'cancelled')").
		Select("COALESCE(SUM(balance_due), 0)").
		Scan(&outstanding)
	stats["outstanding_amount"] = outstanding

	// Overdue invoices
	var overdueCount int64
	r.db.DB.Model(&models.Invoice{}).
		Where("due_date < ? AND status NOT IN ('paid', 'cancelled')", time.Now()).
		Count(&overdueCount)
	stats["overdue_count"] = overdueCount

	return stats, nil
}