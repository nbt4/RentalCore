package models

import (
	"time"
	"encoding/json"
)

// ================================================================
// ANALYTICS & TRACKING MODELS
// ================================================================

type EquipmentUsageLog struct {
	LogID            uint      `gorm:"primaryKey;autoIncrement" json:"logID"`
	DeviceID         string    `gorm:"not null" json:"deviceID"`
	JobID            *uint     `json:"jobID"`
	Action           string    `gorm:"type:enum('assigned','returned','maintenance','available');not null" json:"action"`
	Timestamp        time.Time `gorm:"not null" json:"timestamp"`
	DurationHours    *float64  `gorm:"type:decimal(10,2)" json:"durationHours"`
	RevenueGenerated *float64  `gorm:"type:decimal(12,2)" json:"revenueGenerated"`
	Notes            string    `json:"notes"`
	CreatedAt        time.Time `json:"createdAt"`

	// Relationships
	Device *Device `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	Job    *Job    `gorm:"foreignKey:JobID" json:"job,omitempty"`
}

type FinancialTransaction struct {
	TransactionID     uint      `gorm:"primaryKey;autoIncrement" json:"transactionID"`
	JobID             *uint     `json:"jobID"`
	CustomerID        *uint     `json:"customerID"`
	Type              string    `gorm:"type:enum('rental','deposit','payment','refund','fee','discount');not null" json:"type"`
	Amount            float64   `gorm:"type:decimal(12,2);not null" json:"amount"`
	Currency          string    `gorm:"default:'EUR'" json:"currency"`
	Status            string    `gorm:"type:enum('pending','completed','failed','cancelled');not null" json:"status"`
	PaymentMethod     string    `json:"paymentMethod"`
	TransactionDate   time.Time `gorm:"not null" json:"transactionDate"`
	DueDate           *time.Time `json:"dueDate"`
	ReferenceNumber   string    `json:"referenceNumber"`
	Notes             string    `json:"notes"`
	CreatedBy         *uint     `json:"createdBy"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`

	// Relationships
	Job      *Job      `gorm:"foreignKey:JobID" json:"job,omitempty"`
	Customer *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Creator  *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

type AnalyticsCache struct {
	CacheID    uint            `gorm:"primaryKey;autoIncrement" json:"cacheID"`
	MetricName string          `gorm:"not null" json:"metricName"`
	PeriodType string          `gorm:"type:enum('daily','weekly','monthly','yearly');not null" json:"periodType"`
	PeriodDate time.Time       `gorm:"not null" json:"periodDate"`
	Value      *float64        `gorm:"type:decimal(15,4)" json:"value"`
	Metadata   json.RawMessage `gorm:"type:json" json:"metadata"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

// ================================================================
// DOCUMENT MANAGEMENT MODELS
// ================================================================

type Document struct {
	DocumentID       uint      `gorm:"primaryKey;autoIncrement" json:"documentID"`
	EntityType       string    `gorm:"type:enum('job','device','customer','user','system');not null" json:"entityType"`
	EntityID         string    `gorm:"not null" json:"entityID"`
	Filename         string    `gorm:"not null" json:"filename"`
	OriginalFilename string    `gorm:"not null" json:"originalFilename"`
	FilePath         string    `gorm:"not null" json:"filePath"`
	FileSize         int64     `gorm:"not null" json:"fileSize"`
	MimeType         string    `gorm:"not null" json:"mimeType"`
	DocumentType     string    `gorm:"type:enum('contract','manual','photo','invoice','receipt','signature','other');not null" json:"documentType"`
	Description      string    `json:"description"`
	UploadedBy       *uint     `json:"uploadedBy"`
	UploadedAt       time.Time `json:"uploadedAt"`
	IsPublic         bool      `gorm:"default:false" json:"isPublic"`
	Version          int       `gorm:"default:1" json:"version"`
	ParentDocumentID *uint     `json:"parentDocumentID"`
	Checksum         string    `json:"checksum"`

	// Relationships
	Uploader       *User               `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
	ParentDocument *Document           `gorm:"foreignKey:ParentDocumentID" json:"parentDocument,omitempty"`
	Signatures     []DigitalSignature  `gorm:"foreignKey:DocumentID" json:"signatures,omitempty"`
}

type DigitalSignature struct {
	SignatureID      uint      `gorm:"primaryKey;autoIncrement" json:"signatureID"`
	DocumentID       uint      `gorm:"not null" json:"documentID"`
	SignerName       string    `gorm:"not null" json:"signerName"`
	SignerEmail      string    `json:"signerEmail"`
	SignerRole       string    `json:"signerRole"`
	SignatureData    string    `gorm:"type:longtext;not null" json:"signatureData"`
	SignedAt         time.Time `json:"signedAt"`
	IPAddress        string    `json:"ipAddress"`
	VerificationCode string    `json:"verificationCode"`
	IsVerified       bool      `gorm:"default:false" json:"isVerified"`

	// Relationships
	Document *Document `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
}

// ================================================================
// SEARCH & FILTERS MODELS
// ================================================================

type SavedSearch struct {
	SearchID   uint            `gorm:"primaryKey;autoIncrement" json:"searchID"`
	UserID     uint            `gorm:"not null" json:"userID"`
	Name       string          `gorm:"not null" json:"name"`
	SearchType string          `gorm:"type:enum('global','jobs','devices','customers','cases');not null" json:"searchType"`
	Filters    json.RawMessage `gorm:"type:json;not null" json:"filters"`
	IsDefault  bool            `gorm:"default:false" json:"isDefault"`
	IsPublic   bool            `gorm:"default:false" json:"isPublic"`
	UsageCount int             `gorm:"default:0" json:"usageCount"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
	LastUsed   *time.Time      `json:"lastUsed"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type SearchHistory struct {
	HistoryID       uint            `gorm:"primaryKey;autoIncrement" json:"historyID"`
	UserID          *uint           `json:"userID"`
	SearchTerm      string          `json:"searchTerm"`
	SearchType      string          `json:"searchType"`
	Filters         json.RawMessage `gorm:"type:json" json:"filters"`
	ResultsCount    int             `json:"resultsCount"`
	ExecutionTimeMS int             `json:"executionTimeMS"`
	SearchedAt      time.Time       `json:"searchedAt"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ================================================================
// WORKFLOW & TEMPLATES MODELS
// ================================================================

type JobTemplate struct {
	TemplateID          uint            `gorm:"primaryKey;autoIncrement" json:"templateID"`
	Name                string          `gorm:"not null" json:"name"`
	Description         string          `json:"description"`
	JobCategoryID       *uint           `json:"jobCategoryID"`
	DefaultDurationDays int             `json:"defaultDurationDays"`
	EquipmentList       json.RawMessage `gorm:"type:json" json:"equipmentList"`
	DefaultNotes        string          `json:"defaultNotes"`
	PricingTemplate     json.RawMessage `gorm:"type:json" json:"pricingTemplate"`
	RequiredDocuments   json.RawMessage `gorm:"type:json" json:"requiredDocuments"`
	CreatedBy           *uint           `json:"createdBy"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
	IsActive            bool            `gorm:"default:true" json:"isActive"`
	UsageCount          int             `gorm:"default:0" json:"usageCount"`

	// Relationships
	JobCategory *JobCategory `gorm:"foreignKey:JobCategoryID" json:"jobCategory,omitempty"`
	Creator     *User        `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Jobs        []Job        `gorm:"foreignKey:TemplateID" json:"jobs,omitempty"`
}

type EquipmentPackage struct {
	PackageID        uint            `gorm:"primaryKey;autoIncrement" json:"packageID"`
	Name             string          `gorm:"not null" json:"name"`
	Description      string          `json:"description"`
	PackageItems     json.RawMessage `gorm:"type:json;not null" json:"packageItems"`
	PackagePrice     *float64        `gorm:"type:decimal(12,2)" json:"packagePrice"`
	DiscountPercent  float64         `gorm:"type:decimal(5,2);default:0.00" json:"discountPercent"`
	MinRentalDays    int             `gorm:"default:1" json:"minRentalDays"`
	IsActive         bool            `gorm:"default:true" json:"isActive"`
	CreatedBy        *uint           `json:"createdBy"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
	UsageCount       int             `gorm:"default:0" json:"usageCount"`

	// Relationships
	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// ================================================================
// SECURITY & PERMISSIONS MODELS
// ================================================================

type Role struct {
	RoleID       uint            `gorm:"primaryKey;autoIncrement" json:"roleID"`
	Name         string          `gorm:"uniqueIndex;not null" json:"name"`
	DisplayName  string          `gorm:"not null" json:"displayName"`
	Description  string          `json:"description"`
	Permissions  json.RawMessage `gorm:"type:json;not null" json:"permissions"`
	IsSystemRole bool            `gorm:"default:false" json:"isSystemRole"`
	IsActive     bool            `gorm:"default:true" json:"isActive"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`

	// Relationships
	UserRoles []UserRole `gorm:"foreignKey:RoleID" json:"userRoles,omitempty"`
}

type UserRole struct {
	UserID     uint       `gorm:"primaryKey" json:"userID"`
	RoleID     uint       `gorm:"primaryKey" json:"roleID"`
	AssignedAt time.Time  `json:"assignedAt"`
	AssignedBy *uint      `json:"assignedBy"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	IsActive   bool       `gorm:"default:true" json:"isActive"`

	// Relationships
	User     *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role     *Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Assigner *User `gorm:"foreignKey:AssignedBy" json:"assigner,omitempty"`
}

type AuditLog struct {
	AuditID    uint            `gorm:"primaryKey;autoIncrement" json:"auditID"`
	UserID     *uint           `json:"userID"`
	Action     string          `gorm:"not null" json:"action"`
	EntityType string          `gorm:"not null" json:"entityType"`
	EntityID   string          `gorm:"not null" json:"entityID"`
	OldValues  json.RawMessage `gorm:"type:json" json:"oldValues"`
	NewValues  json.RawMessage `gorm:"type:json" json:"newValues"`
	IPAddress  string          `json:"ipAddress"`
	UserAgent  string          `json:"userAgent"`
	SessionID  string          `json:"sessionID"`
	Timestamp  time.Time       `json:"timestamp"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ================================================================
// MOBILE & PWA MODELS
// ================================================================

type PushSubscription struct {
	SubscriptionID uint      `gorm:"primaryKey;autoIncrement" json:"subscriptionID"`
	UserID         uint      `gorm:"not null" json:"userID"`
	Endpoint       string    `gorm:"type:text;not null" json:"endpoint"`
	KeysP256dh     string    `gorm:"type:text;not null" json:"keysP256dh"`
	KeysAuth       string    `gorm:"type:text;not null" json:"keysAuth"`
	UserAgent      string    `gorm:"type:text" json:"userAgent"`
	DeviceType     string    `json:"deviceType"`
	CreatedAt      time.Time `json:"createdAt"`
	LastUsed       time.Time `json:"lastUsed"`
	IsActive       bool      `gorm:"default:true" json:"isActive"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type OfflineSyncQueue struct {
	QueueID      uint            `gorm:"primaryKey;autoIncrement" json:"queueID"`
	UserID       uint            `gorm:"not null" json:"userID"`
	Action       string          `gorm:"type:enum('create','update','delete');not null" json:"action"`
	EntityType   string          `gorm:"not null" json:"entityType"`
	EntityData   json.RawMessage `gorm:"type:json;not null" json:"entityData"`
	Timestamp    time.Time       `json:"timestamp"`
	Synced       bool            `gorm:"default:false" json:"synced"`
	SyncedAt     *time.Time      `json:"syncedAt"`
	RetryCount   int             `gorm:"default:0" json:"retryCount"`
	ErrorMessage string          `json:"errorMessage"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ================================================================
// ENHANCED EXISTING MODELS (EXTENSIONS)
// ================================================================

// UserEnhanced extends the existing User model with new fields
type UserEnhanced struct {
	User                     // Embed the existing User struct
	Timezone                 string          `gorm:"default:'Europe/Berlin'" json:"timezone"`
	Language                 string          `gorm:"default:'en'" json:"language"`
	AvatarPath               string          `json:"avatarPath"`
	NotificationPreferences  json.RawMessage `gorm:"type:json" json:"notificationPreferences"`
	LastActive               *time.Time      `json:"lastActive"`
	LoginAttempts            int             `gorm:"default:0" json:"loginAttempts"`
	LockedUntil              *time.Time      `json:"lockedUntil"`
	TwoFactorEnabled         bool            `gorm:"default:false" json:"twoFactorEnabled"`
	TwoFactorSecret          string          `json:"twoFactorSecret,omitempty"`

	// New relationships
	UserRoles         []UserRole          `gorm:"foreignKey:UserID" json:"userRoles,omitempty"`
	PushSubscriptions []PushSubscription  `gorm:"foreignKey:UserID" json:"pushSubscriptions,omitempty"`
	SavedSearches     []SavedSearch       `gorm:"foreignKey:UserID" json:"savedSearches,omitempty"`
	OfflineSyncQueue  []OfflineSyncQueue  `gorm:"foreignKey:UserID" json:"offlineSyncQueue,omitempty"`
}

// JobEnhanced extends the existing Job model with new fields
type JobEnhanced struct {
	Job                      // Embed the existing Job struct
	TemplateID               *uint    `json:"templateID"`
	Priority                 string   `gorm:"type:enum('low','normal','high','urgent');default:'normal'" json:"priority"`
	InternalNotes            string   `json:"internalNotes"`
	CustomerNotes            string   `json:"customerNotes"`
	EstimatedRevenue         *float64 `gorm:"type:decimal(12,2)" json:"estimatedRevenue"`
	ActualCost               float64  `gorm:"type:decimal(12,2);default:0.00" json:"actualCost"`
	ProfitMargin             *float64 `gorm:"type:decimal(5,2)" json:"profitMargin"`
	ContractSigned           bool     `gorm:"default:false" json:"contractSigned"`
	ContractDocumentID       *uint    `json:"contractDocumentID"`
	CompletionPercentage     int      `gorm:"default:0" json:"completionPercentage"`

	// New relationships
	Template         *JobTemplate          `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	ContractDocument *Document             `gorm:"foreignKey:ContractDocumentID" json:"contractDocument,omitempty"`
	UsageLogs        []EquipmentUsageLog   `gorm:"foreignKey:JobID" json:"usageLogs,omitempty"`
	Transactions     []FinancialTransaction `gorm:"foreignKey:JobID" json:"transactions,omitempty"`
	Documents        []Document            `gorm:"foreignKey:EntityID;where:entity_type = 'job'" json:"documents,omitempty"`
}

// DeviceEnhanced extends the existing Device model with new fields
type DeviceEnhanced struct {
	Device                   // Embed the existing Device struct
	QRCode                   string   `gorm:"uniqueIndex" json:"qrCode"`
	CurrentLocation          string   `json:"currentLocation"`
	GPSLatitude              *float64 `gorm:"type:decimal(10,8)" json:"gpsLatitude"`
	GPSLongitude             *float64 `gorm:"type:decimal(11,8)" json:"gpsLongitude"`
	ConditionRating          float64  `gorm:"type:decimal(3,1);default:5.0" json:"conditionRating"`
	UsageHours               float64  `gorm:"type:decimal(10,2);default:0.00" json:"usageHours"`
	TotalRevenue             float64  `gorm:"type:decimal(12,2);default:0.00" json:"totalRevenue"`
	LastMaintenanceCost      *float64 `gorm:"type:decimal(10,2)" json:"lastMaintenanceCost"`
	Notes                    string   `json:"notes"`
	Barcode                  string   `json:"barcode"`

	// New relationships
	UsageLogs []EquipmentUsageLog `gorm:"foreignKey:DeviceID" json:"usageLogs,omitempty"`
	Documents []Document          `gorm:"foreignKey:EntityID;where:entity_type = 'device'" json:"documents,omitempty"`
}

// CustomerEnhanced extends the existing Customer model with new fields
type CustomerEnhanced struct {
	Customer                 // Embed the existing Customer struct
	TaxNumber                string   `json:"taxNumber"`
	CreditLimit              float64  `gorm:"type:decimal(12,2);default:0.00" json:"creditLimit"`
	PaymentTerms             int      `gorm:"default:30" json:"paymentTerms"`
	PreferredPaymentMethod   string   `json:"preferredPaymentMethod"`
	CustomerSince            *time.Time `json:"customerSince"`
	TotalLifetimeValue       float64  `gorm:"type:decimal(12,2);default:0.00" json:"totalLifetimeValue"`
	LastJobDate              *time.Time `json:"lastJobDate"`
	Rating                   float64  `gorm:"type:decimal(3,1);default:5.0" json:"rating"`
	BillingAddress           string   `json:"billingAddress"`
	ShippingAddress          string   `json:"shippingAddress"`

	// New relationships
	Transactions []FinancialTransaction `gorm:"foreignKey:CustomerID" json:"transactions,omitempty"`
	Documents    []Document             `gorm:"foreignKey:EntityID;where:entity_type = 'customer'" json:"documents,omitempty"`
}

// ================================================================
// ANALYTICS VIEW MODELS
// ================================================================

type EquipmentUtilization struct {
	DeviceID        string  `json:"deviceID"`
	ProductName     string  `json:"productName"`
	Status          string  `json:"status"`
	UsageHours      float64 `json:"usageHours"`
	TotalRevenue    float64 `json:"totalRevenue"`
	RevenuePerHour  float64 `json:"revenuePerHour"`
	TimesRented     int     `json:"timesRented"`
	ConditionRating float64 `json:"conditionRating"`
	LastMaintenance *time.Time `json:"lastMaintenance"`
}

type CustomerPerformance struct {
	CustomerID      uint       `json:"customerID"`
	CompanyName     string     `json:"companyName"`
	TotalLifetimeValue float64 `json:"totalLifetimeValue"`
	Rating          float64    `json:"rating"`
	CustomerSince   *time.Time `json:"customerSince"`
	TotalJobs       int        `json:"totalJobs"`
	TotalRevenue    float64    `json:"totalRevenue"`
	LastJobDate     *time.Time `json:"lastJobDate"`
	AvgRentalDays   float64    `json:"avgRentalDays"`
}

type MonthlyRevenue struct {
	Year            int     `json:"year"`
	Month           int     `json:"month"`
	TotalJobs       int     `json:"totalJobs"`
	TotalRevenue    float64 `json:"totalRevenue"`
	AvgJobValue     float64 `json:"avgJobValue"`
	UniqueCustomers int     `json:"uniqueCustomers"`
}

// ================================================================
// REQUEST/RESPONSE DTOs
// ================================================================

type AnalyticsRequest struct {
	Period    string    `json:"period"`    // daily, weekly, monthly, yearly
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Metrics   []string  `json:"metrics"`   // revenue, utilization, customers, etc.
}

type SearchRequest struct {
	Query      string                 `json:"query"`
	Type       string                 `json:"type"`       // global, jobs, devices, customers, cases
	Filters    map[string]interface{} `json:"filters"`
	Sort       string                 `json:"sort"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"pageSize"`
	SaveSearch bool                   `json:"saveSearch"`
	SearchName string                 `json:"searchName"`
}

type BulkActionRequest struct {
	Action   string   `json:"action"`
	EntityIDs []string `json:"entityIds"`
	Data     map[string]interface{} `json:"data"`
}