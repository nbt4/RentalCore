package models

import (
	"time"
)

// CompanySettings represents the company/business information
type CompanySettings struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CompanyName  string    `gorm:"not null;column:company_name" json:"companyName" binding:"required"`
	AddressLine1 *string   `gorm:"column:address_line1" json:"addressLine1"`
	AddressLine2 *string   `gorm:"column:address_line2" json:"addressLine2"`
	City         *string   `gorm:"column:city" json:"city"`
	State        *string   `gorm:"column:state" json:"state"`
	PostalCode   *string   `gorm:"column:postal_code" json:"postalCode"`
	Country      *string   `gorm:"column:country" json:"country"`
	Phone        *string   `gorm:"column:phone" json:"phone"`
	Email        *string   `gorm:"column:email" json:"email"`
	Website      *string   `gorm:"column:website" json:"website"`
	TaxNumber    *string   `gorm:"column:tax_number" json:"taxNumber"`
	VATNumber    *string   `gorm:"column:vat_number" json:"vatNumber"`
	LogoPath     *string   `gorm:"column:logo_path" json:"logoPath"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (CompanySettings) TableName() string {
	return "company_settings"
}

// InvoiceTemplate represents customizable invoice layouts
type InvoiceTemplate struct {
	TemplateID   uint      `gorm:"primaryKey;autoIncrement;column:template_id" json:"templateId"`
	Name         string    `gorm:"not null;column:name" json:"name" binding:"required"`
	Description  *string   `gorm:"column:description" json:"description"`
	HTMLTemplate string    `gorm:"type:longtext;not null;column:html_template" json:"htmlTemplate" binding:"required"`
	CSSStyles    *string   `gorm:"type:longtext;column:css_styles" json:"cssStyles"`
	IsDefault    bool      `gorm:"not null;default:false;column:is_default" json:"isDefault"`
	IsActive     bool      `gorm:"not null;default:true;column:is_active" json:"isActive"`
	CreatedBy    *uint     `gorm:"column:created_by" json:"createdBy"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`

	// Relationships
	Creator  *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Invoices []Invoice `gorm:"foreignKey:TemplateID" json:"invoices,omitempty"`
}

func (InvoiceTemplate) TableName() string {
	return "invoice_templates"
}

// Invoice represents an invoice document
type Invoice struct {
	InvoiceID       uint64              `gorm:"primaryKey;autoIncrement;column:invoice_id" json:"invoiceId"`
	InvoiceNumber   string              `gorm:"uniqueIndex;not null;column:invoice_number" json:"invoiceNumber" binding:"required"`
	CustomerID      uint                `gorm:"not null;column:customer_id" json:"customerId" binding:"required"`
	JobID           *uint               `gorm:"column:job_id" json:"jobId"`
	TemplateID      *uint               `gorm:"column:template_id" json:"templateId"`
	Status          string              `gorm:"type:enum('draft','sent','paid','overdue','cancelled');not null;default:'draft';column:status" json:"status"`
	IssueDate       time.Time           `gorm:"type:date;not null;column:issue_date" json:"issueDate" binding:"required"`
	DueDate         time.Time           `gorm:"type:date;not null;column:due_date" json:"dueDate" binding:"required"`
	PaymentTerms    *string             `gorm:"column:payment_terms" json:"paymentTerms"`

	// Financial Details
	Subtotal       float64 `gorm:"type:decimal(12,2);not null;default:0.00;column:subtotal" json:"subtotal"`
	TaxRate        float64 `gorm:"type:decimal(5,2);not null;default:0.00;column:tax_rate" json:"taxRate"`
	TaxAmount      float64 `gorm:"type:decimal(12,2);not null;default:0.00;column:tax_amount" json:"taxAmount"`
	DiscountAmount float64 `gorm:"type:decimal(12,2);not null;default:0.00;column:discount_amount" json:"discountAmount"`
	TotalAmount    float64 `gorm:"type:decimal(12,2);not null;default:0.00;column:total_amount" json:"totalAmount"`
	PaidAmount     float64 `gorm:"type:decimal(12,2);not null;default:0.00;column:paid_amount" json:"paidAmount"`
	BalanceDue     float64 `gorm:"type:decimal(12,2);not null;default:0.00;column:balance_due" json:"balanceDue"`

	// Additional Information
	Notes            *string `gorm:"type:text;column:notes" json:"notes"`
	TermsConditions  *string `gorm:"type:text;column:terms_conditions" json:"termsConditions"`
	InternalNotes    *string `gorm:"type:text;column:internal_notes" json:"internalNotes"`

	// Tracking
	SentAt    *time.Time `gorm:"column:sent_at" json:"sentAt"`
	PaidAt    *time.Time `gorm:"column:paid_at" json:"paidAt"`
	CreatedBy *uint      `gorm:"column:created_by" json:"createdBy"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updatedAt"`

	// Relationships
	Customer     *Customer           `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Job          *Job                `gorm:"foreignKey:JobID" json:"job,omitempty"`
	Template     *InvoiceTemplate    `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Creator      *User               `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	LineItems    []InvoiceLineItem   `gorm:"foreignKey:InvoiceID" json:"lineItems,omitempty"`
	Payments     []InvoicePayment    `gorm:"foreignKey:InvoiceID" json:"payments,omitempty"`
}

func (Invoice) TableName() string {
	return "invoices"
}

// CalculateTotals calculates and updates invoice totals
func (i *Invoice) CalculateTotals() {
	i.Subtotal = 0
	for _, item := range i.LineItems {
		i.Subtotal += item.TotalPrice
	}
	
	i.TaxAmount = (i.Subtotal - i.DiscountAmount) * (i.TaxRate / 100)
	i.TotalAmount = i.Subtotal - i.DiscountAmount + i.TaxAmount
	i.BalanceDue = i.TotalAmount - i.PaidAmount
}

// IsOverdue checks if the invoice is overdue
func (i *Invoice) IsOverdue() bool {
	return time.Now().After(i.DueDate) && i.Status != "paid" && i.Status != "cancelled"
}

// InvoiceLineItem represents individual items on an invoice
type InvoiceLineItem struct {
	LineItemID      uint64    `gorm:"primaryKey;autoIncrement;column:line_item_id" json:"lineItemId"`
	InvoiceID       uint64    `gorm:"not null;column:invoice_id" json:"invoiceId"`
	ItemType        string    `gorm:"type:enum('device','service','package','custom');not null;default:'custom';column:item_type" json:"itemType"`
	DeviceID        *string   `gorm:"column:device_id" json:"deviceId"`
	PackageID       *uint     `gorm:"column:package_id" json:"packageId"`
	Description     string    `gorm:"type:text;not null;column:description" json:"description" binding:"required"`
	Quantity        float64   `gorm:"type:decimal(10,2);not null;default:1.00;column:quantity" json:"quantity"`
	UnitPrice       float64   `gorm:"type:decimal(12,2);not null;default:0.00;column:unit_price" json:"unitPrice"`
	TotalPrice      float64   `gorm:"type:decimal(12,2);not null;default:0.00;column:total_price" json:"totalPrice"`
	RentalStartDate *time.Time `gorm:"type:date;column:rental_start_date" json:"rentalStartDate"`
	RentalEndDate   *time.Time `gorm:"type:date;column:rental_end_date" json:"rentalEndDate"`
	RentalDays      *int      `gorm:"column:rental_days" json:"rentalDays"`
	SortOrder       *uint     `gorm:"column:sort_order" json:"sortOrder"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updatedAt"`

	// Relationships
	Invoice *Invoice           `gorm:"foreignKey:InvoiceID" json:"invoice,omitempty"`
	Device  *Device            `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	Package *EquipmentPackage  `gorm:"foreignKey:PackageID" json:"package,omitempty"`
}

func (InvoiceLineItem) TableName() string {
	return "invoice_line_items"
}

// CalculateTotal calculates the total price for this line item
func (ili *InvoiceLineItem) CalculateTotal() {
	ili.TotalPrice = ili.Quantity * ili.UnitPrice
}

// InvoiceSettings represents configurable invoice settings
type InvoiceSetting struct {
	SettingID    uint      `gorm:"primaryKey;autoIncrement;column:setting_id" json:"settingId"`
	SettingKey   string    `gorm:"uniqueIndex;not null;column:setting_key" json:"settingKey" binding:"required"`
	SettingValue *string   `gorm:"type:text;column:setting_value" json:"settingValue"`
	SettingType  string    `gorm:"type:enum('text','number','boolean','json');not null;default:'text';column:setting_type" json:"settingType"`
	Description  *string   `gorm:"type:text;column:description" json:"description"`
	UpdatedBy    *uint     `gorm:"column:updated_by" json:"updatedBy"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updatedAt"`

	// Relationships
	Updater *User `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

func (InvoiceSetting) TableName() string {
	return "invoice_settings"
}

// InvoicePayment represents payments made against an invoice
type InvoicePayment struct {
	PaymentID       uint64    `gorm:"primaryKey;autoIncrement;column:payment_id" json:"paymentId"`
	InvoiceID       uint64    `gorm:"not null;column:invoice_id" json:"invoiceId"`
	Amount          float64   `gorm:"type:decimal(12,2);not null;column:amount" json:"amount" binding:"required"`
	PaymentMethod   *string   `gorm:"column:payment_method" json:"paymentMethod"`
	PaymentDate     time.Time `gorm:"type:date;not null;column:payment_date" json:"paymentDate" binding:"required"`
	ReferenceNumber *string   `gorm:"column:reference_number" json:"referenceNumber"`
	Notes           *string   `gorm:"type:text;column:notes" json:"notes"`
	CreatedBy       *uint     `gorm:"column:created_by" json:"createdBy"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"createdAt"`

	// Relationships
	Invoice *Invoice `gorm:"foreignKey:InvoiceID" json:"invoice,omitempty"`
	Creator *User    `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (InvoicePayment) TableName() string {
	return "invoice_payments"
}

