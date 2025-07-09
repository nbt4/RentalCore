package handlers

import (
	"net/http"
	"strconv"
	"log"
	"time"

	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/repository"
	"go-barcode-webapp/internal/services"

	"github.com/gin-gonic/gin"
)

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

	viewType := c.DefaultQuery("view", "list") // Default to list view

	if viewType == "categorized" {
		categorizedDevices, err := h.deviceRepo.GetDevicesGroupedByCategory(params)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error(), "user": user})
			return
		}
		c.HTML(http.StatusOK, "devices_standalone.html", gin.H{
			"title":              "Devices by Category",
			"categorizedDevices": categorizedDevices,
			"params":             params,
			"user":               user,
			"viewType":           "categorized",
			"categorized":        true,
		})
	} else {
		devices, err := h.deviceRepo.List(params)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error(), "user": user})
			return
		}
		c.HTML(http.StatusOK, "devices_standalone.html", gin.H{
			"title":       "Devices",
			"devices":     devices,
			"params":      params,
			"user":        user,
			"viewType":    "list",
			"categorized": false,
		})
	}
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
	serialNumber := c.PostForm("serialnumber")
	status := c.PostForm("status")
	notes := c.PostForm("notes")
	
	if status == "" {
		status = "free"
	}
	
	var productID *uint
	if productIDStr := c.PostForm("productID"); productIDStr != "" {
		if pid, err := strconv.ParseUint(productIDStr, 10, 32); err == nil {
			prodID := uint(pid)
			productID = &prodID
		}
	}
	
	device := models.Device{
		DeviceID:     "", // Let database generate the ID automatically
		ProductID:    productID,
		SerialNumber: &serialNumber,
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

	if err := h.deviceRepo.Create(&device); err != nil {
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

func (h *DeviceHandler) GetAvailableDevicesAPI(c *gin.Context) {
	devices, err := h.deviceRepo.GetAvailableDevices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"devices": devices})
}