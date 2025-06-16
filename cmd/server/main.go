package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	jobTemplateRepo := repository.NewJobTemplateRepository(db)
	equipmentPackageRepo := repository.NewEquipmentPackageRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)

	// Initialize services
	barcodeService := services.NewBarcodeService()

	// Auto-migrate database tables in correct order (User first, then Session)
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Printf("Failed to auto-migrate users table: %v", err)
	}
	
	if err := db.AutoMigrate(&models.Session{}); err != nil {
		log.Printf("Failed to auto-migrate sessions table: %v", err)
	}
	
	if err := db.AutoMigrate(&models.EquipmentPackage{}); err != nil {
		log.Printf("Failed to auto-migrate equipment_packages table: %v", err)
	}

	// Initialize handlers
	jobHandler := handlers.NewJobHandler(jobRepo, deviceRepo, customerRepo, statusRepo, jobCategoryRepo)
	deviceHandler := handlers.NewDeviceHandler(deviceRepo, barcodeService, productRepo)
	customerHandler := handlers.NewCustomerHandler(customerRepo)
	statusHandler := handlers.NewStatusHandler(statusRepo)
	productHandler := handlers.NewProductHandler(productRepo)
	barcodeHandler := handlers.NewBarcodeHandler(barcodeService, deviceRepo)
	scannerHandler := handlers.NewScannerHandler(jobRepo, deviceRepo, customerRepo, caseRepo)
	authHandler := handlers.NewAuthHandler(db.DB, cfg)
	
	// Start session cleanup background process
	authHandler.StartSessionCleanup()
	
	caseHandler := handlers.NewCaseHandler(caseRepo, deviceRepo)
	analyticsHandler := handlers.NewAnalyticsHandler(db.DB)
	searchHandler := handlers.NewSearchHandler(db.DB)
	pwaHandler := handlers.NewPWAHandler(db.DB)
	workflowHandler := handlers.NewWorkflowHandler(jobTemplateRepo, jobRepo, customerRepo, equipmentPackageRepo, deviceRepo)
	documentHandler := handlers.NewDocumentHandler(db.DB)
	financialHandler := handlers.NewFinancialHandler(db.DB)
	securityHandler := handlers.NewSecurityHandler(db.DB)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceRepo, customerRepo, jobRepo, deviceRepo, equipmentPackageRepo, &cfg.Email, &cfg.PDF)

	// Setup Gin router with error handling
	r := gin.New()
	
	// Add comprehensive error handling middleware
	r.Use(gin.Logger())
	r.Use(handlers.GlobalErrorHandler()) // Custom recovery with proper error pages


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
		"humanizeBytes": func(bytes int64) string {
			if bytes == 0 {
				return "0 B"
			}
			const unit = 1024
			sizes := []string{"B", "KB", "MB", "GB", "TB"}
			i := 0
			for bytes >= unit && i < len(sizes)-1 {
				bytes /= unit
				i++
			}
			if i == 0 {
				return fmt.Sprintf("%d %s", bytes, sizes[i])
			}
			return fmt.Sprintf("%.1f %s", float64(bytes), sizes[i])
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"title": func(s string) string {
			return strings.Title(s)
		},
		"getStatusColor": func(status string) string {
			switch status {
			case "completed":
				return "success"
			case "pending":
				return "warning"
			case "failed":
				return "danger"
			case "cancelled":
				return "secondary"
			default:
				return "secondary"
			}
		},
		"now": func() time.Time {
			return time.Now()
		},
		"daysAgo": func(date time.Time) int {
			return int(time.Since(date).Hours() / 24)
		},
		"daysUntil": func(date time.Time) int {
			return int(time.Until(date).Hours() / 24)
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

	// PWA public routes (no authentication required)
	r.GET("/manifest.json", func(c *gin.Context) {
		c.File("web/static/manifest.json")
	})
	

	// Initialize default roles
	if err := securityHandler.InitializeDefaultRoles(); err != nil {
		log.Printf("Failed to initialize default roles: %v", err)
	}

	// Routes
	setupRoutes(r, jobHandler, deviceHandler, customerHandler, statusHandler, productHandler, barcodeHandler, scannerHandler, authHandler, caseHandler, analyticsHandler, searchHandler, pwaHandler, workflowHandler, documentHandler, financialHandler, securityHandler, invoiceHandler)
	
	// Add 404 handler as the last route
	r.NoRoute(handlers.NotFoundHandler())

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
	caseHandler *handlers.CaseHandler,
	analyticsHandler *handlers.AnalyticsHandler,
	searchHandler *handlers.SearchHandler,
	pwaHandler *handlers.PWAHandler,
	workflowHandler *handlers.WorkflowHandler,
	documentHandler *handlers.DocumentHandler,
	financialHandler *handlers.FinancialHandler,
	securityHandler *handlers.SecurityHandler,
	invoiceHandler *handlers.InvoiceHandler) {

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
			c.HTML(http.StatusOK, "home.html", gin.H{
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

		// Analytics routes
		analytics := protected.Group("/analytics")
		{
			analytics.GET("", analyticsHandler.Dashboard)
			analytics.GET("/revenue", analyticsHandler.GetRevenueAPI)
			analytics.GET("/equipment", analyticsHandler.GetEquipmentAPI)
			analytics.GET("/export", analyticsHandler.ExportAnalytics)
		}

		// Search routes
		search := protected.Group("/search")
		{
			search.GET("/global", searchHandler.GlobalSearch)
			search.POST("/advanced", searchHandler.AdvancedSearch)
			search.GET("/suggestions", searchHandler.SearchSuggestions)
			search.GET("/saved", searchHandler.SavedSearches)
			search.DELETE("/saved/:id", searchHandler.DeleteSavedSearch)
		}

		// PWA routes
		pwa := protected.Group("/pwa")
		{
			pwa.POST("/subscribe", pwaHandler.SubscribePush)
			pwa.POST("/unsubscribe", pwaHandler.UnsubscribePush)
			pwa.POST("/sync", pwaHandler.SyncOfflineData)
			pwa.GET("/manifest", pwaHandler.GetOfflineManifest)
			pwa.GET("/install", pwaHandler.InstallPrompt)
			pwa.GET("/status", pwaHandler.GetConnectionStatus)
		}

		// Workflow routes
		workflow := protected.Group("/workflow")
		{
			// Job Templates
			templates := workflow.Group("/templates")
			{
				templates.GET("", workflowHandler.ListJobTemplates)
				templates.GET("/new", workflowHandler.NewJobTemplateForm)
				templates.POST("", workflowHandler.CreateJobTemplate)
				templates.GET("/:id", workflowHandler.GetJobTemplate)
				templates.GET("/:id/edit", workflowHandler.EditJobTemplateForm)
				templates.PUT("/:id", workflowHandler.UpdateJobTemplate)
				templates.DELETE("/:id", workflowHandler.DeleteJobTemplate)
				templates.POST("/:id/create-job", workflowHandler.CreateJobFromTemplate)
			}

			// Equipment Packages
			packages := workflow.Group("/packages")
			{
				packages.GET("", workflowHandler.ListEquipmentPackages)
				packages.GET("/new", workflowHandler.NewEquipmentPackageForm)
				packages.GET("/debug", workflowHandler.DebugPackageForm)
				packages.POST("", workflowHandler.CreateEquipmentPackage)
				packages.GET("/:id", workflowHandler.GetEquipmentPackage)
			}

			// Bulk Operations
			bulk := workflow.Group("/bulk")
			{
				bulk.GET("", workflowHandler.BulkOperationsForm)
				bulk.POST("/update-status", workflowHandler.BulkUpdateDeviceStatus)
				bulk.POST("/assign-job", workflowHandler.BulkAssignToJob)
				bulk.POST("/generate-qr", workflowHandler.BulkGenerateQRCodes)
			}

			// Workflow API
			workflow.GET("/stats", workflowHandler.GetWorkflowStats)
		}

		// Document routes
		documents := protected.Group("/documents")
		{
			documents.GET("", documentHandler.ListDocuments)
			documents.GET("/upload", documentHandler.UploadDocumentForm)
			documents.POST("/upload", documentHandler.UploadDocument)
			documents.GET("/:id", documentHandler.GetDocument)
			documents.GET("/:id/view", documentHandler.ViewDocument)
			documents.GET("/:id/download", documentHandler.DownloadDocument)
			documents.DELETE("/:id", documentHandler.DeleteDocument)
			documents.GET("/:id/sign", documentHandler.SignatureForm)
			documents.POST("/:id/sign", documentHandler.AddSignature)
			documents.GET("/signatures/:id/verify", documentHandler.VerifySignature)
		}

		// Financial routes
		financial := protected.Group("/financial")
		{
			financial.GET("", financialHandler.FinancialDashboard)
			financial.GET("/transactions", financialHandler.ListTransactions)
			financial.GET("/transactions/new", financialHandler.NewTransactionForm)
			financial.POST("/transactions", financialHandler.CreateTransaction)
			financial.GET("/transactions/:id", financialHandler.GetTransaction)
			financial.PUT("/transactions/:id/status", financialHandler.UpdateTransactionStatus)
			financial.POST("/jobs/:jobId/invoice", financialHandler.GenerateInvoice)
			financial.GET("/reports", financialHandler.FinancialReports)
			financial.GET("/api/revenue-report", financialHandler.GetRevenueReport)
			financial.GET("/api/payment-report", financialHandler.GetPaymentReport)
		}

		// Invoice routes
		invoices := protected.Group("/invoices")
		{
			invoices.GET("", invoiceHandler.ListInvoices)
			invoices.GET("/new", invoiceHandler.NewInvoiceForm)
			invoices.POST("", invoiceHandler.CreateInvoice)
			invoices.GET("/:id", invoiceHandler.GetInvoice)
			invoices.GET("/:id/edit", invoiceHandler.EditInvoiceForm)
			invoices.PUT("/:id", invoiceHandler.UpdateInvoice)
			invoices.DELETE("/:id", invoiceHandler.DeleteInvoice)
			invoices.PUT("/:id/status", invoiceHandler.UpdateInvoiceStatus)
			invoices.GET("/:id/pdf", invoiceHandler.GenerateInvoicePDF)
			invoices.GET("/:id/preview", invoiceHandler.PreviewInvoicePDF)
			invoices.POST("/:id/email", invoiceHandler.EmailInvoice)
		}

		// Invoice template routes
		templates := protected.Group("/invoice-templates")
		{
			templates.GET("", invoiceHandler.ListInvoiceTemplates)
			templates.GET("/new", invoiceHandler.NewInvoiceTemplateForm)
			templates.POST("", invoiceHandler.CreateInvoiceTemplate)
			templates.GET("/:id", invoiceHandler.GetInvoiceTemplate)
			templates.GET("/:id/edit", invoiceHandler.EditInvoiceTemplateForm)
			templates.PUT("/:id", invoiceHandler.UpdateInvoiceTemplate)
			templates.DELETE("/:id", invoiceHandler.DeleteInvoiceTemplate)
		}

		// Invoice settings routes
		settings := protected.Group("/settings")
		{
			settings.GET("/company", invoiceHandler.CompanySettingsForm)
			settings.PUT("/company", invoiceHandler.UpdateCompanySettings)
			settings.GET("/invoices", invoiceHandler.InvoiceSettingsForm)
			settings.PUT("/invoices", invoiceHandler.UpdateInvoiceSettings)
		}

		// Security & Admin routes
		security := protected.Group("/security")
		{
			// Web interface routes
			security.GET("/roles", func(c *gin.Context) {
				user, _ := handlers.GetCurrentUser(c)
				c.HTML(http.StatusOK, "security_roles.html", gin.H{
					"title": "Role Management",
					"user":  user,
				})
			})
			security.GET("/audit", func(c *gin.Context) {
				user, _ := handlers.GetCurrentUser(c)
				c.HTML(http.StatusOK, "security_audit.html", gin.H{
					"title": "Audit Logs",
					"user":  user,
				})
			})

			// Role management API
			rolesAPI := security.Group("/api/roles")
			{
				rolesAPI.GET("", securityHandler.GetRoles)
				rolesAPI.GET("/:id", securityHandler.GetRole)
				rolesAPI.POST("", securityHandler.CreateRole)
				rolesAPI.PUT("/:id", securityHandler.UpdateRole)
				rolesAPI.DELETE("/:id", securityHandler.DeleteRole)
			}

			// User role management API
			userRolesAPI := security.Group("/api/users")
			{
				userRolesAPI.GET("/:userId/roles", securityHandler.GetUserRoles)
				userRolesAPI.POST("/:userId/roles", securityHandler.AssignUserRole)
				userRolesAPI.DELETE("/:userId/roles/:roleId", securityHandler.RevokeUserRole)
			}

			// Audit API
			auditAPI := security.Group("/api/audit")
			{
				auditAPI.GET("", securityHandler.GetAuditLogs)
				auditAPI.GET("/:id", securityHandler.GetAuditLog)
			}

			// Permissions API
			permissionsAPI := security.Group("/api/permissions")
			{
				permissionsAPI.GET("", securityHandler.GetPermissions)
				permissionsAPI.GET("/definitions", securityHandler.GetPermissionDefinitionsAPI)
				permissionsAPI.GET("/check", securityHandler.CheckPermission)
			}
		}

		// Mobile scanner routes
		protected.GET("/mobile/scanner/:jobId", func(c *gin.Context) {
			jobID := c.Param("jobId")
			user, _ := handlers.GetCurrentUser(c)
			c.HTML(http.StatusOK, "mobile_scanner.html", gin.H{
				"title":   "Mobile Scanner",
				"user":    user,
				"jobID":   jobID,
				"jobName": "Job #" + jobID,
			})
		})
		
		// Enhanced mobile scanner route
		protected.GET("/mobile/scanner/:jobId/enhanced", func(c *gin.Context) {
			jobID := c.Param("jobId")
			user, _ := handlers.GetCurrentUser(c)
			c.HTML(http.StatusOK, "mobile_scanner_enhanced.html", gin.H{
				"title":   "Enhanced Mobile Scanner",
				"user":    user,
				"jobID":   jobID,
				"jobName": "Job #" + jobID,
			})
		})

		// User Management - Use explicit routing without parameter conflicts
		
		// Main user management routes
		protected.GET("/users", authHandler.ListUsers)
		protected.POST("/users", authHandler.CreateUserWeb)
		
		// User form and management routes with no parameter conflicts
		protected.GET("/user-management/new", authHandler.NewUserForm)
		protected.GET("/user-management/:id/edit", authHandler.EditUserForm)
		protected.GET("/user-management/:id/view", authHandler.GetUser)
		protected.PUT("/user-management/:id", authHandler.UpdateUser)
		protected.DELETE("/user-management/:id", authHandler.DeleteUser)
		
		// Direct explicit routes for old paths - NO parameter routes under /users
		protected.GET("/users/new", func(c *gin.Context) {
			c.Redirect(http.StatusSeeOther, "/user-management/new")
		})

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
				apiCases.GET("/:id/devices", caseHandler.GetCaseDevicesAPI)
			}

			// Workflow API
			apiWorkflow := api.Group("/workflow")
			{
				// Job Templates API
				templates := apiWorkflow.Group("/templates")
				{
					templates.GET("", workflowHandler.ListJobTemplatesAPI)
					templates.POST("", workflowHandler.CreateJobTemplate)
					templates.GET("/most-used", workflowHandler.GetMostUsedTemplatesAPI)
					templates.GET("/category/:categoryId", workflowHandler.GetTemplatesByCategoryAPI)
					templates.GET("/:id", workflowHandler.GetJobTemplate)
					templates.PUT("/:id", workflowHandler.UpdateJobTemplate)
					templates.DELETE("/:id", workflowHandler.DeleteJobTemplate)
					templates.POST("/:id/create-job", workflowHandler.CreateJobFromTemplate)
				}
			}

			// Document API
			apiDocuments := api.Group("/documents")
			{
				apiDocuments.GET("", documentHandler.ListDocumentsAPI)
				apiDocuments.GET("/stats", documentHandler.GetDocumentStats)
				apiDocuments.GET("/:id", documentHandler.GetDocument)
				apiDocuments.DELETE("/:id", documentHandler.DeleteDocument)
			}

			// Financial API
			apiFinancial := api.Group("/financial")
			{
				apiFinancial.GET("/transactions", financialHandler.ListTransactionsAPI)
				apiFinancial.GET("/stats", financialHandler.GetFinancialStatsAPI)
				apiFinancial.GET("/revenue-report", financialHandler.GetRevenueReport)
				apiFinancial.GET("/payment-report", financialHandler.GetPaymentReport)
			}

			// Invoice API
			apiInvoices := api.Group("/invoices")
			{
				apiInvoices.GET("", invoiceHandler.GetInvoicesAPI)
				apiInvoices.POST("", invoiceHandler.CreateInvoice)
				apiInvoices.GET("/:id", invoiceHandler.GetInvoice)
				apiInvoices.PUT("/:id", invoiceHandler.UpdateInvoice)
				apiInvoices.DELETE("/:id", invoiceHandler.DeleteInvoice)
				apiInvoices.PUT("/:id/status", invoiceHandler.UpdateInvoiceStatus)
				apiInvoices.GET("/stats", invoiceHandler.GetInvoiceStatsAPI)
			}
		}

		// Additional API routes (outside v1 group for legacy compatibility)
		legacyAPI := protected.Group("/api")
		{
			// Invoice API
			legacyAPI.GET("/invoices", invoiceHandler.GetInvoicesAPI)
			legacyAPI.POST("/invoices", invoiceHandler.CreateInvoice)
			legacyAPI.GET("/invoices/:id", invoiceHandler.GetInvoice)
			legacyAPI.PUT("/invoices/:id", invoiceHandler.UpdateInvoice)
			legacyAPI.DELETE("/invoices/:id", invoiceHandler.DeleteInvoice)
			legacyAPI.PUT("/invoices/:id/status", invoiceHandler.UpdateInvoiceStatus)
			legacyAPI.GET("/invoices/stats", invoiceHandler.GetInvoiceStatsAPI)
			
			// Company settings API
			legacyAPI.GET("/company-settings", invoiceHandler.CompanySettingsForm)
			legacyAPI.PUT("/company-settings", invoiceHandler.UpdateCompanySettings)
			
			// Invoice settings API
			legacyAPI.GET("/invoice-settings", invoiceHandler.InvoiceSettingsForm)
			legacyAPI.PUT("/invoice-settings", invoiceHandler.UpdateInvoiceSettings)
			
			// Email API
			legacyAPI.POST("/test-email", invoiceHandler.TestEmailSettings)

			// Security API
			apiSecurity := legacyAPI.Group("/security")
			{
				// Roles API
				apiSecurity.GET("/roles", securityHandler.GetRoles)
				apiSecurity.GET("/roles/:id", securityHandler.GetRole)
				apiSecurity.POST("/roles", securityHandler.CreateRole)
				apiSecurity.PUT("/roles/:id", securityHandler.UpdateRole)
				apiSecurity.DELETE("/roles/:id", securityHandler.DeleteRole)

				// User roles API
				apiSecurity.GET("/users/:userId/roles", securityHandler.GetUserRoles)
				apiSecurity.POST("/users/:userId/roles", securityHandler.AssignUserRole)
				apiSecurity.DELETE("/users/:userId/roles/:roleId", securityHandler.RevokeUserRole)

				// Audit API
				apiSecurity.GET("/audit", securityHandler.GetAuditLogs)
				apiSecurity.GET("/audit/:id", securityHandler.GetAuditLog)

				// Permissions API
				apiSecurity.GET("/permissions", securityHandler.GetPermissions)
				apiSecurity.GET("/check-permission", securityHandler.CheckPermission)
			}
		}
	}
}