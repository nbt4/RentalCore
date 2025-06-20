# Database Schema Analysis - Foreign Key Issues

## 🔍 Key Naming Inconsistencies Found

### 1. **CRITICAL: invoices table issue**
- `invoices.customer_id` (int) → `customers.customerID` (int) ✅ Types match but naming inconsistent
- `invoices.job_id` (int) → `jobs.jobID` (int) ✅ Types match but naming inconsistent

### 2. **Other Naming Inconsistencies**
- `financial_transactions.customerID` → `customers.customerID` ✅ Consistent
- `jobs.customerID` → `customers.customerID` ✅ Consistent
- `audit_log.userID` → `users.userID` ✅ Consistent

## 📋 Missing Foreign Key Constraints

### **INVOICES TABLE - Missing Constraints**
```sql
-- Missing foreign keys for invoices table
ALTER TABLE `invoices` 
ADD CONSTRAINT `fk_invoices_customer` 
FOREIGN KEY (`customer_id`) REFERENCES `customers` (`customerID`) 
ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE `invoices` 
ADD CONSTRAINT `fk_invoices_job` 
FOREIGN KEY (`job_id`) REFERENCES `jobs` (`jobID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE `invoices` 
ADD CONSTRAINT `fk_invoices_template` 
FOREIGN KEY (`template_id`) REFERENCES `invoice_templates` (`template_id`) 
ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE `invoices` 
ADD CONSTRAINT `fk_invoices_created_by` 
FOREIGN KEY (`created_by`) REFERENCES `users` (`userID`) 
ON DELETE SET NULL ON UPDATE CASCADE;
```

### **INVOICE_LINE_ITEMS TABLE - Missing Constraints**
```sql
-- Missing foreign keys for invoice line items
ALTER TABLE `invoice_line_items` 
ADD CONSTRAINT `fk_invoice_line_items_device` 
FOREIGN KEY (`device_id`) REFERENCES `devices` (`deviceID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE `invoice_line_items` 
ADD CONSTRAINT `fk_invoice_line_items_package` 
FOREIGN KEY (`package_id`) REFERENCES `equipment_packages` (`packageID`) 
ON DELETE SET NULL ON UPDATE CASCADE;
```

### **INVOICE_PAYMENTS TABLE - Missing Constraints**
```sql
-- Missing foreign key for invoice payments
ALTER TABLE `invoice_payments` 
ADD CONSTRAINT `fk_invoice_payments_created_by` 
FOREIGN KEY (`created_by`) REFERENCES `users` (`userID`) 
ON DELETE SET NULL ON UPDATE CASCADE;
```

### **INVOICE_SETTINGS TABLE - Missing Constraints**
```sql
-- Missing foreign key for invoice settings
ALTER TABLE `invoice_settings` 
ADD CONSTRAINT `fk_invoice_settings_updated_by` 
FOREIGN KEY (`updated_by`) REFERENCES `users` (`userID`) 
ON DELETE SET NULL ON UPDATE CASCADE;
```

### **INVOICE_TEMPLATES TABLE - Missing Constraints**
```sql
-- Missing foreign key for invoice templates
ALTER TABLE `invoice_templates` 
ADD CONSTRAINT `fk_invoice_templates_created_by` 
FOREIGN KEY (`created_by`) REFERENCES `users` (`userID`) 
ON DELETE SET NULL ON UPDATE CASCADE;
```

## 🔧 Complete SQL Script to Fix All Missing Foreign Keys

```sql
-- Fix all missing foreign key constraints
-- Run these commands in sequence

-- 1. INVOICES TABLE
ALTER TABLE `invoices` 
ADD CONSTRAINT `fk_invoices_customer` 
FOREIGN KEY (`customer_id`) REFERENCES `customers` (`customerID`) 
ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE `invoices` 
ADD CONSTRAINT `fk_invoices_job` 
FOREIGN KEY (`job_id`) REFERENCES `jobs` (`jobID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE `invoices` 
ADD CONSTRAINT `fk_invoices_template` 
FOREIGN KEY (`template_id`) REFERENCES `invoice_templates` (`template_id`) 
ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE `invoices` 
ADD CONSTRAINT `fk_invoices_created_by` 
FOREIGN KEY (`created_by`) REFERENCES `users` (`userID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

-- 2. INVOICE_LINE_ITEMS TABLE
ALTER TABLE `invoice_line_items` 
ADD CONSTRAINT `fk_invoice_line_items_device` 
FOREIGN KEY (`device_id`) REFERENCES `devices` (`deviceID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE `invoice_line_items` 
ADD CONSTRAINT `fk_invoice_line_items_package` 
FOREIGN KEY (`package_id`) REFERENCES `equipment_packages` (`packageID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

-- 3. INVOICE_PAYMENTS TABLE
ALTER TABLE `invoice_payments` 
ADD CONSTRAINT `fk_invoice_payments_created_by` 
FOREIGN KEY (`created_by`) REFERENCES `users` (`userID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

-- 4. INVOICE_SETTINGS TABLE
ALTER TABLE `invoice_settings` 
ADD CONSTRAINT `fk_invoice_settings_updated_by` 
FOREIGN KEY (`updated_by`) REFERENCES `users` (`userID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

-- 5. INVOICE_TEMPLATES TABLE
ALTER TABLE `invoice_templates` 
ADD CONSTRAINT `fk_invoice_templates_created_by` 
FOREIGN KEY (`created_by`) REFERENCES `users` (`userID`) 
ON DELETE SET NULL ON UPDATE CASCADE;
```

## ✅ Existing Constraints That Are Working

The following relationships already have proper foreign key constraints:
- customers ↔ jobs ✅
- customers ↔ financial_transactions ✅
- devices ↔ products ✅
- devices ↔ insurances ✅
- jobs ↔ jobdevices ✅
- users ↔ sessions ✅
- users ↔ user_roles ✅
- users ↔ roles ✅

## 🎯 Root Cause of Invoice Issues

The customer selection and other invoice issues were likely caused by:
1. **Missing foreign key constraints** - No database-level enforcement
2. **Naming inconsistencies** - `customer_id` vs `customerID` 
3. **No referential integrity** - Invalid references could be inserted

After adding these constraints, the database will:
- ✅ Enforce referential integrity
- ✅ Prevent invalid customer/job references  
- ✅ Cascade updates properly
- ✅ Handle deletions correctly