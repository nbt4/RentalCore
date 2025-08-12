# JobScanner Pro - Go Web Application

A comprehensive job and equipment management system with barcode/QR code scanning capabilities, converted from Python to Go web application.

## Features

### Core Functionality
- **Job Management**: Complete CRUD operations for jobs with customer assignments
- **Device Management**: Equipment inventory with availability tracking
- **Customer Management**: Customer database with job history
- **Barcode/QR Code Generation**: Generate QR codes and barcodes for devices
- **Bulk Device Assignment**: Scan multiple devices to jobs at once
- **Device Tracking**: Real-time availability and assignment status

### Web Interface
- **Responsive Design**: Bootstrap-based responsive web interface
- **Dark/Light Theme**: Configurable theme switching
- **Real-time Updates**: Dynamic content updates without page refresh
- **Mobile-Friendly**: Optimized for mobile and tablet devices

### Technical Features
- **RESTful API**: Complete REST API for all operations
- **MySQL Database**: Robust relational database with proper relationships
- **Connection Pooling**: Efficient database connection management
- **Configuration Management**: JSON-based configuration system
- **Error Handling**: Comprehensive error handling and logging

## Project Structure

```
go-barcode-webapp/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP handlers (controllers)
│   ├── models/          # Data models and structures
│   ├── repository/      # Database access layer
│   └── services/        # Business logic services
├── web/
│   ├── templates/       # HTML templates
│   └── static/          # CSS, JS, images
├── migrations/          # Database migration scripts
├── docs/               # Documentation
├── go.mod              # Go module dependencies
├── config.json         # Application configuration
└── README.md           # This file
```

## Installation & Setup

### Prerequisites
- Go 1.21 or higher
- MySQL 8.0 or higher
- Git

### 1. Clone and Setup

```bash
cd /opt/dev/go-barcode-webapp
go mod tidy
```

### 2. Database Setup

```bash
# Create database and run migrations
mysql -u root -p < migrations/001_initial_schema.sql
```

### 3. Configuration

Edit `config.json` to match your environment:

```json
{
  "database": {
    "host": "localhost",
    "port": 3306,
    "database": "jobscanner",
    "username": "root",
    "password": "your_password",
    "pool_size": 5
  },
  "server": {
    "port": 8080,
    "host": "localhost"
  }
}
```

### 4. Run the Application

**Easy startup with the provided script:**
```bash
./start.sh
```

**Or manually:**
```bash
export PATH=$PATH:~/go/bin
go build -o jobscanner cmd/server/main.go
./jobscanner
```

The application will be available at:
- **Main Dashboard**: http://localhost:8080/jobs
- **Device Management**: http://localhost:8080/devices  
- **Customer Management**: http://localhost:8080/customers
- **REST API**: http://localhost:8080/api/v1/

## API Endpoints

### Jobs
- `GET /api/v1/jobs` - List all jobs with filtering
- `POST /api/v1/jobs` - Create new job
- `GET /api/v1/jobs/{id}` - Get job details
- `PUT /api/v1/jobs/{id}` - Update job
- `DELETE /api/v1/jobs/{id}` - Delete job
- `POST /api/v1/jobs/{id}/devices/{deviceId}` - Assign device to job
- `DELETE /api/v1/jobs/{id}/devices/{deviceId}` - Remove device from job
- `POST /api/v1/jobs/{id}/bulk-scan` - Bulk assign devices

### Devices
- `GET /api/v1/devices` - List all devices with filtering
- `POST /api/v1/devices` - Create new device
- `GET /api/v1/devices/{id}` - Get device details
- `PUT /api/v1/devices/{id}` - Update device
- `DELETE /api/v1/devices/{id}` - Delete device
- `GET /api/v1/devices/available` - Get available devices

### Customers
- `GET /api/v1/customers` - List all customers
- `POST /api/v1/customers` - Create new customer
- `GET /api/v1/customers/{id}` - Get customer details
- `PUT /api/v1/customers/{id}` - Update customer
- `DELETE /api/v1/customers/{id}` - Delete customer

### Barcodes
- `GET /barcodes/device/{serialNo}/qr` - Generate QR code for device
- `GET /barcodes/device/{serialNo}/barcode` - Generate barcode for device

## Web Interface

### Main Pages
- `/` - Dashboard (redirects to jobs)
- `/jobs` - Jobs list and management
- `/jobs/{id}` - Job details with device assignment
- `/devices` - Device inventory management
- `/customers` - Customer management

### Features
- **Job Management**: Create, edit, delete jobs with full device assignment
- **Device Scanning**: Bulk device assignment with validation
- **Barcode Generation**: Generate and display QR codes and barcodes
- **Filtering**: Advanced filtering for jobs and devices
- **Responsive Design**: Works on desktop, tablet, and mobile devices
- **Theme Support**: Dark and light theme switching

## Database Schema

### Main Tables
- **customers** - Customer information
- **statuses** - Job status definitions
- **jobs** - Job records with customer and status relationships
- **devices** - Equipment inventory
- **products** - Product catalog
- **job_devices** - Many-to-many relationship between jobs and devices

### Key Features
- Proper foreign key relationships
- Soft deletion for jobs
- Device availability tracking
- Revenue calculation and tracking
- Audit trails with created/updated timestamps

## Development

### Adding New Features
1. Create models in `internal/models/`
2. Add repository methods in `internal/repository/`
3. Implement business logic in `internal/services/`
4. Create handlers in `internal/handlers/`
5. Add routes in `cmd/server/main.go`
6. Create templates in `web/templates/`

### Testing
```bash
go test ./...
```

### Building for Production
```bash
go build -o jobscanner cmd/server/main.go
```

## Comparison with Python Version

### Converted Features
✅ **Complete**: Job management, device management, customer management  
✅ **Complete**: Database operations with MySQL  
✅ **Complete**: Barcode/QR code generation  
✅ **Complete**: Web interface with responsive design  
✅ **Complete**: Bulk device scanning and assignment  
✅ **Complete**: Configuration management  
✅ **Complete**: RESTful API endpoints  

### Planned Features
🚧 **Pending**: Camera integration for live barcode scanning  
🚧 **Pending**: PDF report generation  
🚧 **Pending**: CSV import/export functionality  
🚧 **Pending**: Advanced analytics and filtering  

### Improvements over Python Version
- **Better Performance**: Go's compiled nature provides better performance
- **Better Concurrency**: Native goroutines for handling concurrent requests
- **Simpler Deployment**: Single binary deployment vs Python dependencies
- **Type Safety**: Compile-time type checking
- **Modern Web Interface**: Bootstrap 5 with responsive design
- **RESTful API**: Complete REST API for integration

## License

This project is converted from the original Python JobScanner Pro application and maintains the same functionality while providing a modern web-based interface.