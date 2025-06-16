package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-barcode-webapp/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AnalyticsHandler struct {
	db *gorm.DB
}

func NewAnalyticsHandler(db *gorm.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

// Dashboard displays the main analytics dashboard
func (h *AnalyticsHandler) Dashboard(c *gin.Context) {
	currentUser, _ := GetCurrentUser(c)
	
	// Get period from query params (default: last 30 days)
	period := c.DefaultQuery("period", "30days")
	
	// Calculate date range
	endDate := time.Now()
	var startDate time.Time
	
	switch period {
	case "7days":
		startDate = endDate.AddDate(0, 0, -7)
	case "30days":
		startDate = endDate.AddDate(0, 0, -30)
	case "90days":
		startDate = endDate.AddDate(0, 0, -90)
	case "1year":
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		startDate = endDate.AddDate(0, 0, -30)
	}

	// Get analytics data
	analytics := h.getAnalyticsData(startDate, endDate)
	
	c.HTML(http.StatusOK, "analytics_dashboard.html", gin.H{
		"title":     "Analytics Dashboard",
		"user":      currentUser,
		"analytics": analytics,
		"period":    period,
		"startDate": startDate.Format("2006-01-02"),
		"endDate":   endDate.Format("2006-01-02"),
	})
}

// getAnalyticsData collects all analytics data for the dashboard
func (h *AnalyticsHandler) getAnalyticsData(startDate, endDate time.Time) map[string]interface{} {
	return map[string]interface{}{
		"revenue":         h.getRevenueAnalytics(startDate, endDate),
		"equipment":       h.getEquipmentAnalytics(startDate, endDate),
		"customers":       h.getCustomerAnalytics(startDate, endDate),
		"jobs":           h.getJobAnalytics(startDate, endDate),
		"topEquipment":   h.getTopEquipment(startDate, endDate, 10),
		"topCustomers":   h.getTopCustomers(startDate, endDate, 10),
		"utilization":    h.getUtilizationMetrics(),
		"trends":         h.getTrendData(startDate, endDate),
	}
}

// getRevenueAnalytics calculates revenue metrics
func (h *AnalyticsHandler) getRevenueAnalytics(startDate, endDate time.Time) map[string]interface{} {
	var totalRevenue, avgJobValue float64
	var totalJobs int64

	// Total revenue and job count
	h.db.Model(&models.Job{}).
		Where("endDate BETWEEN ? AND ? AND final_revenue IS NOT NULL", startDate, endDate).
		Select("COALESCE(SUM(final_revenue), 0) as total, COUNT(*) as count, COALESCE(AVG(final_revenue), 0) as avg").
		Row().Scan(&totalRevenue, &totalJobs, &avgJobValue)

	// Previous period for comparison
	prevStartDate := startDate.AddDate(0, 0, -int(endDate.Sub(startDate).Hours()/24))
	prevEndDate := startDate
	
	var prevRevenue float64
	var prevJobs int64
	h.db.Model(&models.Job{}).
		Where("endDate BETWEEN ? AND ? AND final_revenue IS NOT NULL", prevStartDate, prevEndDate).
		Select("COALESCE(SUM(final_revenue), 0) as total, COUNT(*) as count").
		Row().Scan(&prevRevenue, &prevJobs)

	// Calculate growth rates
	revenueGrowth := float64(0)
	if prevRevenue > 0 {
		revenueGrowth = ((totalRevenue - prevRevenue) / prevRevenue) * 100
	}

	jobsGrowth := float64(0)
	if prevJobs > 0 {
		jobsGrowth = ((float64(totalJobs) - float64(prevJobs)) / float64(prevJobs)) * 100
	}

	return map[string]interface{}{
		"totalRevenue":   totalRevenue,
		"totalJobs":      totalJobs,
		"avgJobValue":    avgJobValue,
		"revenueGrowth":  revenueGrowth,
		"jobsGrowth":     jobsGrowth,
	}
}

// getEquipmentAnalytics calculates equipment metrics
func (h *AnalyticsHandler) getEquipmentAnalytics(startDate, endDate time.Time) map[string]interface{} {
	var totalDevices, activeDevices, maintenanceDevices int64

	// Total devices
	h.db.Model(&models.Device{}).Count(&totalDevices)

	// Active devices (assigned to jobs)
	h.db.Model(&models.Device{}).Where("status IN (?)", []string{"checked out"}).Count(&activeDevices)

	// Devices in maintenance
	h.db.Model(&models.Device{}).Where("status = ?", "maintance").Count(&maintenanceDevices)

	// Utilization rate
	utilizationRate := float64(0)
	if totalDevices > 0 {
		utilizationRate = (float64(activeDevices) / float64(totalDevices)) * 100
	}

	// Revenue per device
	var totalDeviceRevenue float64
	h.db.Raw(`
		SELECT COALESCE(SUM(j.final_revenue), 0)
		FROM jobs j
		INNER JOIN jobdevices jd ON j.jobID = jd.jobID
		WHERE j.endDate BETWEEN ? AND ?
	`, startDate, endDate).Scan(&totalDeviceRevenue)

	revenuePerDevice := float64(0)
	if totalDevices > 0 {
		revenuePerDevice = totalDeviceRevenue / float64(totalDevices)
	}

	return map[string]interface{}{
		"totalDevices":      totalDevices,
		"activeDevices":     activeDevices,
		"maintenanceDevices": maintenanceDevices,
		"utilizationRate":   utilizationRate,
		"revenuePerDevice":  revenuePerDevice,
		"availableDevices":  totalDevices - activeDevices - maintenanceDevices,
	}
}

// getCustomerAnalytics calculates customer metrics
func (h *AnalyticsHandler) getCustomerAnalytics(startDate, endDate time.Time) map[string]interface{} {
	var totalCustomers, activeCustomers, newCustomers int64

	// Total customers
	h.db.Model(&models.Customer{}).Count(&totalCustomers)

	// Active customers (with jobs in period)
	h.db.Model(&models.Customer{}).
		Joins("INNER JOIN jobs ON customers.customerID = jobs.customerID").
		Where("jobs.startDate BETWEEN ? AND ?", startDate, endDate).
		Distinct("customers.customerID").
		Count(&activeCustomers)

	// New customers in period
	h.db.Model(&models.Customer{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&newCustomers)

	// Customer retention rate
	retentionRate := float64(0)
	if totalCustomers > 0 {
		retentionRate = (float64(activeCustomers) / float64(totalCustomers)) * 100
	}

	return map[string]interface{}{
		"totalCustomers":  totalCustomers,
		"activeCustomers": activeCustomers,
		"newCustomers":    newCustomers,
		"retentionRate":   retentionRate,
	}
}

// getJobAnalytics calculates job metrics
func (h *AnalyticsHandler) getJobAnalytics(startDate, endDate time.Time) map[string]interface{} {
	var completedJobs, activeJobs, overdueJobs int64
	var avgJobDuration float64

	// Completed jobs
	h.db.Model(&models.Job{}).
		Where("endDate BETWEEN ? AND ? AND statusID IN (?)", startDate, endDate, []int{3, 4}).
		Count(&completedJobs)

	// Active jobs
	h.db.Model(&models.Job{}).
		Where("startDate <= ? AND (endDate >= ? OR endDate IS NULL) AND statusID IN (?)", 
			endDate, startDate, []int{1, 2}).
		Count(&activeJobs)

	// Overdue jobs
	h.db.Model(&models.Job{}).
		Where("endDate < ? AND statusID NOT IN (?)", time.Now(), []int{3, 4}).
		Count(&overdueJobs)

	// Average job duration
	h.db.Model(&models.Job{}).
		Where("endDate BETWEEN ? AND ? AND startDate IS NOT NULL AND endDate IS NOT NULL", 
			startDate, endDate).
		Select("AVG(DATEDIFF(endDate, startDate))").
		Scan(&avgJobDuration)

	return map[string]interface{}{
		"completedJobs":    completedJobs,
		"activeJobs":       activeJobs,
		"overdueJobs":      overdueJobs,
		"avgJobDuration":   avgJobDuration,
	}
}

// getTopEquipment returns top performing equipment
func (h *AnalyticsHandler) getTopEquipment(startDate, endDate time.Time, limit int) []map[string]interface{} {
	var results []map[string]interface{}

	rows, err := h.db.Raw(`
		SELECT 
			d.deviceID,
			p.name as product_name,
			COUNT(jd.jobID) as rental_count,
			COALESCE(SUM(j.final_revenue), 0) as total_revenue,
			COALESCE(AVG(j.final_revenue), 0) as avg_revenue
		FROM devices d
		LEFT JOIN products p ON d.productID = p.productID
		LEFT JOIN jobdevices jd ON d.deviceID = jd.deviceID
		LEFT JOIN jobs j ON jd.jobID = j.jobID AND j.endDate BETWEEN ? AND ?
		GROUP BY d.deviceID, p.name
		ORDER BY total_revenue DESC
		LIMIT ?
	`, startDate, endDate, limit).Rows()

	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var deviceID, productName string
		var rentalCount int
		var totalRevenue, avgRevenue float64

		rows.Scan(&deviceID, &productName, &rentalCount, &totalRevenue, &avgRevenue)
		
		results = append(results, map[string]interface{}{
			"deviceID":     deviceID,
			"productName":  productName,
			"rentalCount":  rentalCount,
			"totalRevenue": totalRevenue,
			"avgRevenue":   avgRevenue,
		})
	}

	return results
}

// getTopCustomers returns top customers by revenue
func (h *AnalyticsHandler) getTopCustomers(startDate, endDate time.Time, limit int) []map[string]interface{} {
	var results []map[string]interface{}

	rows, err := h.db.Raw(`
		SELECT 
			c.customerID,
			COALESCE(c.companyname, CONCAT(c.firstname, ' ', c.lastname)) as customer_name,
			COUNT(j.jobID) as job_count,
			COALESCE(SUM(j.final_revenue), 0) as total_revenue,
			COALESCE(AVG(j.final_revenue), 0) as avg_revenue
		FROM customers c
		LEFT JOIN jobs j ON c.customerID = j.customerID AND j.endDate BETWEEN ? AND ?
		GROUP BY c.customerID, customer_name
		ORDER BY total_revenue DESC
		LIMIT ?
	`, startDate, endDate, limit).Rows()

	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var customerID int
		var customerName string
		var jobCount int
		var totalRevenue, avgRevenue float64

		rows.Scan(&customerID, &customerName, &jobCount, &totalRevenue, &avgRevenue)
		
		results = append(results, map[string]interface{}{
			"customerID":   customerID,
			"customerName": customerName,
			"jobCount":     jobCount,
			"totalRevenue": totalRevenue,
			"avgRevenue":   avgRevenue,
		})
	}

	return results
}

