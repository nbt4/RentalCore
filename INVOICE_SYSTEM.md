# RentalCore Invoice System

## 🎉 Complete Invoice Generation System

The RentalCore invoice system is now fully implemented with comprehensive features for professional invoice management, PDF generation, and email delivery.

## ✅ Features Implemented

### Core Invoice Management
- **Create & Edit Invoices** - Full CRUD operations with dynamic line items
- **Invoice Numbering** - Configurable formats and automatic generation
- **Multiple Line Item Types** - Devices, equipment packages, services, and custom items
- **Status Management** - Draft, sent, paid, overdue, cancelled states
- **Payment Tracking** - Partial payments and balance calculations

### PDF Generation
- **Professional PDF Export** - Clean, branded invoice PDFs
- **Company Logo Support** - Automatic logo inclusion in invoices
- **Print-Ready Format** - Optimized for A4 printing
- **Multiple Output Options** - Download or inline preview

### Email Functionality
- **Direct Email Sending** - Send invoices directly to customers
- **PDF Attachments** - Automatic PDF generation and attachment
- **Custom Email Templates** - Professional HTML and text email formats
- **Status Updates** - Automatic status marking when sent

### Configuration & Customization
- **Company Settings** - Logo upload, contact information, branding
- **Invoice Settings** - Number formats, tax rates, payment terms, currency
- **Template Customization** - HTML/CSS based invoice layouts
- **Localization** - Currency symbols, date formats, language options

## 📂 File Structure

```
/opt/dev/go-barcode-webapp/
├── internal/
│   ├── models/
│   │   └── invoice_models.go          # Invoice data models and DTOs
│   ├── repository/
│   │   └── invoice_repository.go      # Database operations and queries
│   ├── handlers/
│   │   └── invoice_handler.go         # Web handlers and API endpoints
│   └── services/
│       ├── pdf_service.go             # PDF generation service
│       └── email_service.go           # Email sending service
├── web/templates/
│   ├── invoices_list.html             # Invoice list and filtering
│   ├── invoice_form.html              # Create/edit invoice form
│   ├── invoice_detail.html            # Invoice view with PDF/email
│   ├── company_settings_form.html     # Company information setup
│   └── invoice_settings_form.html     # Invoice configuration
├── migrations/
│   └── 005_invoice_system.sql         # Database schema and initial data
└── cmd/server/main.go                 # Updated with invoice routes
```

## 🚀 Getting Started

### 1. Database Setup
Execute the migration to create invoice tables:
```sql
-- Run the migration file
source /opt/dev/go-barcode-webapp/migrations/005_invoice_system.sql
```

### 2. Configure Email (Optional)
Set environment variables for email functionality:
```bash
export SMTP_HOST="smtp.gmail.com"
export SMTP_PORT="587"
export SMTP_USERNAME="your-email@gmail.com"
export SMTP_PASSWORD="your-app-password"
export FROM_EMAIL="invoices@yourcompany.com"
export FROM_NAME="Your Company Name"
```

### 3. Install PDF Dependencies (Optional)
For PDF generation, install one of these tools:

**Option A: wkhtmltopdf (Recommended)**
```bash
# Ubuntu/Debian
sudo apt-get install wkhtmltopdf

# macOS
brew install wkhtmltopdf

# CentOS/RHEL
sudo yum install wkhtmltopdf
```

**Option B: Chrome/Chromium**
```bash
# Ubuntu/Debian
sudo apt-get install chromium-browser

# macOS
brew install chromium

# CentOS/RHEL
sudo yum install chromium
```

### 4. Access the System
Navigate to **Tools → Invoices** in the RentalCore interface.

## 🎯 Usage Guide

### Creating Your First Invoice

1. **Set up Company Information**
   - Go to **Tools → Invoices → Settings → Company**
   - Upload your logo and enter company details
   - Configure tax numbers and contact information

2. **Configure Invoice Settings**
   - Go to **Tools → Invoices → Settings → Invoice Settings**
   - Set invoice number format and prefixes
   - Configure default tax rates and payment terms
   - Set currency and date formats

3. **Create an Invoice**
   - Go to **Tools → Invoices → New Invoice**
   - Select customer and job (optional)
   - Add line items (devices, packages, services, or custom)
   - Set dates, tax rates, and terms
   - Save as draft or immediately finalize

