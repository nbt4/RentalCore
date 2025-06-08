package handlers

import (
	"net/http"
	"strconv"

	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/repository"

	"github.com/gin-gonic/gin"
)

type ScannerHandler struct {
	deviceRepo   *repository.DeviceRepository
	jobRepo      *repository.JobRepository
	customerRepo *repository.CustomerRepository
}

func NewScannerHandler(jobRepo *repository.JobRepository, deviceRepo *repository.DeviceRepository, customerRepo *repository.CustomerRepository) *ScannerHandler {
	return &ScannerHandler{
		deviceRepo:   deviceRepo,
		jobRepo:      jobRepo,
		customerRepo: customerRepo,
	}
}

func (h *ScannerHandler) ScanJobSelection(c *gin.Context) {
	// Get active jobs for selection
	jobs, err := h.jobRepo.List(&models.FilterParams{})
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	// Get device statistics for the dashboard
	totalDevices, err := h.deviceRepo.List(&models.FilterParams{})
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	// Count available devices (status = 'free')
	availableCount := 0
	assignedCount := 0
	for _, device := range totalDevices {
		if device.Status == "free" {
			availableCount++
		} else {
			assignedCount++
		}
	}

	c.HTML(http.StatusOK, "scan_select_job.html", gin.H{
		"title":           "Select Job for Scanning",
		"jobs":            jobs,
		"totalDevices":    availableCount,
		"assignedDevices": assignedCount,
	})
}

func (h *ScannerHandler) ScanJob(c *gin.Context) {
	jobID, err := strconv.ParseUint(c.Param("jobId"), 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid job ID"})
		return
	}

	job, err := h.jobRepo.GetByID(uint(jobID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Job not found"})
		return
	}

	// Get assigned devices for this job
	assignedDevices, err := h.jobRepo.GetJobDevices(uint(jobID))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "scan_job.html", gin.H{
		"title":           "Scanning Job #" + strconv.FormatUint(jobID, 10),
		"job":             job,
		"assignedDevices": assignedDevices,
	})
}

type ScanDeviceRequest struct {
	JobID    uint   `json:"job_id" binding:"required"`
	DeviceID string `json:"device_id" binding:"required"`
	Price    float64 `json:"price"`
}

func (h *ScannerHandler) ScanDevice(c *gin.Context) {
	var req ScanDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Try to get device by ID first, then by serial number
	var device *models.Device
	var err error

	device, err = h.deviceRepo.GetByID(req.DeviceID)
	if err != nil {
		// Try by serial number
		device, err = h.deviceRepo.GetBySerialNo(req.DeviceID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
			return
		}
	}

	// Check if device is available (status should be 'free')
	if device.Status != "free" {
		c.JSON(http.StatusConflict, gin.H{
			"error":  "Device is not available",
			"device": device,
		})
		return
	}

	// Assign device to job
	price := req.Price
	// Note: device price is stored in product table, not device table in your schema

	if err := h.jobRepo.AssignDevice(req.JobID, device.DeviceID, price); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Device successfully assigned to job",
		"device":  device,
		"price":   price,
	})
}

func (h *ScannerHandler) RemoveDevice(c *gin.Context) {
	jobID, err := strconv.ParseUint(c.Param("jobId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	deviceID := c.Param("deviceId")

	if err := h.jobRepo.RemoveDevice(uint(jobID), deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device removed from job successfully"})
}