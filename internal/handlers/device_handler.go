package handlers

import (
	"net/http"
	"strconv"
	"log"
	"time"
	"sync"

	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/repository"
	"go-barcode-webapp/internal/services"

	"github.com/gin-gonic/gin"
)

// Simple cache for devices
type DeviceCache struct {
	data      []models.DeviceWithJobInfo
	timestamp time.Time
	mutex     sync.RWMutex
}

var deviceCache = &DeviceCache{
	timestamp: time.Time{}, // Force cache miss initially
}

type DeviceHandler struct {
	deviceRepo     *repository.DeviceRepository
	barcodeService *services.BarcodeService
	productRepo    *repository.ProductRepository
}

func NewDeviceHandler(deviceRepo *repository.DeviceRepository, barcodeService *services.BarcodeService, productRepo *repository.ProductRepository) *DeviceHandler {
	return &DeviceHandler{
		deviceRepo:     deviceRepo,
		barcodeService: barcodeService,
		productRepo:    productRepo,
	}
}


// Web interface handlers
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	startTime := time.Now()
	log.Printf("🚀 DeviceHandler.ListDevices() started")
	
	user, _ := GetCurrentUser(c)
	
	params := &models.FilterParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		log.Printf("❌ Error binding query parameters: %v", err)
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error(), "user": user})
		return
	}
	
	// FIX: Ensure search parameter is properly handled
	searchParam := c.Query("search")
	if searchParam != "" {
		params.SearchTerm = searchParam
	}

	// Handle pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	
	limit := 20 // Reduce items per page for better performance
	params.Limit = limit
	params.Offset = (page - 1) * limit
	params.Page = page

	viewType := c.DefaultQuery("view", "list") // Default to list view

	// Use cache for basic list view without search
	var devices []models.DeviceWithJobInfo
	var err error
	
	if params.SearchTerm == "" && page == 1 {
		// Try to use cache for first page without search
		deviceCache.mutex.RLock()
		if time.Since(deviceCache.timestamp) < 30*time.Second && len(deviceCache.data) > 0 {
			// Use cached data
			devices = deviceCache.data
			if len(devices) > limit {
				devices = devices[:limit]
			}
			deviceCache.mutex.RUnlock()
			log.Printf("⚡ Using cached devices (%d items)", len(devices))
		} else {
			deviceCache.mutex.RUnlock()
			
			// Fetch fresh data
			dbStart := time.Now()
			devices, err = h.deviceRepo.List(params)
			dbTime := time.Since(dbStart)
			log.Printf("⏱️  Database query took: %v", dbTime)
			
			if err != nil {
				log.Printf("❌ Database error: %v", err)
				c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error(), "user": user})
				return
			}
			
			// Cache the result
			deviceCache.mutex.Lock()
			deviceCache.data = devices
			deviceCache.timestamp = time.Now()
			deviceCache.mutex.Unlock()
			log.Printf("💾 Cached %d devices", len(devices))
		}
	} else {
		// For search or pagination, always query database
		dbStart := time.Now()
		devices, err = h.deviceRepo.List(params)
		dbTime := time.Since(dbStart)
		log.Printf("⏱️  Database query took: %v", dbTime)
		
		if err != nil {
			log.Printf("❌ Database error: %v", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error(), "user": user})
			return
		}
	}
	
	templateStart := time.Now()
	if viewType == "categorized" {
		// For categorized view, we still use the same data but render it differently
		c.HTML(http.StatusOK, "devices_standalone.html", gin.H{
			"title":       "Devices by Category",
			"devices":     devices,
			"params":      params,
			"user":        user,
			"viewType":    "categorized",
			"categorized": true,
			"currentPage": page,
			"hasNextPage": len(devices) == limit,
		})
	} else {
		c.HTML(http.StatusOK, "devices_standalone.html", gin.H{
			"title":       "Devices",
			"devices":     devices,
			"params":      params,
			"user":        user,
			"viewType":    "list",
			"categorized": false,
			"currentPage": page,
			"hasNextPage": len(devices) == limit,
		})
	}
	templateTime := time.Since(templateStart)
	totalTime := time.Since(startTime)
	log.Printf("⏱️  Template rendering took: %v", templateTime)
	log.Printf("🏁 DeviceHandler.ListDevices() completed in %v", totalTime)
}

