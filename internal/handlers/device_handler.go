package handlers

import (
	"net/http"
	"strconv"

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
	params := &models.FilterParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error()})
		return
	}

	devices, err := h.deviceRepo.List(params)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "devices_new.html", gin.H{
		"title":   "Devices",
		"devices": devices,
		"params":  params,
	})
}

func (h *DeviceHandler) NewDeviceForm(c *gin.Context) {
	products, err := h.productRepo.List(&models.FilterParams{})
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "device_form_new.html", gin.H{
		"title":    "New Device",
		"device":   &models.Device{},
		"products": products,
	})
}

func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	deviceID := c.PostForm("deviceID")
	serialNumber := c.PostForm("serialnumber")
	status := c.PostForm("status")
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
		DeviceID:     deviceID,
		ProductID:    productID,
		SerialNumber: &serialNumber,
		Status:       status,
	}

	if err := h.deviceRepo.Create(&device); err != nil {
		products, _ := h.productRepo.List(&models.FilterParams{})
		c.HTML(http.StatusInternalServerError, "device_form_new.html", gin.H{
			"title":    "New Device",
			"device":   &device,
			"products": products,
			"error":    err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/devices")
}

func (h *DeviceHandler) GetDevice(c *gin.Context) {
	deviceID := c.Param("id")

	device, err := h.deviceRepo.GetByID(deviceID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Device not found"})
		return
	}

	c.HTML(http.StatusOK, "device_detail.html", gin.H{
		"device": device,
	})
}

func (h *DeviceHandler) EditDeviceForm(c *gin.Context) {
	deviceID := c.Param("id")

	device, err := h.deviceRepo.GetByID(deviceID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Device not found"})
		return
	}

	products, err := h.productRepo.List(&models.FilterParams{})
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "device_form_new.html", gin.H{
		"title":    "Edit Device",
		"device":   device,
		"products": products,
	})
}

func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	deviceID := c.Param("id")
	serialNumber := c.PostForm("serialnumber")
	status := c.PostForm("status")
	
	var productID *uint
	if productIDStr := c.PostForm("productID"); productIDStr != "" {
		if pid, err := strconv.ParseUint(productIDStr, 10, 32); err == nil {
			prodID := uint(pid)
			productID = &prodID
		}
	}
	
	device := models.Device{
		DeviceID:     deviceID,
		ProductID:    productID,
		SerialNumber: &serialNumber,
		Status:       status,
	}

	if err := h.deviceRepo.Update(&device); err != nil {
		products, _ := h.productRepo.List(&models.FilterParams{})
		c.HTML(http.StatusInternalServerError, "device_form_new.html", gin.H{
			"title":    "Edit Device",
			"device":   &device,
			"products": products,
			"error":    err.Error(),
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
	deviceID := c.Param("id")

	device, err := h.deviceRepo.GetByID(deviceID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Device not found"})
		return
	}

	// Use serial number if available, otherwise use device ID
	identifier := deviceID
	if device.SerialNumber != nil && *device.SerialNumber != "" {
		identifier = *device.SerialNumber
	}
	
	qrCode, err := h.barcodeService.GenerateQRCode(identifier, 256)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/png", qrCode)
}

func (h *DeviceHandler) GetDeviceBarcode(c *gin.Context) {
	deviceID := c.Param("id")

	device, err := h.deviceRepo.GetByID(deviceID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Device not found"})
		return
	}

	// Use serial number if available, otherwise use device ID
	identifier := deviceID
	if device.SerialNumber != nil && *device.SerialNumber != "" {
		identifier = *device.SerialNumber
	}
	
	barcode, err := h.barcodeService.GenerateBarcode(identifier)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
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

	devices, err := h.deviceRepo.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"devices": devices})
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