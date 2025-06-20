package services

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go-barcode-webapp/internal/config"
	"go-barcode-webapp/internal/models"
	
	"github.com/jung-kurt/gofpdf"
)

type PDFService struct {
	tempDir   string
	pdfConfig *config.PDFConfig
}

func NewPDFService(pdfConfig *config.PDFConfig) *PDFService {
	tempDir := os.TempDir()
	return &PDFService{
		tempDir:   tempDir,
		pdfConfig: pdfConfig,
	}
}

// GenerateInvoicePDF generates a PDF from an invoice
func (s *PDFService) GenerateInvoicePDF(invoice *models.Invoice, company *models.CompanySettings, settings *models.InvoiceSettings) ([]byte, error) {
	log.Printf("PDFService: Generating PDF for invoice %s with %d line items", invoice.InvoiceNumber, len(invoice.LineItems))
	
	// Create HTML content for the invoice
	htmlContent, err := s.generateInvoiceHTML(invoice, company, settings)
	if err != nil {
		log.Printf("PDFService: Failed to generate HTML: %v", err)
		return nil, fmt.Errorf("failed to generate HTML: %v", err)
	}

	log.Printf("PDFService: Generated HTML content (%d bytes)", len(htmlContent))

	// Convert HTML to PDF
	pdfBytes, err := s.convertHTMLToPDF(htmlContent)
	if err != nil {
		log.Printf("PDFService: Failed to convert HTML to PDF: %v", err)
		return nil, fmt.Errorf("failed to convert HTML to PDF: %v", err)
	}

	log.Printf("PDFService: Generated PDF (%d bytes)", len(pdfBytes))
	return pdfBytes, nil
}