4. **Send the Invoice**
   - Open the invoice from the list
   - Click **Email** to send directly to customer
   - Or click **PDF** to download and send manually
   - Mark as sent to track status

### Advanced Features

#### Custom Line Items
- **Devices**: Auto-populate from device inventory with daily rates
- **Equipment Packages**: Use predefined equipment bundles
- **Services**: Add labor, delivery, or other services
- **Custom**: Flexible custom line items

#### Payment Tracking
- Record partial payments with dates and methods
- Automatic balance calculations
- Payment history and references
- Overdue invoice highlighting

#### Reporting and Filtering
- Filter invoices by status, customer, job, date range
- Search functionality across invoice numbers and notes
- Export capabilities for accounting
- Statistics and analytics dashboard

## 🔧 Configuration Options

### Invoice Number Formats
- `{prefix}{year}{month}{sequence:4}` → INV-202512-0001
- `{prefix}{year}-{sequence:4}` → INV-2025-0001
- `{prefix}{sequence:6}` → INV-000001
- `{year}{month}{sequence:4}` → 202512-0001

### Email Templates
The system includes professional email templates with:
- Company branding and logo
- Invoice details and payment information
- Overdue notices and payment reminders
- Customizable subject lines and content

### PDF Styling
Professional PDF layout with:
- Company logo and branding
- Clean table formatting
- Tax and total calculations
- Payment terms and conditions
- Print-optimized design

## 🔍 API Endpoints

### Web Interface
- `GET /invoices` - Invoice list
- `GET /invoices/new` - New invoice form
- `GET /invoices/:id` - Invoice detail view
- `GET /invoices/:id/edit` - Edit invoice form

### API Endpoints
- `POST /api/invoices` - Create invoice
- `PUT /api/invoices/:id` - Update invoice
- `GET /api/invoices/:id/pdf` - Download PDF
- `POST /api/invoices/:id/email` - Send email
- `PUT /api/invoices/:id/status` - Update status

### Settings
- `GET /settings/company` - Company settings form
- `PUT /settings/company` - Update company settings
- `GET /settings/invoices` - Invoice settings form
- `PUT /settings/invoices` - Update invoice settings
- `POST /api/test-email` - Test email configuration

## 🛠️ Troubleshooting

### PDF Generation Issues
1. **No PDF converter found**: Install wkhtmltopdf or Chrome
2. **Permission errors**: Ensure temp directory is writable
3. **Missing fonts**: Install required fonts for PDF rendering

### Email Issues
1. **SMTP not configured**: Set environment variables
2. **Authentication failed**: Check username/password
3. **Connection timeout**: Verify SMTP host and port

### Common Issues
1. **Empty dropdowns**: Ensure customers and devices exist
2. **Calculation errors**: Check tax rate and discount settings
3. **Permission denied**: Verify user authentication

## 🎨 Customization

### Invoice Templates
Edit `pdf_service.go` to customize the HTML template:
- Modify CSS styles for branding
- Add custom fields and calculations
- Adjust layout and formatting
- Include additional company information

### Email Templates
Edit `email_service.go` to customize email content:
- Modify HTML and text email templates
- Add custom variables and formatting
- Include additional customer information
- Customize subject line patterns

## 📈 Integration

The invoice system integrates seamlessly with:
- **Customer Management** - Auto-populate customer details
- **Job Management** - Link invoices to specific jobs
- **Device Inventory** - Include rental equipment with rates
- **Equipment Packages** - Use predefined equipment bundles
- **Financial Tracking** - Monitor payments and outstanding amounts

## 🔒 Security

- **Authentication Required** - All invoice operations require login
- **Input Validation** - Comprehensive data validation and sanitization
- **CSRF Protection** - Forms protected against cross-site request forgery
- **SQL Injection Prevention** - Prepared statements and ORM protection
- **File Upload Security** - Logo uploads with type and size validation

## 💡 Tips for Best Results

1. **Set up company information first** - This ensures professional-looking invoices
2. **Configure email settings** - Test email functionality before sending to customers
3. **Use consistent numbering** - Choose a format and stick with it
4. **Track payments promptly** - Keep accurate records for better cash flow
5. **Regular backups** - Back up invoice data and company settings

---

**🎉 Your professional invoice system is ready!** 

The RentalCore invoice system provides everything you need for professional invoice management, from creation to payment tracking, with PDF generation and email delivery capabilities.