package handlers

import (
	"fmt"
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
	caseRepo     *repository.CaseRepository
}

func NewScannerHandler(jobRepo *repository.JobRepository, deviceRepo *repository.DeviceRepository, customerRepo *repository.CustomerRepository, caseRepo *repository.CaseRepository) *ScannerHandler {
	return &ScannerHandler{
		deviceRepo:   deviceRepo,
		jobRepo:      jobRepo,
		customerRepo: customerRepo,
		caseRepo:     caseRepo,
	}
}

func (h *ScannerHandler) ScanJobSelection(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	// Get active jobs for selection (only Open status)
	jobs, err := h.jobRepo.List(&models.FilterParams{
		Status: "Open",
	})
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error(), "user": user})
		return
	}

	// Get device statistics for the dashboard
	totalDevices, err := h.deviceRepo.List(&models.FilterParams{})
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error(), "user": user})
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
		"user":            user,
	})
}

func (h *ScannerHandler) ScanJob(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	jobID, err := strconv.ParseUint(c.Param("jobId"), 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid job ID", "user": user})
		return
	}

	job, err := h.jobRepo.GetByID(uint(jobID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Job not found", "user": user})
		return
	}

	// Get assigned devices for this job
	assignedDevices, err := h.jobRepo.GetJobDevices(uint(jobID))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error(), "user": user})
		return
	}

	// Group devices by product
	productGroups := make(map[string]*ProductGroup)
	totalDevices := len(assignedDevices)

	for _, jd := range assignedDevices {
		var productName string
		if jd.Device.Product != nil {
			productName = jd.Device.Product.Name
		} else {
			productName = "Unknown Product"
		}

		if _, exists := productGroups[productName]; !exists {
			productGroups[productName] = &ProductGroup{
				Product: jd.Device.Product,
				Devices: []models.JobDevice{},
			}
		}
		productGroups[productName].Devices = append(productGroups[productName].Devices, jd)
	}

	// Get available cases for case scanning functionality
	cases, err := h.caseRepo.List(&models.FilterParams{})
	if err != nil {
		// If we can't get cases, continue without them - don't fail the page
		cases = []models.Case{}
	}

	c.HTML(http.StatusOK, "scan_job.html", gin.H{
		"title":           "Scanning Job #" + strconv.FormatUint(jobID, 10),
		"job":             job,
		"assignedDevices": assignedDevices,
		"productGroups":   productGroups,
		"totalDevices":    totalDevices,
		"cases":           cases,
		"user":            user,
	})
}

type ScanDeviceRequest struct {
	JobID    uint     `json:"job_id" binding:"required"`
	DeviceID string   `json:"device_id" binding:"required"`
	Price    *float64 `json:"price"`
}

type ScanCaseRequest struct {
	JobID  uint `json:"job_id" binding:"required"`
	CaseID uint `json:"case_id" binding:"required"`
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
	var price float64
	// Only use custom price if explicitly provided, otherwise pass 0 (which means NULL in DB)
	if req.Price != nil {
		price = *req.Price
	} else {
		price = 0.0 // This will result in NULL custom_price in database
	}

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

func (h *ScannerHandler) ScanCase(c *gin.Context) {
	var req ScanCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the case and its devices
	case_, err := h.caseRepo.GetByID(req.CaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	// Get all devices in the case
	devicesInCase, err := h.caseRepo.GetDevicesInCase(req.CaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get devices in case"})
		return
	}

	if len(devicesInCase) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Case is empty - no devices to assign"})
		return
	}

	// Track results
	var results []map[string]interface{}
	successCount := 0
	errorCount := 0

	// Assign all devices in the case to the job
	for _, deviceCase := range devicesInCase {
		device := deviceCase.Device
		
		// Check if device is available
		if device.Status != "free" {
			results = append(results, map[string]interface{}{
				"device_id": device.DeviceID,
				"success":   false,
				"message":   "Device is not available (status: " + device.Status + ")",
			})
			errorCount++
			continue
		}

		// Assign device to job using default pricing (no custom price for case scanning)
		if err := h.jobRepo.AssignDevice(req.JobID, device.DeviceID, 0.0); err != nil {
			results = append(results, map[string]interface{}{
				"device_id": device.DeviceID,
				"success":   false,
				"message":   err.Error(),
			})
			errorCount++
		} else {
			results = append(results, map[string]interface{}{
				"device_id": device.DeviceID,
				"success":   true,
				"message":   "Device assigned successfully",
			})
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       fmt.Sprintf("Case scan complete: %d devices assigned, %d errors", successCount, errorCount),
		"case_id":       req.CaseID,
		"case_name":     case_.Name,
		"total_devices": len(devicesInCase),
		"success_count": successCount,
		"error_count":   errorCount,
		"results":       results,
	})
}