// getUtilizationMetrics calculates equipment utilization rates
func (h *AnalyticsHandler) getUtilizationMetrics() map[string]interface{} {
	var results []map[string]interface{}

	rows, err := h.db.Raw(`
		SELECT 
			p.name as product_name,
			COUNT(d.deviceID) as total_devices,
			SUM(CASE WHEN d.status = 'checked out' THEN 1 ELSE 0 END) as active_devices,
			ROUND((SUM(CASE WHEN d.status = 'checked out' THEN 1 ELSE 0 END) * 100.0) / COUNT(d.deviceID), 2) as utilization_rate
		FROM products p
		LEFT JOIN devices d ON p.productID = d.productID
		GROUP BY p.productID, p.name
		HAVING COUNT(d.deviceID) > 0
		ORDER BY utilization_rate DESC
	`).Rows()

	if err != nil {
		return map[string]interface{}{"categories": results}
	}
	defer rows.Close()

	for rows.Next() {
		var productName string
		var totalDevices, activeDevices int
		var utilizationRate float64

		rows.Scan(&productName, &totalDevices, &activeDevices, &utilizationRate)
		
		results = append(results, map[string]interface{}{
			"productName":     productName,
			"totalDevices":    totalDevices,
			"activeDevices":   activeDevices,
			"utilizationRate": utilizationRate,
		})
	}

	return map[string]interface{}{
		"categories": results,
	}
}

