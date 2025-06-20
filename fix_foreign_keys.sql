-- =============================================================================
-- FOREIGN KEY CONSTRAINTS FIX SCRIPT
-- =============================================================================
-- This script adds all missing foreign key constraints to ensure referential integrity
-- Run these commands in sequence on your TS-Lager database

SET FOREIGN_KEY_CHECKS = 0;

-- =============================================================================
-- 1. INVOICES TABLE - Critical Missing Constraints
-- =============================================================================

-- Add foreign key for customer relationship
ALTER TABLE `invoices` 
ADD CONSTRAINT `fk_invoices_customer` 
FOREIGN KEY (`customer_id`) REFERENCES `customers` (`customerID`) 
ON DELETE RESTRICT ON UPDATE CASCADE;

-- Add foreign key for job relationship  
ALTER TABLE `invoices` 
ADD CONSTRAINT `fk_invoices_job` 
FOREIGN KEY (`job_id`) REFERENCES `jobs` (`jobID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

-- Add foreign key for template relationship
ALTER TABLE `invoices` 
ADD CONSTRAINT `fk_invoices_template` 
FOREIGN KEY (`template_id`) REFERENCES `invoice_templates` (`template_id`) 
ON DELETE SET NULL ON UPDATE CASCADE;

-- Add foreign key for created_by user relationship
ALTER TABLE `invoices` 
ADD CONSTRAINT `fk_invoices_created_by` 
FOREIGN KEY (`created_by`) REFERENCES `users` (`userID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

-- =============================================================================
-- 2. INVOICE_LINE_ITEMS TABLE - Missing Device and Package Constraints
-- =============================================================================

-- Add foreign key for device relationship
ALTER TABLE `invoice_line_items` 
ADD CONSTRAINT `fk_invoice_line_items_device` 
FOREIGN KEY (`device_id`) REFERENCES `devices` (`deviceID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

-- Add foreign key for package relationship
ALTER TABLE `invoice_line_items` 
ADD CONSTRAINT `fk_invoice_line_items_package` 
FOREIGN KEY (`package_id`) REFERENCES `equipment_packages` (`packageID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

-- =============================================================================
-- 3. INVOICE_PAYMENTS TABLE - Missing User Constraint
-- =============================================================================

-- Add foreign key for created_by user relationship
ALTER TABLE `invoice_payments` 
ADD CONSTRAINT `fk_invoice_payments_created_by` 
FOREIGN KEY (`created_by`) REFERENCES `users` (`userID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

-- =============================================================================
-- 4. INVOICE_SETTINGS TABLE - Missing User Constraint
-- =============================================================================

-- Add foreign key for updated_by user relationship
ALTER TABLE `invoice_settings` 
ADD CONSTRAINT `fk_invoice_settings_updated_by` 
FOREIGN KEY (`updated_by`) REFERENCES `users` (`userID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

-- =============================================================================
-- 5. INVOICE_TEMPLATES TABLE - Missing User Constraint
-- =============================================================================

-- Add foreign key for created_by user relationship
ALTER TABLE `invoice_templates` 
ADD CONSTRAINT `fk_invoice_templates_created_by` 
FOREIGN KEY (`created_by`) REFERENCES `users` (`userID`) 
ON DELETE SET NULL ON UPDATE CASCADE;

-- =============================================================================
-- 6. OTHER MISSING CONSTRAINTS (Non-Invoice Related)
-- =============================================================================

-- Fix devicestatushistory table (deviceID should reference devices table)
-- Note: Currently has varchar(10) but devices.deviceID is varchar(50)
-- This might need data type adjustment first:
-- ALTER TABLE `devicestatushistory` MODIFY `deviceID` varchar(50);
-- Then add the constraint:
-- ALTER TABLE `devicestatushistory` 
-- ADD CONSTRAINT `fk_devicestatushistory_device` 
-- FOREIGN KEY (`deviceID`) REFERENCES `devices` (`deviceID`) 
-- ON DELETE CASCADE ON UPDATE CASCADE;

-- Fix maintenanceLogs table (deviceID should reference devices table)
-- Note: Currently has int but devices.deviceID is varchar(50)
-- This would need data type adjustment:
-- ALTER TABLE `maintenanceLogs` MODIFY `deviceID` varchar(50);
-- Then add the constraint:
-- ALTER TABLE `maintenanceLogs` 
-- ADD CONSTRAINT `fk_maintenanceLogs_device` 
-- FOREIGN KEY (`deviceID`) REFERENCES `devices` (`deviceID`) 
-- ON DELETE CASCADE ON UPDATE CASCADE;

-- Fix jobdevices table (deviceID should match devices.deviceID length)
-- Note: Currently has varchar(10) but devices.deviceID is varchar(50)
-- This would need data type adjustment:
-- ALTER TABLE `jobdevices` MODIFY `deviceID` varchar(50);
-- The constraint already exists and should work after type fix

SET FOREIGN_KEY_CHECKS = 1;

-- =============================================================================
-- 7. VERIFICATION QUERIES
-- =============================================================================

-- Run these queries after applying the constraints to verify they were added:

-- Check all foreign keys for invoices table
SELECT 
    CONSTRAINT_NAME,
    COLUMN_NAME,
    REFERENCED_TABLE_NAME,
    REFERENCED_COLUMN_NAME
FROM information_schema.KEY_COLUMN_USAGE 
WHERE TABLE_SCHEMA = 'TS-Lager' 
    AND TABLE_NAME = 'invoices' 
    AND REFERENCED_TABLE_NAME IS NOT NULL;

-- Check all foreign keys for invoice_line_items table  
SELECT 
    CONSTRAINT_NAME,
    COLUMN_NAME,
    REFERENCED_TABLE_NAME,
    REFERENCED_COLUMN_NAME
FROM information_schema.KEY_COLUMN_USAGE 
WHERE TABLE_SCHEMA = 'TS-Lager' 
    AND TABLE_NAME = 'invoice_line_items' 
    AND REFERENCED_TABLE_NAME IS NOT NULL;

-- Check for any orphaned records that might prevent constraint creation
-- (Run these BEFORE applying constraints to identify data issues)

-- Check for invoices with invalid customer_id
SELECT i.invoice_id, i.customer_id, 'Invalid customer_id' as issue
FROM invoices i 
LEFT JOIN customers c ON i.customer_id = c.customerID 
WHERE c.customerID IS NULL AND i.customer_id IS NOT NULL;

-- Check for invoices with invalid job_id
SELECT i.invoice_id, i.job_id, 'Invalid job_id' as issue
FROM invoices i 
LEFT JOIN jobs j ON i.job_id = j.jobID 
WHERE j.jobID IS NULL AND i.job_id IS NOT NULL;

-- Check for invoice line items with invalid device_id
SELECT ili.line_item_id, ili.device_id, 'Invalid device_id' as issue
FROM invoice_line_items ili 
LEFT JOIN devices d ON ili.device_id = d.deviceID 
WHERE d.deviceID IS NULL AND ili.device_id IS NOT NULL;

-- =============================================================================
-- END OF SCRIPT
-- =============================================================================