package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	caseRepo := repository.NewCaseRepository(db)

	// Initialize services
	barcodeService := services.NewBarcodeService()

	// Auto-migrate database tables in correct order (User first, then Session)
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Printf("Failed to auto-migrate users table: %v", err)
	}
	
	if err := db.AutoMigrate(&models.Session{}); err != nil {
		log.Printf("Failed to auto-migrate sessions table: %v", err)
	}

	// Initialize handlers
	jobHandler := handlers.NewJobHandler(jobRepo, deviceRepo, customerRepo, statusRepo, jobCategoryRepo)
	deviceHandler := handlers.NewDeviceHandler(deviceRepo, barcodeService, productRepo)
	customerHandler := handlers.NewCustomerHandler(customerRepo)
	statusHandler := handlers.NewStatusHandler(statusRepo)
	productHandler := handlers.NewProductHandler(productRepo)
	barcodeHandler := handlers.NewBarcodeHandler(barcodeService, deviceRepo)
	scannerHandler := handlers.NewScannerHandler(jobRepo, deviceRepo, customerRepo, caseRepo)
	authHandler := handlers.NewAuthHandler(db.DB)
	caseHandler := handlers.NewCaseHandler(caseRepo, deviceRepo)

	// Setup Gin router
	r := gin.Default()


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

	// Demo routes (no authentication required)
	r.GET("/demo/case-management", caseHandler.CaseManagementDemo)
	r.GET("/demo/case-management-minimal", caseHandler.CaseManagementDemoMinimal)
	r.GET("/demo/case-management-real", caseHandler.CaseManagement)
	r.GET("/demo/case-management-simple", caseHandler.CaseManagementSimple)

	// Routes
	setupRoutes(r, jobHandler, deviceHandler, customerHandler, statusHandler, productHandler, barcodeHandler, scannerHandler, authHandler, caseHandler)

	// Start server
	addr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	// Wrap Gin router with method override support
	methodOverrideHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		originalMethod := req.Method
		if req.Method == "POST" {
			contentType := req.Header.Get("Content-Type")
			log.Printf("POST request to %s with Content-Type: '%s'", req.URL.Path, contentType)
			
			if contentType == "application/x-www-form-urlencoded" || 
			   strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
				
				if err := req.ParseForm(); err == nil {
					methodParam := req.FormValue("_method")
					log.Printf("Form _method parameter: '%s'", methodParam)
					if methodParam == "PUT" || methodParam == "DELETE" {
						log.Printf("Method override: %s -> %s for path: %s", originalMethod, methodParam, req.URL.Path)
						req.Method = methodParam
					}
				}
			}
		}
		// Pass to Gin router
		r.ServeHTTP(w, req)
	})

	log.Printf("Server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, methodOverrideHandler))
}

func setupRoutes(r *gin.Engine, 
	jobHandler *handlers.JobHandler,
	deviceHandler *handlers.DeviceHandler,
	customerHandler *handlers.CustomerHandler,
	statusHandler *handlers.StatusHandler,
	productHandler *handlers.ProductHandler,
	barcodeHandler *handlers.BarcodeHandler,
	scannerHandler *handlers.ScannerHandler,
	authHandler *handlers.AuthHandler,
	caseHandler *handlers.CaseHandler) {

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

		// Case routes
		cases := protected.Group("/cases")
		{
			cases.GET("", caseHandler.ListCases)
			cases.GET("/management", caseHandler.CaseManagement)
			cases.GET("/new", caseHandler.NewCaseForm)
			cases.POST("", caseHandler.CreateCase)
			cases.GET("/:id", caseHandler.GetCase)
			cases.GET("/:id/edit", caseHandler.EditCaseForm)
			cases.PUT("/:id", caseHandler.UpdateCase)
			cases.DELETE("/:id", caseHandler.DeleteCase)
			cases.GET("/:id/devices", caseHandler.CaseDeviceMapping)
			cases.POST("/:id/devices", caseHandler.ScanDeviceToCase)
			cases.DELETE("/:id/devices/:deviceId", caseHandler.RemoveDeviceFromCase)
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

		// User Management routes
		users := protected.Group("/users")
		{
			users.GET("", authHandler.ListUsers)
			users.GET("/new", authHandler.NewUserForm)
			users.POST("", authHandler.CreateUserWeb)
			users.GET("/:id", authHandler.GetUser)
			users.GET("/:id/edit", authHandler.EditUserForm)
			users.PUT("/:id", authHandler.UpdateUser)
			users.DELETE("/:id", authHandler.DeleteUser)
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
				apiJobs.POST("/:id/assign-case", scannerHandler.ScanCase)
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

			// Case API
			apiCases := api.Group("/cases")
			{
				apiCases.GET("", caseHandler.ListCasesAPI)
				apiCases.POST("", caseHandler.CreateCaseAPI)
				apiCases.GET("/:id", caseHandler.GetCaseAPI)
				apiCases.PUT("/:id", caseHandler.UpdateCaseAPI)
				apiCases.DELETE("/:id", caseHandler.DeleteCaseAPI)
			}
		}
	}
}