// getTrendData returns daily/weekly trend data for charts
func (h *AnalyticsHandler) getTrendData(startDate, endDate time.Time) map[string]interface{} {
	// Daily revenue trend
	revenueRows, err := h.db.Raw(`
		SELECT 
			DATE(j.endDate) as date,
			COALESCE(SUM(j.final_revenue), 0) as revenue,
			COUNT(j.jobID) as jobs
		FROM jobs j
		WHERE j.endDate BETWEEN ? AND ?
		GROUP BY DATE(j.endDate)
		ORDER BY date
	`, startDate, endDate).Rows()

	var revenueTrend []map[string]interface{}
	if err == nil {
		defer revenueRows.Close()
		for revenueRows.Next() {
			var date time.Time
			var revenue float64
			var jobs int

			revenueRows.Scan(&date, &revenue, &jobs)
			revenueTrend = append(revenueTrend, map[string]interface{}{
				"date":    date.Format("2006-01-02"),
				"revenue": revenue,
				"jobs":    jobs,
			})
		}
	}

	return map[string]interface{}{
		"revenue": revenueTrend,
	}
}

// GetRevenueAPI returns revenue data as JSON API
func (h *AnalyticsHandler) GetRevenueAPI(c *gin.Context) {
	period := c.DefaultQuery("period", "30days")
	
	endDate := time.Now()
	var startDate time.Time
	
	switch period {
	case "7days":
		startDate = endDate.AddDate(0, 0, -7)
	case "30days":
		startDate = endDate.AddDate(0, 0, -30)
	case "90days":
		startDate = endDate.AddDate(0, 0, -90)
	case "1year":
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		startDate = endDate.AddDate(0, 0, -30)
	}

	analytics := h.getRevenueAnalytics(startDate, endDate)
	c.JSON(http.StatusOK, analytics)
}