// generateInvoiceHTML creates HTML content for the invoice
func (s *PDFService) generateInvoiceHTML(invoice *models.Invoice, company *models.CompanySettings, settings *models.InvoiceSettings) (string, error) {
	tmplContent := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Invoice {{.Invoice.InvoiceNumber}}</title>
    <style>
        @page {
            size: A4;
            margin: 1cm;
        }
        
        body {
            font-family: Arial, sans-serif;
            font-size: 12px;
            line-height: 1.4;
            color: #333;
            margin: 0;
            padding: 0;
        }
        
        .invoice-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 30px;
            border-bottom: 2px solid #007bff;
            padding-bottom: 20px;
        }
        
        .company-info {
            flex: 1;
        }
        
        .company-logo {
            max-height: 80px;
            margin-bottom: 10px;
        }
        
        .invoice-details {
            text-align: right;
            flex: 1;
        }
        
        .invoice-title {
            font-size: 28px;
            color: #007bff;
            margin-bottom: 10px;
            font-weight: bold;
        }
        
        .invoice-meta table {
            margin-left: auto;
            border-collapse: collapse;
        }
        
        .invoice-meta td {
            padding: 3px 8px;
            border: 1px solid #ddd;
        }
        
        .invoice-meta td:first-child {
            font-weight: bold;
            background-color: #f8f9fa;
        }
        
        .billing-section {
            display: flex;
            justify-content: space-between;
            margin-bottom: 30px;
        }
        
        .bill-to, .job-info {
            flex: 1;
            margin-right: 20px;
        }
        
        .bill-to h3, .job-info h3 {
            margin-bottom: 10px;
            color: #007bff;
            font-size: 14px;
        }
        
        .address-box {
            border: 1px solid #ddd;
            padding: 15px;
            background-color: #f8f9fa;
        }
        
        .line-items {
            margin-bottom: 30px;
        }
        
        .items-table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 20px;
        }
        
        .items-table th {
            background-color: #007bff;
            color: white;
            padding: 10px;
            text-align: left;
            font-weight: bold;
        }
        
        .items-table td {
            padding: 8px 10px;
            border-bottom: 1px solid #ddd;
        }
        
        .items-table tbody tr:nth-child(even) {
            background-color: #f8f9fa;
        }
        
        .text-right {
            text-align: right;
        }
        
        .totals-section {
            display: flex;
            justify-content: space-between;
            margin-bottom: 30px;
        }
        
        .notes {
            flex: 1;
            margin-right: 30px;
        }
        
        .totals {
            flex: 0 0 300px;
        }
        
        .totals-table {
            width: 100%;
            border-collapse: collapse;
        }
        
        .totals-table td {
            padding: 5px 10px;
            border-bottom: 1px solid #ddd;
        }
        
        .totals-table .total-row {
            font-weight: bold;
            font-size: 14px;
            background-color: #007bff;
            color: white;
        }
        
        .footer-info {
            border-top: 1px solid #ddd;
            padding-top: 20px;
            text-align: center;
            font-size: 11px;
            color: #666;
        }
        
        .status-badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 11px;
            font-weight: bold;
            text-transform: uppercase;
        }
        
        .status-draft { background-color: #6c757d; color: white; }
        .status-sent { background-color: #17a2b8; color: white; }
        .status-paid { background-color: #28a745; color: white; }
        .status-overdue { background-color: #dc3545; color: white; }
        .status-cancelled { background-color: #343a40; color: white; }
        
        .overdue-warning {
            background-color: #dc3545;
            color: white;
            padding: 10px;
            text-align: center;
            font-weight: bold;
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
    <!-- Overdue Warning -->
    {{if .Invoice.IsOverdue}}
    <div class="overdue-warning">
        ⚠️ THIS INVOICE IS OVERDUE - IMMEDIATE PAYMENT REQUIRED
    </div>
    {{end}}

    <!-- Invoice Header -->
    <div class="invoice-header">
        <div class="company-info">
            {{if and .Company.LogoPath .Settings.ShowLogoOnInvoice}}
            <img src="{{.Company.LogoPath}}" alt="{{.Company.CompanyName}}" class="company-logo">
            {{end}}
            <h2>{{.Company.CompanyName}}</h2>
            {{if .Company.AddressLine1}}<div>{{.Company.AddressLine1}}</div>{{end}}
            {{if .Company.AddressLine2}}<div>{{.Company.AddressLine2}}</div>{{end}}
            {{if or .Company.City .Company.PostalCode}}
            <div>
                {{if .Company.PostalCode}}{{.Company.PostalCode}} {{end}}{{.Company.City}}
                {{if .Company.State}}, {{.Company.State}}{{end}}
            </div>
            {{end}}
            {{if .Company.Country}}<div>{{.Company.Country}}</div>{{end}}
            {{if .Company.Phone}}<div><strong>Phone:</strong> {{.Company.Phone}}</div>{{end}}
            {{if .Company.Email}}<div><strong>Email:</strong> {{.Company.Email}}</div>{{end}}
            {{if .Company.Website}}<div><strong>Website:</strong> {{.Company.Website}}</div>{{end}}
        </div>
        
        <div class="invoice-details">
            <div class="invoice-title">INVOICE</div>
            <div class="invoice-meta">
                <table>
                    <tr>
                        <td>Invoice #:</td>
                        <td>{{.Invoice.InvoiceNumber}}</td>
                    </tr>
                    <tr>
                        <td>Issue Date:</td>
                        <td>{{.Invoice.IssueDate.Format "02.01.2006"}}</td>
                    </tr>
                    <tr>
                        <td>Due Date:</td>
                        <td>{{.Invoice.DueDate.Format "02.01.2006"}}</td>
                    </tr>
                    <tr>
                        <td>Status:</td>
                        <td><span class="status-badge status-{{.Invoice.Status}}">{{.Invoice.Status}}</span></td>
                    </tr>
                </table>
            </div>
        </div>
    </div>

    <!-- Billing Information -->
    <div class="billing-section">
        <div class="bill-to">
            <h3>Bill To:</h3>
            {{if .Invoice.Customer}}
            <div class="address-box">
                <strong>{{.Invoice.Customer.GetDisplayName}}</strong><br>
                {{if .Invoice.Customer.Email}}{{.Invoice.Customer.Email}}<br>{{end}}
                {{if .Invoice.Customer.PhoneNumber}}{{.Invoice.Customer.PhoneNumber}}<br>{{end}}
                {{if .Invoice.Customer.Street}}{{.Invoice.Customer.Street}}{{if .Invoice.Customer.HouseNumber}} {{.Invoice.Customer.HouseNumber}}{{end}}<br>{{end}}
                {{if .Invoice.Customer.ZIP}}{{.Invoice.Customer.ZIP}} {{end}}{{if .Invoice.Customer.City}}{{.Invoice.Customer.City}}{{end}}
            </div>
            {{else}}
            <div class="address-box">Customer information not available</div>
            {{end}}
        </div>
        
        {{if .Invoice.Job}}
        <div class="job-info">
            <h3>Job Reference:</h3>
            <div class="address-box">
                <strong>{{.Invoice.Job.Description}}</strong><br>
                {{if .Invoice.Job.StartDate}}<small>Start: {{.Invoice.Job.StartDate.Format "02.01.2006"}}</small><br>{{end}}
                {{if .Invoice.Job.EndDate}}<small>End: {{.Invoice.Job.EndDate.Format "02.01.2006"}}</small>{{end}}
            </div>
            
            {{if .Invoice.PaymentTerms}}
            <h3 style="margin-top: 20px;">Payment Terms:</h3>
            <div class="address-box">{{.Invoice.PaymentTerms}}</div>
            {{end}}
        </div>
        {{else if .Invoice.PaymentTerms}}
        <div class="job-info">
            <h3>Payment Terms:</h3>
            <div class="address-box">{{.Invoice.PaymentTerms}}</div>
        </div>
        {{end}}
    </div>

    <!-- Line Items -->
    <div class="line-items">
        <h3>Invoice Items</h3>
        <table class="items-table">
            <thead>
                <tr>
                    <th>Description</th>
                    <th width="10%">Quantity</th>
                    <th width="12%">Unit Price</th>
                    <th width="15%">Rental Period</th>
                    <th width="12%" class="text-right">Total</th>
                </tr>
            </thead>
            <tbody>
                {{if .Invoice.LineItems}}
                {{range .Invoice.LineItems}}
                <tr>
                    <td>
                        <strong>{{.Description}}</strong>
                        {{if eq .ItemType "device"}}
                            {{if .Device}}<br><small>Device: {{.Device.Brand}} {{.Device.Model}} ({{.Device.DeviceID}})</small>{{end}}
                        {{else if eq .ItemType "package"}}
                            {{if .Package}}<br><small>Package: {{.Package.PackageName}}</small>{{end}}
                        {{else if eq .ItemType "service"}}
                            <br><small>Service</small>
                        {{end}}
                    </td>
                    <td>{{printf "%.2f" .Quantity}}</td>
                    <td>{{$.Settings.CurrencySymbol}}{{printf "%.2f" .UnitPrice}}</td>
                    <td>
                        {{if and .RentalStartDate .RentalEndDate}}
                        {{.RentalStartDate.Format "02.01"}} - {{.RentalEndDate.Format "02.01.2006"}}
                        {{if .RentalDays}}<br><small>({{.RentalDays}} days)</small>{{end}}
                        {{else}}
                        -
                        {{end}}
                    </td>
                    <td class="text-right">{{$.Settings.CurrencySymbol}}{{printf "%.2f" .TotalPrice}}</td>
                </tr>
                {{end}}
                {{else}}
                <tr>
                    <td colspan="5" style="text-align: center; padding: 20px; color: #666;">
                        No line items have been added to this invoice yet.
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>

    <!-- Totals and Notes -->
    <div class="totals-section">
        {{if .Invoice.Notes}}
        <div class="notes">
            <h3>Notes:</h3>
            <div class="address-box">{{.Invoice.Notes}}</div>
        </div>
        {{end}}
        
        <div class="totals">
            <table class="totals-table">
                <tr>
                    <td><strong>Subtotal:</strong></td>
                    <td class="text-right">{{.Settings.CurrencySymbol}}{{printf "%.2f" .Invoice.Subtotal}}</td>
                </tr>
                <tr>
                    <td><strong>Tax ({{printf "%.1f" .Invoice.TaxRate}}%):</strong></td>
                    <td class="text-right">{{.Settings.CurrencySymbol}}{{printf "%.2f" .Invoice.TaxAmount}}</td>
                </tr>
                <tr class="total-row">
                    <td><strong>Total Amount:</strong></td>
                    <td class="text-right"><strong>{{.Settings.CurrencySymbol}}{{printf "%.2f" .Invoice.TotalAmount}}</strong></td>
                </tr>
                <tr style="background-color: #ffc107; color: black;">
                    <td><strong>Balance Due:</strong></td>
                    <td class="text-right"><strong>{{.Settings.CurrencySymbol}}{{printf "%.2f" .Invoice.BalanceDue}}</strong></td>
                </tr>
            </table>
        </div>
    </div>

    <!-- Terms and Conditions -->
    {{if .Invoice.TermsConditions}}
    <div style="margin-bottom: 30px;">
        <h3>Terms & Conditions:</h3>
        <div class="address-box" style="font-size: 11px;">{{.Invoice.TermsConditions}}</div>
    </div>
    {{end}}

    <!-- Footer -->
    <div class="footer-info">
        {{if .Company.TaxNumber}}<strong>Tax Number:</strong> {{.Company.TaxNumber}} | {{end}}
        {{if .Company.VATNumber}}<strong>VAT Number:</strong> {{.Company.VATNumber}} | {{end}}
        {{if .Company.Email}}{{.Company.Email}} | {{end}}
        {{if .Company.Website}}{{.Company.Website}}{{end}}
        <br><br>
        <small>Generated on {{.GeneratedAt.Format "02.01.2006 15:04:05"}}</small>
    </div>
</body>
</html>
`

	// Create template
	tmpl, err := template.New("invoice").Parse(tmplContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %v", err)
	}

	// Prepare template data
	data := struct {
		Invoice     *models.Invoice
		Company     *models.CompanySettings
		Settings    *models.InvoiceSettings
		GeneratedAt time.Time
	}{
		Invoice:     invoice,
		Company:     company,
		Settings:    settings,
		GeneratedAt: time.Now(),
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}

// convertHTMLToPDF converts HTML to PDF using available methods
func (s *PDFService) convertHTMLToPDF(htmlContent string) ([]byte, error) {
	// Try different PDF generation methods
	
	// Method 1: Try wkhtmltopdf (if available)
	if pdfBytes, err := s.convertWithWKHTMLToPDF(htmlContent); err == nil {
		return pdfBytes, nil
	}

	// Method 2: Try chrome/chromium headless
	if pdfBytes, err := s.convertWithChrome(htmlContent); err == nil {
		return pdfBytes, nil
	}

	// Method 3: Generate a simple PDF with gofpdf (fallback)
	if pdfBytes, err := s.convertWithGofpdf(htmlContent); err == nil {
		return pdfBytes, nil
	}

	// Method 4: Fallback to simple HTML generation (not PDF, but functional)
	log.Println("Warning: No PDF converter available, returning HTML content")
	return []byte(htmlContent), nil
}

// convertWithWKHTMLToPDF uses wkhtmltopdf to convert HTML to PDF
func (s *PDFService) convertWithWKHTMLToPDF(htmlContent string) ([]byte, error) {
	// Check if wkhtmltopdf is available
	if _, err := exec.LookPath("wkhtmltopdf"); err != nil {
		return nil, fmt.Errorf("wkhtmltopdf not found: %v", err)
	}

	// Create temporary HTML file
	htmlFile := filepath.Join(s.tempDir, fmt.Sprintf("invoice_%d.html", time.Now().UnixNano()))
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write HTML file: %v", err)
	}
	defer os.Remove(htmlFile)

	// Create temporary PDF file
	pdfFile := filepath.Join(s.tempDir, fmt.Sprintf("invoice_%d.pdf", time.Now().UnixNano()))
	defer os.Remove(pdfFile)

	// Execute wkhtmltopdf with config values
	paperSize := s.pdfConfig.PaperSize
	if paperSize == "" {
		paperSize = "A4"
	}
	
	marginTop := s.pdfConfig.Margins["top"]
	if marginTop == "" {
		marginTop = "1cm"
	}
	
	marginBottom := s.pdfConfig.Margins["bottom"]
	if marginBottom == "" {
		marginBottom = "1cm"
	}
	
	marginLeft := s.pdfConfig.Margins["left"]
	if marginLeft == "" {
		marginLeft = "1cm"
	}
	
	marginRight := s.pdfConfig.Margins["right"]
	if marginRight == "" {
		marginRight = "1cm"
	}
	
	cmd := exec.Command("wkhtmltopdf", 
		"--page-size", paperSize,
		"--margin-top", marginTop,
		"--margin-bottom", marginBottom,
		"--margin-left", marginLeft,
		"--margin-right", marginRight,
		"--enable-local-file-access",
		htmlFile, pdfFile)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("wkhtmltopdf failed: %v", err)
	}

	// Read PDF file
	pdfBytes, err := os.ReadFile(pdfFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF file: %v", err)
	}

	return pdfBytes, nil
}

// convertWithChrome uses Chrome/Chromium headless to convert HTML to PDF
func (s *PDFService) convertWithChrome(htmlContent string) ([]byte, error) {
	// Check for Chrome/Chromium
	chromePaths := []string{"google-chrome", "chromium", "chromium-browser", "chrome"}
	var chromePath string
	
	for _, path := range chromePaths {
		if _, err := exec.LookPath(path); err == nil {
			chromePath = path
			break
		}
	}
	
	if chromePath == "" {
		return nil, fmt.Errorf("no Chrome/Chromium found")
	}

	// Create temporary HTML file
	htmlFile := filepath.Join(s.tempDir, fmt.Sprintf("invoice_%d.html", time.Now().UnixNano()))
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write HTML file: %v", err)
	}
	defer os.Remove(htmlFile)

	// Create temporary PDF file
	pdfFile := filepath.Join(s.tempDir, fmt.Sprintf("invoice_%d.pdf", time.Now().UnixNano()))
	defer os.Remove(pdfFile)

	// Execute Chrome headless
	cmd := exec.Command(chromePath,
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--print-to-pdf=" + pdfFile,
		"--print-to-pdf-no-header",
		"file://" + htmlFile)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Chrome headless failed: %v", err)
	}

	// Read PDF file
	pdfBytes, err := os.ReadFile(pdfFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF file: %v", err)
	}

	return pdfBytes, nil
}

// convertWithRod uses Rod (Chrome DevTools Protocol) to convert HTML to PDF
// TODO: Fix Rod API usage - currently commented out due to API changes
/*
func (s *PDFService) convertWithRod(htmlContent string) ([]byte, error) {
	// Implementation temporarily disabled
	return nil, fmt.Errorf("Rod PDF conversion temporarily disabled")
}
*/

// convertWithGofpdf creates a simple PDF using gofpdf library
func (s *PDFService) convertWithGofpdf(htmlContent string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	
	// This is a very basic fallback - just creates a simple PDF with basic text
	pdf.Cell(40, 10, "Invoice Document")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "PDF generation fallback mode")
	pdf.Ln(5)
	pdf.Cell(40, 10, "Original HTML content could not be converted to PDF")
	pdf.Ln(5)
	pdf.Cell(40, 10, "Please contact support for assistance")
	
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF with gofpdf: %v", err)
	}
	
	return buf.Bytes(), nil
}

// SavePDF saves PDF content to a file
func (s *PDFService) SavePDF(content []byte, filepath string) error {
	return os.WriteFile(filepath, content, 0644)
}

// GetPDFReader returns an io.Reader for PDF content
func (s *PDFService) GetPDFReader(content []byte) io.Reader {
	return bytes.NewReader(content)
}