func (h *DeviceHandler) NewDeviceForm(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	products, err := h.productRepo.List(&models.FilterParams{})
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error(), "user": user})
		return
	}

	c.HTML(http.StatusOK, "device_form.html", gin.H{
		"title":    "New Device",
		"device":   &models.Device{},
		"products": products,
		"user":     user,
	})
}

func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	log.Printf("🔥 CREATE DEVICE HANDLER CALLED")
	
	// Get form values
	serialNumber := c.PostForm("serialnumber")
	status := c.PostForm("status")
	notes := c.PostForm("notes")
	
	log.Printf("📝 Form values: serialNumber='%s', status='%s', notes='%s'", serialNumber, status, notes)
	
	if status == "" {
		status = "free"
	}
	
	var productID *uint
	if productIDStr := c.PostForm("productID"); productIDStr != "" {
		if pid, err := strconv.ParseUint(productIDStr, 10, 32); err == nil {
			prodID := uint(pid)
			productID = &prodID
			log.Printf("📝 Product ID: %d", prodID)
		}
	}
	
	if productID == nil {
		log.Printf("❌ No product ID provided")
		user, _ := GetCurrentUser(c)
		products, _ := h.productRepo.List(&models.FilterParams{})
		c.HTML(http.StatusBadRequest, "device_form.html", gin.H{
			"title":    "New Device",
			"device":   &models.Device{},
			"products": products,
			"error":    "Please select a product",
			"user":     user,
		})
		return
	}
	
	device := models.Device{
		DeviceID:     "", // Let database generate the ID automatically
		ProductID:    productID,
		Status:       status,
	}
	
	// Handle optional string fields
	if serialNumber != "" {
		device.SerialNumber = &serialNumber
	}
	if notes != "" {
		device.Notes = &notes
	}
	
	// Handle date fields
	if purchaseDateStr := c.PostForm("purchase_date"); purchaseDateStr != "" {
		if purchaseDate, err := time.Parse("2006-01-02", purchaseDateStr); err == nil {
			device.PurchaseDate = &purchaseDate
		}
	}
	if lastMaintenanceStr := c.PostForm("last_maintenance"); lastMaintenanceStr != "" {
		if lastMaintenance, err := time.Parse("2006-01-02", lastMaintenanceStr); err == nil {
			device.LastMaintenance = &lastMaintenance
		}
	}

	log.Printf("💾 Creating device: %+v", device)
	
	if err := h.deviceRepo.Create(&device); err != nil {
		log.Printf("❌ Error creating device: %v", err)
		user, _ := GetCurrentUser(c)
		products, _ := h.productRepo.List(&models.FilterParams{})
		c.HTML(http.StatusInternalServerError, "device_form.html", gin.H{
			"title":    "New Device",
			"device":   &device,
			"products": products,
			"error":    err.Error(),
			"user":     user,
		})
		return
	}

	log.Printf("✅ Device created successfully")
	c.Redirect(http.StatusFound, "/devices")
}

func (h *DeviceHandler) GetDevice(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	deviceID := c.Param("id")

	device, err := h.deviceRepo.GetByID(deviceID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Device not found", "user": user})
		return
	}

	c.HTML(http.StatusOK, "device_detail.html", gin.H{
		"device": device,
		"user":   user,
	})
}

func (h *DeviceHandler) EditDeviceForm(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	deviceID := c.Param("id")

	device, err := h.deviceRepo.GetByID(deviceID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Device not found", "user": user})
		return
	}

	products, err := h.productRepo.List(&models.FilterParams{})
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error(), "user": user})
		return
	}

	c.HTML(http.StatusOK, "device_form.html", gin.H{
		"title":    "Edit Device",
		"device":   device,
		"products": products,
		"user":     user,
	})
}

