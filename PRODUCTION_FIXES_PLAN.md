# Production Fixes Plan

## Critical Issues Identified (Priority: HIGH)

### 1. Routing & Template Issues
- [ ] **Templates route** redirects to role manager instead of job templates
- [ ] **View Documents** redirects to role page instead of documents list
- [ ] **Scanner launch** redirects to role section instead of scanner
- [ ] Multiple routes pointing to wrong templates - systematic review needed

### 2. Search & Analytics Issues  
- [ ] **Global search** returns JSON error "Search query is required"
- [ ] **Export buttons** under analytics not working
- [ ] Need proper search handler implementation

### 3. Transaction Management
- [ ] **New transaction page** - Save button not working
- [ ] Transaction form not submitting properly
- [ ] Need transaction creation handler

### 4. Case-Device Assignment Issues
- [ ] **Devices not displayed** in drag&drop zone for case assignment
- [ ] **Case shows 0 devices** when devices are assigned
- [ ] Need to fix case-device relationship display

## Feature Implementation (Priority: MEDIUM)

### 5. Role Management Improvements
- [ ] **Readable permission names** instead of technical names
- [ ] **Hover tooltips** explaining what each permission does
- [ ] **Dedicated role management card** in admin dashboard
- [ ] Better role assignment interface

### 6. Missing Features
- [ ] **Invoice generation** (currently shows "coming soon")
- [ ] **Document management** system
- [ ] **Advanced analytics** with working exports

## Technical Tasks

### Route Mapping Review
```
Current Issues:
/workflow/templates -> security_roles.html (WRONG)
/documents -> security_roles.html (WRONG) 
/mobile/scanner -> security_roles.html (WRONG)
/search -> JSON error response

Should be:
/workflow/templates -> job_templates_list.html
/documents -> documents_list.html  
/mobile/scanner -> mobile_scanner.html
/search -> search_results.html
```

### Template Files Needed
- [ ] `search_results.html` - Global search results page
- [ ] `documents_list.html` - Document management page  
- [ ] `job_templates_list.html` - Job templates listing
- [ ] `analytics_export.html` - Analytics export functionality
- [ ] `role_management_card.html` - Dedicated role management widget

### Handler Implementation
- [ ] Search handler with proper query processing
- [ ] Document management handlers
- [ ] Transaction creation/editing handlers  
- [ ] Analytics export handlers
- [ ] Invoice generation system

### Database Queries
- [ ] Fix case-device relationship queries
- [ ] Optimize search performance
- [ ] Add proper transaction storage

## UI/UX Improvements

### Permission Names Mapping
```javascript
Technical -> Readable
"users.manage" -> "Manage Users" (tooltip: "Create, edit, and delete user accounts")
"jobs.manage" -> "Manage Jobs" (tooltip: "Full access to job creation and management")
"devices.manage" -> "Manage Equipment" (tooltip: "Add, edit, and track equipment inventory")
"customers.manage" -> "Manage Customers" (tooltip: "Create and edit customer information")
"reports.view" -> "View Reports" (tooltip: "Access analytics and generate reports")
"settings.manage" -> "System Settings" (tooltip: "Configure application settings")
"scan.use" -> "Use Scanner" (tooltip: "Access mobile barcode scanning features")
```

### Role Management Card Design
- Card in admin dashboard
- Quick role assignment
- Permission overview
- Audit trail display

## Implementation Order
1. **Fix routing issues** (Templates, Documents, Scanner)
2. **Implement search functionality**
3. **Fix transaction management**
4. **Resolve case-device assignment**
5. **Improve role management UI**
6. **Add missing features** (invoices, exports)

## Files to Modify
- `cmd/server/main.go` - Route mappings
- `internal/handlers/` - Missing handlers
- `web/templates/` - New template files
- `web/static/js/app.js` - Frontend functionality
- `internal/repository/` - Database queries

## Testing Checklist
- [ ] All navigation links work correctly
- [ ] Search returns proper results
- [ ] Export functions download files
- [ ] Transaction forms save data
- [ ] Case-device assignment displays correctly
- [ ] Role management is intuitive
- [ ] Mobile scanner launches properly

---
*Created: 2025-06-11*
*Status: Planning Phase*