// GetEquipmentAPI returns equipment analytics as JSON API
func (h *AnalyticsHandler) GetEquipmentAPI(c *gin.Context) {
	period := c.DefaultQuery("period", "30days")
	
	endDate := time.Now()
	var startDate time.Time
	
	switch period {
	case "7days":
		startDate = endDate.AddDate(0, 0, -7)
	case "30days":
		startDate = endDate.AddDate(0, 0, -30)
	case "90days":
		startDate = endDate.AddDate(0, 0, -90)
	case "1year":
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		startDate = endDate.AddDate(0, 0, -30)
	}

	analytics := h.getEquipmentAnalytics(startDate, endDate)
	c.JSON(http.StatusOK, analytics)
}

// ExportAnalytics exports analytics data to CSV/Excel
func (h *AnalyticsHandler) ExportAnalytics(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	period := c.DefaultQuery("period", "30days")
	
	endDate := time.Now()
	var startDate time.Time
	
	switch period {
	case "7days":
		startDate = endDate.AddDate(0, 0, -7)
	case "30days":
		startDate = endDate.AddDate(0, 0, -30)
	case "90days":
		startDate = endDate.AddDate(0, 0, -90)
	case "1year":
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		startDate = endDate.AddDate(0, 0, -30)
	}

	if format == "csv" {
		h.exportToCSV(c, startDate, endDate)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported format"})
	}
}

// exportToCSV exports analytics data to CSV format
func (h *AnalyticsHandler) exportToCSV(c *gin.Context, startDate, endDate time.Time) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="analytics_`+time.Now().Format("2006-01-02")+`.csv"`)

	// Get analytics data
	analytics := h.getAnalyticsData(startDate, endDate)
	
	// Write CSV headers and data
	csvData := "Metric,Value\n"
	
	// Revenue metrics
	if revenue, ok := analytics["revenue"].(map[string]interface{}); ok {
		csvData += "Total Revenue," + strconv.FormatFloat(revenue["totalRevenue"].(float64), 'f', 2, 64) + "\n"
		csvData += "Total Jobs," + strconv.FormatInt(revenue["totalJobs"].(int64), 10) + "\n"
		csvData += "Average Job Value," + strconv.FormatFloat(revenue["avgJobValue"].(float64), 'f', 2, 64) + "\n"
		if growth, ok := revenue["revenueGrowth"].(float64); ok {
			csvData += "Revenue Growth %," + strconv.FormatFloat(growth, 'f', 1, 64) + "\n"
		}
	}
	
	// Equipment metrics
	if equipment, ok := analytics["equipment"].(map[string]interface{}); ok {
		csvData += "Total Devices," + strconv.FormatInt(equipment["totalDevices"].(int64), 10) + "\n"
		csvData += "Active Devices," + strconv.FormatInt(equipment["activeDevices"].(int64), 10) + "\n"
		csvData += "Utilization Rate %," + strconv.FormatFloat(equipment["utilizationRate"].(float64), 'f', 1, 64) + "\n"
		if revenue, ok := equipment["revenuePerDevice"].(float64); ok {
			csvData += "Revenue per Device," + strconv.FormatFloat(revenue, 'f', 2, 64) + "\n"
		}
	}
	
	// Customer metrics
	if customers, ok := analytics["customers"].(map[string]interface{}); ok {
		csvData += "Total Customers," + strconv.FormatInt(customers["totalCustomers"].(int64), 10) + "\n"
		csvData += "Active Customers," + strconv.FormatInt(customers["activeCustomers"].(int64), 10) + "\n"
		if retention, ok := customers["retentionRate"].(float64); ok {
			csvData += "Customer Retention %," + strconv.FormatFloat(retention, 'f', 1, 64) + "\n"
		}
	}
	
	// Top equipment section
	csvData += "\nTop Equipment by Revenue\n"
	csvData += "Device ID,Product Name,Rental Count,Total Revenue\n"
	if topEquipment, ok := analytics["topEquipment"].([]map[string]interface{}); ok {
		for _, equipment := range topEquipment {
			csvData += fmt.Sprintf("%s,%s,%v,%.2f\n",
				equipment["deviceID"].(string),
				equipment["productName"].(string),
				equipment["rentalCount"],
				equipment["totalRevenue"].(float64),
			)
		}
	}
	
	// Top customers section
	csvData += "\nTop Customers by Revenue\n"
	csvData += "Customer Name,Job Count,Total Revenue\n"
	if topCustomers, ok := analytics["topCustomers"].([]map[string]interface{}); ok {
		for _, customer := range topCustomers {
			csvData += fmt.Sprintf("%s,%v,%.2f\n",
				customer["customerName"].(string),
				customer["jobCount"],
				customer["totalRevenue"].(float64),
			)
		}
	}

	c.String(http.StatusOK, csvData)
}