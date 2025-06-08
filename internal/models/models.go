package models

import (
	"time"
)

type Customer struct {
	CustomerID   uint      `json:"customerID" gorm:"primaryKey;column:customerID"`
	CompanyName  *string   `json:"companyname" gorm:"column:companyname"`
	LastName     *string   `json:"lastname" gorm:"column:lastname"`
	FirstName    *string   `json:"firstname" gorm:"column:firstname"`
	Street       *string   `json:"street" gorm:"column:street"`
	HouseNumber  *string   `json:"housenumber" gorm:"column:housenumber"`
	ZIP          *string   `json:"ZIP" gorm:"column:ZIP"`
	City         *string   `json:"city" gorm:"column:city"`
	FederalState *string   `json:"federalstate" gorm:"column:federalstate"`
	Country      *string   `json:"country" gorm:"column:country"`
	PhoneNumber  *string   `json:"phonenumber" gorm:"column:phonenumber"`
	Email        *string   `json:"email" gorm:"column:email"`
	CustomerType *string   `json:"customertype" gorm:"column:customertype"`
	Notes        *string   `json:"notes" gorm:"column:notes"`
	Jobs         []Job     `json:"jobs,omitempty" gorm:"foreignKey:CustomerID"`
}

func (Customer) TableName() string {
	return "customers"
}

func (c Customer) GetDisplayName() string {
	if c.CompanyName != nil && *c.CompanyName != "" {
		return *c.CompanyName
	}
	if c.FirstName != nil && c.LastName != nil && *c.FirstName != "" && *c.LastName != "" {
		return *c.FirstName + " " + *c.LastName
	}
	if c.LastName != nil && *c.LastName != "" {
		return *c.LastName
	}
	if c.FirstName != nil && *c.FirstName != "" {
		return *c.FirstName
	}
	return "Unknown Customer"
}

type Status struct {
	StatusID uint   `json:"statusID" gorm:"primaryKey;column:statusID"`
	Status   string `json:"status" gorm:"not null;column:status"`
	Jobs     []Job  `json:"jobs,omitempty" gorm:"foreignKey:StatusID"`
}

func (Status) TableName() string {
	return "status"
}

type Job struct {
	JobID           uint        `json:"jobID" gorm:"primaryKey;column:jobID"`
	CustomerID      uint        `json:"customerID" gorm:"not null;column:customerID"`
	Customer        Customer    `json:"customer,omitempty" gorm:"foreignKey:CustomerID;references:CustomerID"`
	StatusID        uint        `json:"statusID" gorm:"not null;column:statusID"`
	Status          Status      `json:"status,omitempty" gorm:"foreignKey:StatusID;references:StatusID"`
	JobCategoryID   *uint       `json:"jobcategoryID" gorm:"column:jobcategoryID"`
	Description     *string     `json:"description" gorm:"column:description"`
	Discount        float64     `json:"discount" gorm:"column:discount;default:0"`
	DiscountType    string      `json:"discount_type" gorm:"column:discount_type;default:amount"`
	Revenue         float64     `json:"revenue" gorm:"column:revenue;default:0"`
	FinalRevenue    *float64    `json:"final_revenue" gorm:"column:final_revenue"`
	StartDate       *time.Time  `json:"startDate" gorm:"column:startDate;type:date"`
	EndDate         *time.Time  `json:"endDate" gorm:"column:endDate;type:date"`
	JobDevices      []JobDevice `json:"job_devices,omitempty" gorm:"foreignKey:JobID;references:JobID"`
}

func (Job) TableName() string {
	return "jobs"
}

type Device struct {
	DeviceID         string      `json:"deviceID" gorm:"primaryKey;column:deviceID"`
	ProductID        *uint       `json:"productID" gorm:"column:productID"`
	Product          *Product    `json:"product,omitempty" gorm:"foreignKey:ProductID;references:ProductID"`
	SerialNumber     *string     `json:"serialnumber" gorm:"column:serialnumber"`
	PurchaseDate     *time.Time  `json:"purchaseDate" gorm:"column:purchaseDate;type:date"`
	LastMaintenance  *time.Time  `json:"lastmaintenance" gorm:"column:lastmaintenance;type:date"`
	NextMaintenance  *time.Time  `json:"nextmaintenance" gorm:"column:nextmaintenance;type:date"`
	InsuranceNumber  *string     `json:"insurancenumber" gorm:"column:insurancenumber"`
	Status           string      `json:"status" gorm:"column:status;default:free"`
	InsuranceID      *uint       `json:"insuranceID" gorm:"column:insuranceID"`
	JobDevices       []JobDevice `json:"job_devices,omitempty" gorm:"foreignKey:DeviceID;references:DeviceID"`
}

func (Device) TableName() string {
	return "devices"
}