// ================================================================
// DTOs and Request/Response Models
// ================================================================

// InvoiceCreateRequest represents the request to create an invoice
type InvoiceCreateRequest struct {
	CustomerID      uint                         `json:"customerId" binding:"required"`
	JobID           *uint                        `json:"jobId"`
	TemplateID      *uint                        `json:"templateId"`
	IssueDate       time.Time                    `json:"issueDate" binding:"required"`
	DueDate         time.Time                    `json:"dueDate" binding:"required"`
	PaymentTerms    *string                      `json:"paymentTerms"`
	TaxRate         float64                      `json:"taxRate"`
	DiscountAmount  float64                      `json:"discountAmount"`
	Notes           *string                      `json:"notes"`
	TermsConditions *string                      `json:"termsConditions"`
	LineItems       []InvoiceLineItemCreateRequest `json:"lineItems" binding:"required,dive"`
}

// InvoiceLineItemCreateRequest represents a line item in the create request
type InvoiceLineItemCreateRequest struct {
	ItemType        string     `json:"itemType" binding:"required,oneof=device service package custom"`
	DeviceID        *string    `json:"deviceId"`
	PackageID       *uint      `json:"packageId"`
	Description     string     `json:"description" binding:"required"`
	Quantity        float64    `json:"quantity" binding:"required,gt=0"`
	UnitPrice       float64    `json:"unitPrice" binding:"required,gte=0"`
	RentalStartDate *time.Time `json:"rentalStartDate"`
	RentalEndDate   *time.Time `json:"rentalEndDate"`
	RentalDays      *int       `json:"rentalDays"`
}

