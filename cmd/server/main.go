package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go-barcode-webapp/internal/cache"
	"go-barcode-webapp/internal/config"
	"go-barcode-webapp/internal/handlers"
	"go-barcode-webapp/internal/logger"
	"go-barcode-webapp/internal/middleware"
	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/monitoring"
	"go-barcode-webapp/internal/repository"
	"go-barcode-webapp/internal/services"
	"go-barcode-webapp/internal/compliance"

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
		cfg.Database.Host = "tsunami-events.de"
		cfg.Database.Port = 3306
		cfg.Database.Database = "TS-Lager"
		cfg.Database.Username = "tsweb"
		cfg.Database.Password = "N1KO4cCYnp3Tyf"
		cfg.Database.PoolSize = 5
		cfg.Server.Host = "0.0.0.0"
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

	// Apply performance indexes for optimal database performance (commented out for faster startup)
	// if err := config.ApplyPerformanceIndexes(db.DB); err != nil {
	//	log.Printf("Warning: Failed to apply performance indexes: %v", err)
	// }

	// Initialize structured logger
	environment := "development"
	if os.Getenv("GIN_MODE") == "release" {
		environment = "production"
	}
	
	loggerConfig := logger.LoggerConfig{
		Level:        logger.INFO,
		Service:      "go-barcode-webapp",
		Version:      "1.0.0",
		Environment:  environment,
		OutputPath:   "", // stdout
		EnableCaller: true,
	}
	
	if err := logger.InitializeLogger(loggerConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.GlobalLogger.Close()
	
	// Initialize error tracker
	monitoring.InitializeErrorTracker(1000, 7*24*time.Hour) // 1000 errors, 7 days retention
	
	// Initialize cache manager
	cacheManager := cache.NewCacheManager()
	
	// Initialize performance monitor
	perfMonitor := middleware.NewPerformanceMonitor(500 * time.Millisecond) // 500ms slow threshold
	
	// Initialize compliance system
	complianceMiddleware, err := compliance.NewComplianceMiddleware(
		db.DB, 
		"./archives", 
		cfg.Security.EncryptionKey,
	)
	if err != nil {
		log.Printf("Warning: Failed to create compliance middleware: %v", err)
		// Use a dummy middleware for development
		complianceMiddleware = nil
	} else {
		// Initialize compliance database tables
		if err := complianceMiddleware.InitializeCompliance(); err != nil {
			log.Printf("Warning: Failed to initialize compliance system: %v", err)
		}
	}
	
	// Start compliance background tasks
	if complianceMiddleware != nil {
		go complianceMiddleware.PeriodicComplianceCheck(context.Background())
	}

	// Initialize repositories
	jobRepo := repository.NewJobRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	statusRepo := repository.NewStatusRepository(db)
	productRepo := repository.NewProductRepository(db)
	jobCategoryRepo := repository.NewJobCategoryRepository(db)
	caseRepo := repository.NewCaseRepository(db)
	equipmentPackageRepo := repository.NewEquipmentPackageRepository(db)
	invoiceRepo := repository.NewInvoiceRepositoryNew(db) // Using NEW fixed repository

	// Initialize services
	barcodeService := services.NewBarcodeService()

	// Auto-migration disabled - database schema managed manually
	log.Printf("Database auto-migration disabled - using manual schema management")

	// Initialize handlers
	jobHandler := handlers.NewJobHandler(jobRepo, deviceRepo, customerRepo, statusRepo, jobCategoryRepo)
	deviceHandler := handlers.NewDeviceHandler(deviceRepo, barcodeService, productRepo)
	customerHandler := handlers.NewCustomerHandler(customerRepo)
	statusHandler := handlers.NewStatusHandler(statusRepo)
	productHandler := handlers.NewProductHandler(productRepo)
	barcodeHandler := handlers.NewBarcodeHandler(barcodeService, deviceRepo)
	scannerHandler := handlers.NewScannerHandler(jobRepo, deviceRepo, customerRepo, caseRepo)
	authHandler := handlers.NewAuthHandler(db.DB, cfg)
	homeHandler := handlers.NewHomeHandler(jobRepo, deviceRepo, customerRepo, caseRepo, db.DB)
	
	// Start session cleanup background process
	authHandler.StartSessionCleanup()
	
	caseHandler := handlers.NewCaseHandler(caseRepo, deviceRepo)
	analyticsHandler := handlers.NewAnalyticsHandler(db.DB)
	searchHandler := handlers.NewSearchHandler(db.DB)
	pwaHandler := handlers.NewPWAHandler(db.DB)
	workflowHandler := handlers.NewWorkflowHandler(jobRepo, customerRepo, equipmentPackageRepo, deviceRepo, db.DB, barcodeService)
	equipmentPackageHandler := handlers.NewEquipmentPackageHandler(equipmentPackageRepo, deviceRepo)
	documentHandler := handlers.NewDocumentHandler(db.DB)
	financialHandler := handlers.NewFinancialHandler(db.DB)
	securityHandler := handlers.NewSecurityHandler(db.DB)
	invoiceHandler := handlers.NewInvoiceHandlerNew(invoiceRepo, customerRepo, jobRepo, deviceRepo, equipmentPackageRepo, productRepo, &cfg.PDF)
	templateHandler := handlers.NewInvoiceTemplateHandler(invoiceRepo)
	companyHandler := handlers.NewCompanyHandler(db.DB)
	monitoringHandler := handlers.NewMonitoringHandler(db.DB, monitoring.GlobalErrorTracker, perfMonitor, cacheManager)

	// Create default invoice template if none exists
	if err := createDefaultTemplate(templateHandler, invoiceRepo); err != nil {
		log.Printf("Warning: Failed to create default template: %v", err)
	}

	// Setup Gin router with error handling
	r := gin.New()
	
	// Add monitoring, logging and compliance middleware
	r.Use(logger.GlobalLogger.LoggingMiddleware())
	r.Use(monitoring.GlobalErrorTracker.ErrorTrackingMiddleware())
	r.Use(perfMonitor.PerformanceMiddleware())
	
	// Add route debugging middleware
	r.Use(func(c *gin.Context) {
		log.Printf("Route Debug: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})
	
	if complianceMiddleware != nil {
		r.Use(complianceMiddleware.AuditMiddleware())
		r.Use(complianceMiddleware.ComplianceStatusMiddleware())
	}
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
		"split": func(s, sep string) []string {
			if s == "" {
				return []string{}
			}
			return strings.Split(s, sep)
		},
		"trim": func(s string) string {
			return strings.TrimSpace(s)
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"timeAgo": func(t *time.Time) string {
			if t == nil {
				return "Never"
			}
			duration := time.Since(*t)
			if duration < time.Minute {
				return "Just now"
			} else if duration < time.Hour {
				minutes := int(duration.Minutes())
				return fmt.Sprintf("%d min ago", minutes)
			} else if duration < 24*time.Hour {
				hours := int(duration.Hours())
				return fmt.Sprintf("%d hours ago", hours)
			} else {
				days := int(duration.Hours() / 24)
				if days == 1 {
					return "Yesterday"
				}
				return fmt.Sprintf("%d days ago", days)
			}
		},
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02 15:04")
		},
		"substr": func(s string, start, end int) string {
			if len(s) == 0 {
				return ""
			}
			if start >= len(s) {
				return ""
			}
			if end > len(s) {
				end = len(s)
			}
			if start < 0 {
				start = 0
			}
			return s[start:end]
		},
		"mul": func(a, b interface{}) float64 {
			var aVal, bVal float64
			switch v := a.(type) {
			case float64:
				aVal = v
			case *float64:
				if v != nil {
					aVal = *v
				}
			case int:
				aVal = float64(v)
			case uint:
				aVal = float64(v)
			}
			switch v := b.(type) {
			case float64:
				bVal = v
			case *float64:
				if v != nil {
					bVal = *v
				}
			case int:
				bVal = float64(v)
			case uint:
				bVal = float64(v)
			}
			return aVal * bVal
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"gt": func(a, b int) bool {
			return a > b
		},
		"len": func(slice interface{}) int {
			switch v := slice.(type) {
			case []interface{}:
				return len(v)
			case string:
				return len(v)
			default:
				return 0
			}
		},
	}
	r.SetFuncMap(funcMap)
	r.LoadHTMLGlob("web/templates/*")
	
	// Add caching for static files
	r.StaticFS("/static", http.Dir("web/static"))
	r.StaticFS("/uploads", http.Dir("uploads"))
	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/static/") {
			c.Header("Cache-Control", "public, max-age=3600")
			c.Header("ETag", fmt.Sprintf(`"%x"`, time.Now().Unix()))
		}
		c.Next()
	})
	
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
	setupRoutes(r, jobHandler, deviceHandler, customerHandler, statusHandler, productHandler, barcodeHandler, scannerHandler, authHandler, homeHandler, caseHandler, analyticsHandler, searchHandler, pwaHandler, workflowHandler, equipmentPackageHandler, documentHandler, financialHandler, securityHandler, invoiceHandler, templateHandler, companyHandler, monitoringHandler, complianceMiddleware)
	
	// Add dedicated error route
	r.GET("/error", func(c *gin.Context) {
		code := c.DefaultQuery("code", "500")
		message := c.DefaultQuery("message", "Internal Server Error")
		details := c.DefaultQuery("details", "Something went wrong on the server")
		
		// Convert code to integer for template comparison
		codeInt, _ := strconv.Atoi(code)
		
		c.HTML(http.StatusOK, "error_page.html", gin.H{
			"error_code":    codeInt,
			"error_message": message,
			"error_details": details,
			"request_id":    c.GetHeader("X-Request-Id"),
			"timestamp":     time.Now().Format("2006-01-02 15:04:05"),
			"user":          nil,
		})
	})
	
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
	homeHandler *handlers.HomeHandler,
	caseHandler *handlers.CaseHandler,
	analyticsHandler *handlers.AnalyticsHandler,
	searchHandler *handlers.SearchHandler,
	pwaHandler *handlers.PWAHandler,
	workflowHandler *handlers.WorkflowHandler,
	equipmentPackageHandler *handlers.EquipmentPackageHandler,
	documentHandler *handlers.DocumentHandler,
	financialHandler *handlers.FinancialHandler,
	securityHandler *handlers.SecurityHandler,
	invoiceHandler *handlers.InvoiceHandlerNew,
	templateHandler *handlers.InvoiceTemplateHandler,
	companyHandler *handlers.CompanyHandler,
	monitoringHandler *handlers.MonitoringHandler,
	complianceMiddleware *compliance.ComplianceMiddleware) {

	// Root route - redirect to dashboard if authenticated, login if not
	r.GET("/", func(c *gin.Context) {
		// Check if user is authenticated by looking for session
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
	})

	// Authentication routes (no auth required)
	r.GET("/login", authHandler.LoginForm)
	r.POST("/login", authHandler.Login)
	r.GET("/logout", authHandler.Logout)
	
	// Debug route (no auth required)
	r.GET("/debug/invoice-test", func(c *gin.Context) {
		c.String(http.StatusOK, "Invoice test route is working! Time: %s", time.Now().Format("2006-01-02 15:04:05"))
	})
	
	// Debug route for customer selection
	r.GET("/debug/customer-selection", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusOK, "Customer selection debug route working!")
	})
	
	// Debug route for route testing
	r.GET("/debug/routes", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusOK, "Route Debug:\n/settings/company -> Company Settings\n/monitoring -> Monitoring Dashboard\nTime: %s", time.Now().Format("2006-01-02 15:04:05"))
	})
	
	

	// Protected routes - require authentication
	protected := r.Group("/")
	protected.Use(authHandler.AuthMiddleware())
	{
		// Web interface routes
		protected.GET("/dashboard", homeHandler.Dashboard)

		// Job routes
		jobs := protected.Group("/jobs")
		{
			jobs.GET("", jobHandler.ListJobs)
			jobs.GET("/new", jobHandler.NewJobForm)
			jobs.POST("", jobHandler.CreateJob)
			jobs.GET("/:id", jobHandler.GetJob)
			jobs.GET("/:id/edit", jobHandler.EditJobForm)
			jobs.PUT("/:id", jobHandler.UpdateJob)
			jobs.POST("/:id/update", jobHandler.UpdateJob) // Additional POST route for form updates
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
			devices.GET("/:id/stats", deviceHandler.GetDeviceStatsAPI)
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
			scan.GET("", scannerHandler.ScanJobSelection)  // Direct /scan route
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
			// Equipment Packages
			packages := workflow.Group("/packages")
			{
				packages.GET("", equipmentPackageHandler.ShowPackagesList)
				packages.GET("/new", equipmentPackageHandler.ShowPackageForm)
				packages.GET("/:id", equipmentPackageHandler.ShowPackageDetail)
				packages.GET("/:id/edit", equipmentPackageHandler.ShowPackageForm)
				packages.POST("", equipmentPackageHandler.CreatePackage)
				packages.PUT("/:id", equipmentPackageHandler.UpdatePackage)
				packages.DELETE("/:id", equipmentPackageHandler.DeletePackage)
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
			
			// Export routes
			financial.GET("/api/export/transactions", financialHandler.ExportTransactions)
			financial.GET("/api/export/revenue", financialHandler.ExportRevenue)
			financial.GET("/api/export/tax-report", financialHandler.ExportTaxReportCSV)
		}

		// Invoice routes (using NEW fixed invoice system with GoBD compliance)
		invoices := protected.Group("/invoices")
		if complianceMiddleware != nil {
			invoices.Use(complianceMiddleware.InvoiceComplianceMiddleware())
			invoices.Use(complianceMiddleware.DataProcessingMiddleware(compliance.FinancialData, "invoice_management", "Contract performance and accounting", "Art. 6(1)(b) GDPR"))
		}
		{
			// All routes now use the NEW FIXED handler
			invoices.GET("", invoiceHandler.ListInvoices)
			invoices.GET("/new", invoiceHandler.NewInvoiceForm)
			invoices.POST("", invoiceHandler.CreateInvoice)
			invoices.GET("/:id", invoiceHandler.GetInvoice)
			invoices.GET("/:id/edit", invoiceHandler.EditInvoiceForm)
			invoices.GET("/:id/preview", invoiceHandler.PreviewInvoice)
			invoices.PUT("/:id", invoiceHandler.UpdateInvoice)
			invoices.PUT("/:id/status", invoiceHandler.UpdateInvoiceStatus)
			invoices.DELETE("/:id", invoiceHandler.DeleteInvoice)
			// FIXED PDF generation - always returns PDF, never HTML
			invoices.GET("/:id/pdf", invoiceHandler.GenerateInvoicePDF)
		}

		// Invoice template routes - full implementation
		invoiceTemplates := protected.Group("/invoice-templates")
		{
			invoiceTemplates.GET("", templateHandler.ListTemplates)
			invoiceTemplates.GET("/new", templateHandler.NewTemplateForm)
			invoiceTemplates.POST("", templateHandler.CreateTemplate)
			invoiceTemplates.GET("/:id/edit", templateHandler.EditTemplateForm)
			invoiceTemplates.PUT("/:id", templateHandler.UpdateTemplate)
			invoiceTemplates.DELETE("/:id", templateHandler.DeleteTemplate)
			invoiceTemplates.GET("/:id/preview", templateHandler.PreviewTemplate)
			invoiceTemplates.POST("/:id/set-default", templateHandler.SetDefaultTemplate)
		}

		// Company Settings routes - NOW ACTIVE
		settings := protected.Group("/settings")
		{
			// Company settings form route with enhanced debugging
			settings.GET("/company", func(c *gin.Context) {
				log.Printf("🏢 COMPANY SETTINGS ROUTE CALLED - URL: %s, Method: %s", c.Request.URL.Path, c.Request.Method)
				companyHandler.CompanySettingsForm(c)
			})
			settings.POST("/company", companyHandler.UpdateCompanySettingsForm)
			settings.GET("/company/api", companyHandler.GetCompanySettings)
			settings.PUT("/company/api", companyHandler.UpdateCompanySettings)
			settings.POST("/company/logo", companyHandler.UploadCompanyLogo)
			settings.DELETE("/company/logo", companyHandler.DeleteCompanyLogo)
			
		}

		// Security & Admin routes
		security := protected.Group("/security")
		{
			// Web interface routes
			security.GET("/roles", func(c *gin.Context) {
				user, _ := handlers.GetCurrentUser(c)
				c.HTML(http.StatusOK, "security_roles_standalone.html", gin.H{
					"title":       "Role Management",
					"user":        user,
					"currentPage": "security",
				})
			})
			security.GET("/audit", func(c *gin.Context) {
				user, _ := handlers.GetCurrentUser(c)
				c.HTML(http.StatusOK, "security_audit.html", gin.H{
					"title":       "Audit Logs",
					"user":        user,
					"currentPage": "security",
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
				auditAPI.GET("/export", securityHandler.ExportAuditLogs)
			}

			// Permissions API
			permissionsAPI := security.Group("/api/permissions")
			{
				permissionsAPI.GET("", securityHandler.GetPermissions)
				permissionsAPI.GET("/definitions", securityHandler.GetPermissionDefinitionsAPI)
				permissionsAPI.GET("/check", securityHandler.CheckPermission)
			}
		}

		// System Monitoring routes
		monitoring := protected.Group("/monitoring")
		{
			monitoring.GET("", func(c *gin.Context) {
				log.Printf("DEBUG: /monitoring route called - URL: %s, Method: %s", c.Request.URL.Path, c.Request.Method)
				monitoringHandler.Dashboard(c)
			})
			monitoring.GET("/health", monitoringHandler.GetApplicationHealth)
			monitoring.GET("/metrics", monitoringHandler.GetSystemMetrics)
			monitoring.GET("/metrics/prometheus", monitoringHandler.ExportMetrics)
			monitoring.GET("/performance", monitoringHandler.GetPerformanceMetrics)
			monitoring.GET("/errors", monitoringHandler.GetErrorDetails)
			monitoring.POST("/errors/:fingerprint/resolve", monitoringHandler.ResolveError)
			monitoring.POST("/test-error", monitoringHandler.TriggerTestError)
			monitoring.GET("/logs", monitoringHandler.GetLogStream)
		}

		// Compliance routes (GoBD & GDPR)
		if complianceMiddleware != nil {
			compliance := protected.Group("/compliance")
			{
				compliance.GET("/status", complianceMiddleware.GetComplianceStatus())
				compliance.POST("/retention/cleanup", complianceMiddleware.RetentionCleanupMiddleware())
				compliance.POST("/gdpr/request", complianceMiddleware.GDPRRequestMiddleware())
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

		// Profile Settings routes (moved to end to avoid potential conflicts)
		profile := protected.Group("/profile")
		{
			profile.GET("/settings", authHandler.ProfileSettingsForm)
			profile.POST("/settings", authHandler.UpdateProfileSettings)
			profile.GET("/preferences", authHandler.GetUserPreferences)
		}
		

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
				apiJobs.DELETE("/:id/devices/bulk-remove", scannerHandler.BulkRemoveDevices)
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
				apiCases.POST("/:id/devices", caseHandler.ScanDeviceToCase)
				apiCases.DELETE("/:id/devices/:deviceId", caseHandler.RemoveDeviceFromCase)
			}

			// Workflow API
			apiWorkflow := api.Group("/workflow")
			{
				// Equipment Packages API
				apiPackages := apiWorkflow.Group("/packages")
				{
					apiPackages.GET("", equipmentPackageHandler.GetPackages)
					apiPackages.POST("", equipmentPackageHandler.CreatePackage)
					apiPackages.GET("/:id", equipmentPackageHandler.GetPackage)
					apiPackages.PUT("/:id", equipmentPackageHandler.UpdatePackage)
					apiPackages.DELETE("/:id", equipmentPackageHandler.DeletePackage)
					apiPackages.POST("/:id/clone", equipmentPackageHandler.ClonePackage)
					apiPackages.GET("/:id/validate", equipmentPackageHandler.ValidatePackage)
					apiPackages.GET("/:id/stats", equipmentPackageHandler.GetPackageStats)
					apiPackages.GET("/search", equipmentPackageHandler.SearchPackages)
					apiPackages.GET("/categories", equipmentPackageHandler.GetPackageCategories)
					apiPackages.GET("/popular", equipmentPackageHandler.GetPopularPackages)
					apiPackages.GET("/available-devices", equipmentPackageHandler.GetAvailableDevices)
					apiPackages.PUT("/bulk", equipmentPackageHandler.BulkUpdatePackages)
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

			// Invoice API (using NEW fixed invoice system)
			apiInvoices := api.Group("/invoices")
			{
				// All operations use NEW FIXED handler
				apiInvoices.GET("", invoiceHandler.GetInvoicesAPI)
				apiInvoices.POST("", invoiceHandler.CreateInvoice)
				apiInvoices.GET("/:id", invoiceHandler.GetInvoice)
				apiInvoices.PUT("/:id", invoiceHandler.UpdateInvoice)
				apiInvoices.PUT("/:id/status", invoiceHandler.UpdateInvoiceStatus)
				apiInvoices.DELETE("/:id", invoiceHandler.DeleteInvoice)
				apiInvoices.GET("/stats", invoiceHandler.GetInvoiceStatsAPI)
				// New API endpoints for product/device selection
				apiInvoices.GET("/products/:productId", invoiceHandler.GetProductDetails)
				apiInvoices.GET("/products/:productId/devices", invoiceHandler.GetDevicesByProduct)
			}

			// Security API
			apiSecurity := api.Group("/security")
			{
				// Audit API
				auditAPI := apiSecurity.Group("/audit")
				{
					auditAPI.GET("", securityHandler.GetAuditLogs)
					auditAPI.GET("/:id", securityHandler.GetAuditLog)
					auditAPI.GET("/export", securityHandler.ExportAuditLogs)
				}

				// Auth API
				authAPI := apiSecurity.Group("/auth")
				{
					authAPI.GET("/users", authHandler.ListUsersAPI)
				}
			}
		}

		// Additional API routes (outside v1 group for legacy compatibility)
		legacyAPI := protected.Group("/api")
		{
			// Legacy Invoice API (using NEW fixed invoice system)
			legacyAPI.GET("/invoices", invoiceHandler.GetInvoicesAPI)
			legacyAPI.POST("/invoices", invoiceHandler.CreateInvoice)
			legacyAPI.GET("/invoices/:id", invoiceHandler.GetInvoice)
			legacyAPI.PUT("/invoices/:id", invoiceHandler.UpdateInvoice)
			legacyAPI.PUT("/invoices/:id/status", invoiceHandler.UpdateInvoiceStatus)
			legacyAPI.DELETE("/invoices/:id", invoiceHandler.DeleteInvoice)
			legacyAPI.GET("/invoices/stats", invoiceHandler.GetInvoiceStatsAPI)
			
			// Company settings API - NOW ACTIVE
			legacyAPI.GET("/company-settings", companyHandler.GetCompanySettings)
			legacyAPI.PUT("/company-settings", companyHandler.UpdateCompanySettings)
			legacyAPI.POST("/company-settings/logo", companyHandler.UploadCompanyLogo)
			legacyAPI.DELETE("/company-settings/logo", companyHandler.DeleteCompanyLogo)
			
			// SMTP Configuration API
			settingsAPI := legacyAPI.Group("/settings")
			{
				settingsAPI.GET("/smtp", companyHandler.GetSMTPConfig)
				settingsAPI.POST("/smtp", companyHandler.UpdateSMTPConfig)
				settingsAPI.POST("/smtp/test", companyHandler.TestSMTPConnection)
			}
			
			// TODO: Invoice settings API - temporarily disabled
			// Will be re-implemented in new system when needed  
			// legacyAPI.GET("/invoice-settings", invoiceHandler.InvoiceSettingsForm)
			// legacyAPI.PUT("/invoice-settings", invoiceHandler.UpdateInvoiceSettings)
			
			// TODO: Email API - temporarily disabled
			// Will be re-implemented in new system when needed
			// legacyAPI.POST("/test-email", invoiceHandler.TestEmailSettings)

			// Security API - Legacy routes kept for backward compatibility with frontend
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
				apiSecurity.GET("/audit/export", securityHandler.ExportAuditLogs)

				// Permissions API
				apiSecurity.GET("/permissions", securityHandler.GetPermissions)
				apiSecurity.GET("/permissions/definitions", securityHandler.GetPermissionDefinitionsAPI)
				apiSecurity.GET("/permissions/check", securityHandler.CheckPermission)
			}
		}
	}
}

// createDefaultTemplate creates a default invoice template if none exists
func createDefaultTemplate(templateHandler *handlers.InvoiceTemplateHandler, repo *repository.InvoiceRepositoryNew) error {
	// Check if any templates exist
	templates, err := repo.GetAllTemplates()
	if err != nil {
		return fmt.Errorf("failed to check existing templates: %v", err)
	}

	// If templates exist, skip creation
	if len(templates) > 0 {
		log.Printf("Templates already exist (%d), skipping default template creation", len(templates))
		return nil
	}

	// Create a default German standard template
	description := "Standard German invoice template compliant with DIN 5008"
	cssStyles := `{"templateType":"german-din","primaryFont":"Arial","headerFontSize":"18","bodyFontSize":"12","primaryColor":"#2563eb","textColor":"#000000","backgroundColor":"#ffffff","pageMargins":"20","elementSpacing":"15","borderStyle":"solid"}`
	
	defaultTemplate := &models.InvoiceTemplate{
		Name:         "German Standard (DIN 5008)",
		Description:  &description,
		HTMLTemplate: getDefaultTemplateHTML(),
		CSSStyles:    &cssStyles,
		IsDefault:    true,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = repo.CreateTemplate(defaultTemplate)
	if err != nil {
		return fmt.Errorf("failed to create default template: %v", err)
	}

	log.Printf("Successfully created default invoice template")
	return nil
}

// getDefaultTemplateHTML returns the HTML for the default German standard template
func getDefaultTemplateHTML() string {
	return `<div style="padding: 20mm; font-family: Arial, sans-serif; font-size: 12px;">
    <!-- Company Header -->
    <div style="text-align: right; margin-bottom: 30px;">
        <div style="font-weight: bold; font-size: 16px;">{{.company.CompanyName}}</div>
        <div>{{if .company.AddressLine1}}{{.company.AddressLine1}}{{end}}</div>
        <div>{{if .company.PostalCode}}{{.company.PostalCode}} {{end}}{{if .company.City}}{{.company.City}}{{end}}</div>
        <div>{{if .company.Phone}}Tel: {{.company.Phone}}{{end}}</div>
        <div>{{if .company.Email}}E-Mail: {{.company.Email}}{{end}}</div>
    </div>

    <!-- Sender Address Line -->
    <div style="font-size: 8px; margin-bottom: 10px; border-bottom: 1px solid #000; padding-bottom: 5px;">
        {{.company.CompanyName}}, {{if .company.AddressLine1}}{{.company.AddressLine1}}, {{end}}{{if .company.PostalCode}}{{.company.PostalCode}} {{end}}{{if .company.City}}{{.company.City}}{{end}}
    </div>

    <!-- Customer Address -->
    <div style="margin-bottom: 30px;">
        <div style="font-weight: bold;">{{.customer.GetDisplayName}}</div>
        <div>{{if .customer.Street}}{{.customer.Street}}{{if .customer.HouseNumber}} {{.customer.HouseNumber}}{{end}}{{end}}</div>
        <div>{{if .customer.ZIP}}{{.customer.ZIP}} {{end}}{{if .customer.City}}{{.customer.City}}{{end}}</div>
    </div>

    <!-- Invoice Header -->
    <h1 style="font-size: 24px; font-weight: bold; margin-bottom: 20px;">RECHNUNG</h1>

    <!-- Invoice Details -->
    <table style="width: 100%; margin-bottom: 20px; border-collapse: collapse;">
        <tr>
            <td style="width: 30%; padding: 5px 0;">Rechnungsnummer:</td>
            <td style="font-weight: bold;">{{.invoice.InvoiceNumber}}</td>
        </tr>
        <tr>
            <td style="padding: 5px 0;">Rechnungsdatum:</td>
            <td>{{.invoice.IssueDate.Format "02.01.2006"}}</td>
        </tr>
        <tr>
            <td style="padding: 5px 0;">Fälligkeitsdatum:</td>
            <td>{{.invoice.DueDate.Format "02.01.2006"}}</td>
        </tr>
        <tr>
            <td style="padding: 5px 0;">Kundennummer:</td>
            <td>{{.customer.CustomerID}}</td>
        </tr>
    </table>

    <!-- Line Items -->
    <table style="width: 100%; border-collapse: collapse; margin-bottom: 20px;">
        <thead>
            <tr style="background: #f5f5f5;">
                <th style="border: 1px solid #000; padding: 8px; text-align: left;">Pos.</th>
                <th style="border: 1px solid #000; padding: 8px; text-align: left;">Beschreibung</th>
                <th style="border: 1px solid #000; padding: 8px; text-align: center;">Menge</th>
                <th style="border: 1px solid #000; padding: 8px; text-align: right;">Einzelpreis</th>
                <th style="border: 1px solid #000; padding: 8px; text-align: right;">Gesamtpreis</th>
            </tr>
        </thead>
        <tbody>
            {{range $index, $item := .invoice.LineItems}}
            <tr>
                <td style="border: 1px solid #000; padding: 8px;">{{add $index 1}}</td>
                <td style="border: 1px solid #000; padding: 8px;">{{$item.Description}}</td>
                <td style="border: 1px solid #000; padding: 8px; text-align: center;">{{$item.Quantity}}</td>
                <td style="border: 1px solid #000; padding: 8px; text-align: right;">{{printf "%.2f" $item.UnitPrice}} €</td>
                <td style="border: 1px solid #000; padding: 8px; text-align: right;">{{printf "%.2f" $item.TotalPrice}} €</td>
            </tr>
            {{end}}
        </tbody>
    </table>

    <!-- Totals -->
    <div style="text-align: right; margin-bottom: 30px;">
        <table style="width: 200px; margin-left: auto; border-collapse: collapse;">
            <tr>
                <td style="padding: 5px 10px; border-bottom: 1px solid #ddd;">Nettobetrag:</td>
                <td style="text-align: right; padding: 5px 10px; border-bottom: 1px solid #ddd;">{{printf "%.2f" .invoice.Subtotal}} €</td>
            </tr>
            <tr>
                <td style="padding: 5px 10px; border-bottom: 1px solid #ddd;">MwSt. ({{.invoice.TaxRate}}%):</td>
                <td style="text-align: right; padding: 5px 10px; border-bottom: 1px solid #ddd;">{{printf "%.2f" .invoice.TaxAmount}} €</td>
            </tr>
            <tr style="font-weight: bold; border-top: 2px solid #000;">
                <td style="padding: 8px 10px;">Gesamtbetrag:</td>
                <td style="text-align: right; padding: 8px 10px;">{{printf "%.2f" .invoice.TotalAmount}} €</td>
            </tr>
        </table>
    </div>

    <!-- Footer -->
    <div style="font-size: 10px; margin-top: 40px; border-top: 1px solid #ddd; padding-top: 20px;">
        <div style="text-align: center;">
            <div>{{if .company.TaxNumber}}Steuernummer: {{.company.TaxNumber}}{{end}}{{if and .company.TaxNumber .company.VATNumber}} | {{end}}{{if .company.VATNumber}}USt-IdNr.: {{.company.VATNumber}}{{end}}</div>
            <div style="margin-top: 10px;">
                Zahlungsziel: 14 Tage netto | Vielen Dank für Ihr Vertrauen!
            </div>
        </div>
    </div>
</div>`
}