type Product struct {
	ProductID             uint     `json:"productID" gorm:"primaryKey;column:productID"`
	Name                  string   `json:"name" gorm:"not null;column:name"`
	CategoryID            *uint    `json:"categoryID" gorm:"column:categoryID"`
	SubcategoryID         *string  `json:"subcategoryID" gorm:"column:subcategoryID"`
	SubbiercategoryID     *string  `json:"subbiercategoryID" gorm:"column:subbiercategoryID"`
	ManufacturerID        *uint    `json:"manufacturerID" gorm:"column:manufacturerID"`
	BrandID               *uint    `json:"brandID" gorm:"column:brandID"`
	Description           *string  `json:"description" gorm:"column:description"`
	MaintenanceInterval   *uint    `json:"maintenanceInterval" gorm:"column:maintenanceInterval"`
	ItemCostPerDay        *float64 `json:"itemcostperday" gorm:"column:itemcostperday"`
	Weight                *float64 `json:"weight" gorm:"column:weight"`
	Height                *float64 `json:"height" gorm:"column:height"`
	Width                 *float64 `json:"width" gorm:"column:width"`
	Depth                 *float64 `json:"depth" gorm:"column:depth"`
	PowerConsumption      *float64 `json:"powerconsumption" gorm:"column:powerconsumption"`
	PosInCategory         *uint    `json:"pos_in_category" gorm:"column:pos_in_category"`
}

func (Product) TableName() string {
	return "products"
}

type JobDevice struct {
	JobID       uint     `json:"jobID" gorm:"primaryKey;column:jobID"`
	DeviceID    string   `json:"deviceID" gorm:"primaryKey;column:deviceID"`
	Job         Job      `json:"job,omitempty" gorm:"foreignKey:JobID;references:JobID"`
	Device      Device   `json:"device,omitempty" gorm:"foreignKey:DeviceID;references:DeviceID"`
	CustomPrice *float64 `json:"custom_price" gorm:"column:custom_price"`
}

func (JobDevice) TableName() string {
	return "jobdevices"
}

// JobWithDetails represents a job with aggregated information
type JobWithDetails struct {
	JobID        uint       `json:"jobID"`
	CustomerID   uint       `json:"customerID"`
	StatusID     uint       `json:"statusID"`
	Description  *string    `json:"description"`
	StartDate    *time.Time `json:"startDate"`
	EndDate      *time.Time `json:"endDate"`
	Revenue      float64    `json:"revenue"`
	FinalRevenue *float64   `json:"final_revenue"`
	CustomerName string     `json:"customer_name"`
	StatusName   string     `json:"status_name"`
	DeviceCount  int        `json:"device_count"`
	TotalRevenue float64    `json:"total_revenue"`
}

// DeviceWithJobInfo represents a device with its current job assignment
type DeviceWithJobInfo struct {
	Device
	JobID      *uint   `json:"job_id"`
	JobTitle   *string `json:"job_title"`
	IsAssigned bool    `json:"is_assigned"`
}

// BulkScanRequest represents a request for bulk device scanning
type BulkScanRequest struct {
	JobID     uint     `json:"job_id" binding:"required"`
	DeviceIDs []string `json:"device_ids" binding:"required"`
	Price     float64  `json:"price"`
}

// ScanResult represents the result of a device scan operation
type ScanResult struct {
	DeviceID string  `json:"device_id"`
	Success  bool    `json:"success"`
	Message  string  `json:"message"`
	Device   *Device `json:"device,omitempty"`
}

// Additional models matching your database schema

type JobCategory struct {
	JobCategoryID uint    `json:"jobcategoryID" gorm:"primaryKey;column:jobcategoryID"`
	Name          string  `json:"name" gorm:"column:name"`
	Abbreviation  *string `json:"abbreviation" gorm:"column:abbreviation"`
}

func (JobCategory) TableName() string {
	return "jobCategory"
}

type Category struct {
	CategoryID   uint    `json:"categoryID" gorm:"primaryKey;column:categoryID"`
	Name         string  `json:"name" gorm:"column:name"`
	Abbreviation string  `json:"abbreviation" gorm:"column:abbreviation"`
}

func (Category) TableName() string {
	return "categories"
}

type Brand struct {
	BrandID        uint    `json:"brandID" gorm:"primaryKey;column:brandID"`
	Name           string  `json:"name" gorm:"column:name"`
	ManufacturerID *uint   `json:"manufacturerID" gorm:"column:manufacturerID"`
}

func (Brand) TableName() string {
	return "brands"
}

type Manufacturer struct {
	ManufacturerID uint    `json:"manufacturerID" gorm:"primaryKey;column:manufacturerID"`
	Name           string  `json:"name" gorm:"column:name"`
	Website        *string `json:"website" gorm:"column:website"`
}

func (Manufacturer) TableName() string {
	return "manufacturer"
}

// FilterParams represents parameters for filtering jobs and devices
type FilterParams struct {
	StartDate    *time.Time `form:"start_date"`
	EndDate      *time.Time `form:"end_date"`
	CustomerID   *uint      `form:"customer_id"`
	StatusID     *uint      `form:"status_id"`
	MinRevenue   *float64   `form:"min_revenue"`
	MaxRevenue   *float64   `form:"max_revenue"`
	SearchTerm   string     `form:"search"`
	Category     string     `form:"category"`
	Available    *bool      `form:"available"`
	Limit        int        `form:"limit"`
	Offset       int        `form:"offset"`
}