// InvoiceResponse represents the response when returning invoice data
type InvoiceResponse struct {
	Invoice
	Customer         Customer            `json:"customer"`
	Job              *Job                `json:"job,omitempty"`
	Template         *InvoiceTemplate    `json:"template,omitempty"`
	LineItems        []InvoiceLineItem   `json:"lineItems"`
	Payments         []InvoicePayment    `json:"payments"`
	DaysOverdue      int                 `json:"daysOverdue"`
	PaymentStatus    string              `json:"paymentStatus"`
}

// InvoiceListResponse represents paginated invoice list
type InvoiceListResponse struct {
	Invoices   []InvoiceResponse `json:"invoices"`
	TotalCount int64            `json:"totalCount"`
	Page       int              `json:"page"`
	PageSize   int              `json:"pageSize"`
	TotalPages int              `json:"totalPages"`
}

// InvoiceSettings represents all invoice configuration settings
type InvoiceSettings struct {
	InvoiceNumberPrefix     string  `json:"invoiceNumberPrefix"`
	InvoiceNumberFormat     string  `json:"invoiceNumberFormat"`
	DefaultPaymentTerms     int     `json:"defaultPaymentTerms"`
	DefaultTaxRate          float64 `json:"defaultTaxRate"`
	AutoCalculateRentalDays bool    `json:"autoCalculateRentalDays"`
	ShowLogoOnInvoice       bool    `json:"showLogoOnInvoice"`
	CurrencySymbol          string  `json:"currencySymbol"`
	CurrencyCode            string  `json:"currencyCode"`
	DateFormat              string  `json:"dateFormat"`
}

// InvoiceTemplateVariables represents variables available in templates
type InvoiceTemplateVariables struct {
	Company   CompanySettings   `json:"company"`
	Invoice   Invoice          `json:"invoice"`
	Customer  Customer         `json:"customer"`
	Job       *Job             `json:"job,omitempty"`
	LineItems []InvoiceLineItem `json:"lineItems"`
	Settings  InvoiceSettings  `json:"settings"`
	LogoURL   string           `json:"logoUrl"`
}

// InvoiceFilter represents filters for listing invoices
type InvoiceFilter struct {
	Status        string     `form:"status" json:"status"`
	CustomerID    *uint      `form:"customer_id" json:"customerId"`
	JobID         *uint      `form:"job_id" json:"jobId"`
	StartDate     *time.Time `form:"start_date" json:"startDate"`
	EndDate       *time.Time `form:"end_date" json:"endDate"`
	MinAmount     *float64   `form:"min_amount" json:"minAmount"`
	MaxAmount     *float64   `form:"max_amount" json:"maxAmount"`
	OverdueOnly   bool       `form:"overdue_only" json:"overdueOnly"`
	SearchTerm    string     `form:"search" json:"searchTerm"`
	Page          int        `form:"page" json:"page"`
	PageSize      int        `form:"page_size" json:"pageSize"`
}