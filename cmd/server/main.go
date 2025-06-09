package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"go-barcode-webapp/internal/config"
	"go-barcode-webapp/internal/handlers"
	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/repository"
	"go-barcode-webapp/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.json", "Configuration file path")
	flag.Parse()

	// Set production mode if environment variable is set
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
		log.Println("Running in production mode")
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Printf("Failed to load config, using defaults: %v", err)
		cfg = &config.Config{}
		cfg.Database.Host = "localhost"
		cfg.Database.Port = 3306
		cfg.Database.Database = "jobscanner"
		cfg.Database.Username = "root"
		cfg.Database.Password = ""
		cfg.Database.PoolSize = 5
		cfg.Server.Host = "localhost"
		cfg.Server.Port = 8080
	}

	// Initialize database
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize repositories
	jobRepo := repository.NewJobRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	statusRepo := repository.NewStatusRepository(db)
	productRepo := repository.NewProductRepository(db)
	jobCategoryRepo := repository.NewJobCategoryRepository(db)

	// Initialize services
	barcodeService := services.NewBarcodeService()

	// Auto-migrate database tables (add this for user authentication)
	if err := db.GetDB().AutoMigrate(&models.User{}, &models.Session{}); err != nil {
		log.Printf("Failed to auto-migrate auth tables: %v", err)
	}

	// Initialize handlers
	jobHandler := handlers.NewJobHandler(jobRepo, deviceRepo, customerRepo, statusRepo, jobCategoryRepo)
	deviceHandler := handlers.NewDeviceHandler(deviceRepo, barcodeService, productRepo)
	customerHandler := handlers.NewCustomerHandler(customerRepo)
	statusHandler := handlers.NewStatusHandler(statusRepo)
	productHandler := handlers.NewProductHandler(productRepo)
	barcodeHandler := handlers.NewBarcodeHandler(barcodeService, deviceRepo)
	scannerHandler := handlers.NewScannerHandler(jobRepo, deviceRepo, customerRepo)
	authHandler := handlers.NewAuthHandler(db.GetDB())

	// Setup Gin router
	r := gin.Default()

	// Method override middleware for HTML forms
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "POST" {
			// Check if this is a form submission
			contentType := c.GetHeader("Content-Type")
			if contentType == "application/x-www-form-urlencoded" || 
			   (contentType != "" && len(contentType) > 33 && contentType[:33] == "application/x-www-form-urlencoded") {
				
				// Parse the form to access _method field
				if err := c.Request.ParseForm(); err == nil {
					if method := c.Request.FormValue("_method"); method == "PUT" || method == "DELETE" {
						c.Request.Method = method
					}
				}
			}
		}
		c.Next()
	})

	// Load HTML templates with custom functions
	funcMap := template.FuncMap{
		"deref": func(p *uint) uint {
			if p != nil {
				return *p
			}
			return 0
		},
		"derefString": func(p *string) string {
			if p != nil {
				return *p
			}
			return ""
		},
		"derefFloat": func(p *float64) float64 {
			if p != nil {
				return *p
			}
			return 0.0
		},
	}
	r.SetFuncMap(funcMap)
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "web/static")
	
	// PWA Service Worker route
	r.GET("/sw.js", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.File("web/static/sw.js")
	})

	// Routes
	setupRoutes(r, jobHandler, deviceHandler, customerHandler, statusHandler, productHandler, barcodeHandler, scannerHandler, authHandler)

	// Start server
	addr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func setupRoutes(r *gin.Engine, 
	jobHandler *handlers.JobHandler,
	deviceHandler *handlers.DeviceHandler,
	customerHandler *handlers.CustomerHandler,
	statusHandler *handlers.StatusHandler,
	productHandler *handlers.ProductHandler,
	barcodeHandler *handlers.BarcodeHandler,
	scannerHandler *handlers.ScannerHandler,
	authHandler *handlers.AuthHandler) {

	// Authentication routes (no auth required)
	r.GET("/login", authHandler.LoginForm)
	r.POST("/login", authHandler.Login)
	r.GET("/logout", authHandler.Logout)

	// Protected routes - require authentication
	protected := r.Group("/")
	protected.Use(authHandler.AuthMiddleware())
	{
		// Web interface routes
		protected.GET("/", func(c *gin.Context) {
			user, _ := handlers.GetCurrentUser(c)
			c.HTML(http.StatusOK, "home_new.html", gin.H{
				"title": "Home",
				"user":  user,
			})
		})

		// Job routes
		jobs := protected.Group("/jobs")
		{
			jobs.GET("", jobHandler.ListJobs)
			jobs.GET("/new", jobHandler.NewJobForm)
			jobs.POST("", jobHandler.CreateJob)
			jobs.GET("/:id", jobHandler.GetJob)
			jobs.GET("/:id/edit", jobHandler.EditJobForm)
			jobs.PUT("/:id", jobHandler.UpdateJob)
			jobs.DELETE("/:id", jobHandler.DeleteJob)
			jobs.GET("/:id/devices", jobHandler.GetJobDevices)
			jobs.POST("/:id/devices", jobHandler.AssignDevice)
			jobs.DELETE("/:id/devices/:deviceId", jobHandler.RemoveDevice)
			jobs.POST("/:id/bulk-scan", jobHandler.BulkScanDevices)
		}

		// Device routes
		devices := protected.Group("/devices")
		{
			devices.GET("", deviceHandler.ListDevices)
			devices.GET("/new", deviceHandler.NewDeviceForm)
			devices.POST("", deviceHandler.CreateDevice)
			devices.GET("/:id", deviceHandler.GetDevice)
			devices.GET("/:id/edit", deviceHandler.EditDeviceForm)
			devices.PUT("/:id", deviceHandler.UpdateDevice)
			devices.DELETE("/:id", deviceHandler.DeleteDevice)
			devices.GET("/:id/qr", deviceHandler.GetDeviceQR)
			devices.GET("/:id/barcode", deviceHandler.GetDeviceBarcode)
			devices.GET("/available", deviceHandler.GetAvailableDevices)
		}

		// Customer routes
		customers := protected.Group("/customers")
		{
			customers.GET("", customerHandler.ListCustomers)
			customers.GET("/new", customerHandler.NewCustomerForm)
			customers.POST("", customerHandler.CreateCustomer)
			customers.GET("/:id", customerHandler.GetCustomer)
			customers.GET("/:id/edit", customerHandler.EditCustomerForm)
			customers.PUT("/:id", customerHandler.UpdateCustomer)
			customers.DELETE("/:id", customerHandler.DeleteCustomer)
		}

		// Status routes
		statuses := protected.Group("/statuses")
		{
			statuses.GET("", statusHandler.ListStatuses)
		}

		// Product routes
		products := protected.Group("/products")
		{
			products.GET("", productHandler.ListProducts)
		}

		// Barcode routes
		barcodes := protected.Group("/barcodes")
		{
			barcodes.GET("/device/:serialNo/qr", barcodeHandler.GenerateDeviceQR)
			barcodes.GET("/device/:serialNo/barcode", barcodeHandler.GenerateDeviceBarcode)
		}

		// Scanner routes
		scan := protected.Group("/scan")
		{
			scan.GET("/select", scannerHandler.ScanJobSelection)
			scan.GET("/:jobId", scannerHandler.ScanJob)
			scan.POST("/:jobId/assign", scannerHandler.ScanDevice)
			scan.DELETE("/:jobId/devices/:deviceId", scannerHandler.RemoveDevice)
		}

		// API routes
		api := protected.Group("/api/v1")
		{
			// Job API
			apiJobs := api.Group("/jobs")
			{
				apiJobs.GET("", jobHandler.ListJobsAPI)
				apiJobs.POST("", jobHandler.CreateJobAPI)
				apiJobs.GET("/:id", jobHandler.GetJobAPI)
				apiJobs.PUT("/:id", jobHandler.UpdateJobAPI)
				apiJobs.DELETE("/:id", jobHandler.DeleteJobAPI)
				apiJobs.POST("/:id/devices/:deviceId", jobHandler.AssignDeviceAPI)
				apiJobs.DELETE("/:id/devices/:deviceId", jobHandler.RemoveDeviceAPI)
				apiJobs.POST("/:id/bulk-scan", jobHandler.BulkScanDevicesAPI)
				
				// Scanner API endpoints
				apiJobs.POST("/:id/assign-device", scannerHandler.ScanDevice)
			}

			// Device API
			apiDevices := api.Group("/devices")
			{
				apiDevices.GET("", deviceHandler.ListDevicesAPI)
				apiDevices.POST("", deviceHandler.CreateDeviceAPI)
				apiDevices.GET("/:id", deviceHandler.GetDeviceAPI)
				apiDevices.PUT("/:id", deviceHandler.UpdateDeviceAPI)
				apiDevices.DELETE("/:id", deviceHandler.DeleteDeviceAPI)
				apiDevices.GET("/available", deviceHandler.GetAvailableDevicesAPI)
			}

			// Customer API
			apiCustomers := api.Group("/customers")
			{
				apiCustomers.GET("", customerHandler.ListCustomersAPI)
				apiCustomers.POST("", customerHandler.CreateCustomerAPI)
				apiCustomers.GET("/:id", customerHandler.GetCustomerAPI)
				apiCustomers.PUT("/:id", customerHandler.UpdateCustomerAPI)
				apiCustomers.DELETE("/:id", customerHandler.DeleteCustomerAPI)
			}
		}
	}
}