func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	deviceID := c.Param("id")
	serialNumber := c.PostForm("serialnumber")
	status := c.PostForm("status")
	notes := c.PostForm("notes")
	
	var productID *uint
	if productIDStr := c.PostForm("productID"); productIDStr != "" {
		if pid, err := strconv.ParseUint(productIDStr, 10, 32); err == nil {
			prodID := uint(pid)
			productID = &prodID
		}
	}
	
	device := models.Device{
		DeviceID:  deviceID,
		ProductID: productID,
		Status:    status,
	}
	
	// Handle optional string fields
	if serialNumber != "" {
		device.SerialNumber = &serialNumber
	}
	if notes != "" {
		device.Notes = &notes
	}
	
	// Handle date fields
	if purchaseDateStr := c.PostForm("purchase_date"); purchaseDateStr != "" {
		if purchaseDate, err := time.Parse("2006-01-02", purchaseDateStr); err == nil {
			device.PurchaseDate = &purchaseDate
		}
	}
	if lastMaintenanceStr := c.PostForm("last_maintenance"); lastMaintenanceStr != "" {
		if lastMaintenance, err := time.Parse("2006-01-02", lastMaintenanceStr); err == nil {
			device.LastMaintenance = &lastMaintenance
		}
	}

	if err := h.deviceRepo.Update(&device); err != nil {
		user, _ := GetCurrentUser(c)
		products, _ := h.productRepo.List(&models.FilterParams{})
		c.HTML(http.StatusInternalServerError, "device_form.html", gin.H{
			"title":    "Edit Device",
			"device":   &device,
			"products": products,
			"error":    err.Error(),
			"user":     user,
		})
		return
	}

	c.Redirect(http.StatusFound, "/devices")
}

func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	deviceID := c.Param("id")

	if err := h.deviceRepo.Delete(deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device deleted successfully"})
}

func (h *DeviceHandler) GetDeviceQR(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	deviceID := c.Param("id")

	device, err := h.deviceRepo.GetByID(deviceID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Device not found", "user": user})
		return
	}

	// Use serial number if available, otherwise use device ID
	identifier := deviceID
	if device.SerialNumber != nil && *device.SerialNumber != "" {
		identifier = *device.SerialNumber
	}
	
	qrCode, err := h.barcodeService.GenerateQRCode(identifier, 256)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error(), "user": user})
		return
	}

	c.Data(http.StatusOK, "image/png", qrCode)
}

func (h *DeviceHandler) GetDeviceBarcode(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	deviceID := c.Param("id")

	device, err := h.deviceRepo.GetByID(deviceID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Device not found", "user": user})
		return
	}

	// Use serial number if available, otherwise use device ID
	identifier := deviceID
	if device.SerialNumber != nil && *device.SerialNumber != "" {
		identifier = *device.SerialNumber
	}
	
	barcode, err := h.barcodeService.GenerateBarcode(identifier)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error(), "user": user})
		return
	}

	c.Data(http.StatusOK, "image/png", barcode)
}

func (h *DeviceHandler) GetAvailableDevices(c *gin.Context) {
	devices, err := h.deviceRepo.GetAvailableDevices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}

// API handlers
func (h *DeviceHandler) ListDevicesAPI(c *gin.Context) {
	params := &models.FilterParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use the new method with categories for case management
	devices, err := h.deviceRepo.ListWithCategories(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}

func (h *DeviceHandler) CreateDeviceAPI(c *gin.Context) {
	var device models.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.deviceRepo.Create(&device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, device)
}

func (h *DeviceHandler) GetDeviceAPI(c *gin.Context) {
	deviceID := c.Param("id")
	device, err := h.deviceRepo.GetByID(deviceID)
	if err != nil {
		// Try by serial number if not found by ID
		device, err = h.deviceRepo.GetBySerialNo(deviceID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"device": device})
}

func (h *DeviceHandler) UpdateDeviceAPI(c *gin.Context) {
	deviceID := c.Param("id")

	var device models.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device.DeviceID = deviceID
	if err := h.deviceRepo.Update(&device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, device)
}

func (h *DeviceHandler) DeleteDeviceAPI(c *gin.Context) {
	deviceID := c.Param("id")

	if err := h.deviceRepo.Delete(deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device deleted successfully"})
}

func (h *DeviceHandler) GetDeviceStatsAPI(c *gin.Context) {
	deviceID := c.Param("id")
	
	// Get device details
	device, err := h.deviceRepo.GetByID(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	// Get device statistics
	stats, err := h.deviceRepo.GetDeviceStats(deviceID)
	if err != nil {
		log.Printf("Error getting device stats: %v", err)
		// Return basic device info even if stats fail
		c.JSON(http.StatusOK, gin.H{
			"device": device,
			"stats": gin.H{
				"totalJobs": 0,
				"totalEarnings": 0.0,
				"totalDaysRented": 0,
				"averageRentalDuration": 0.0,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"device": device,
		"stats": stats,
	})
}

func (h *DeviceHandler) GetAvailableDevicesAPI(c *gin.Context) {
	devices, err := h.deviceRepo.GetAvailableDevices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"devices": devices})
}