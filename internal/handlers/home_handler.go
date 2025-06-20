package handlers

import (
	"net/http"

	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HomeHandler struct {
	jobRepo      *repository.JobRepository
	deviceRepo   *repository.DeviceRepository
	customerRepo *repository.CustomerRepository
	caseRepo     *repository.CaseRepository
	db           *gorm.DB
}

func NewHomeHandler(jobRepo *repository.JobRepository, deviceRepo *repository.DeviceRepository, customerRepo *repository.CustomerRepository, caseRepo *repository.CaseRepository, db *gorm.DB) *HomeHandler {
	return &HomeHandler{
		jobRepo:      jobRepo,
		deviceRepo:   deviceRepo,
		customerRepo: customerRepo,
		caseRepo:     caseRepo,
		db:           db,
	}
}

func (h *HomeHandler) Dashboard(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	// Get real counts from database using direct queries
	var totalJobs int64
	var totalDevices int64 
	var totalCustomers int64
	var totalCases int64
	
	// Use the DB connection to count records
	h.db.Model(&models.Job{}).Count(&totalJobs)
	h.db.Model(&models.Device{}).Count(&totalDevices)
	h.db.Model(&models.Customer{}).Count(&totalCustomers)
	h.db.Model(&models.Case{}).Count(&totalCases)
	
	stats := gin.H{
		"TotalJobs":      totalJobs,
		"TotalDevices":   totalDevices,  
		"TotalCustomers": totalCustomers,
		"TotalCases":     totalCases,
	}
	
	// Get recent jobs (limit to 5 for performance)
	recentJobs, _ := h.jobRepo.List(&models.FilterParams{
		Limit: 5,
	})
	
	c.HTML(http.StatusOK, "home.html", gin.H{
		"title":      "Home",
		"user":       user,
		"stats":      stats,
		"recentJobs": recentJobs,
	})
}