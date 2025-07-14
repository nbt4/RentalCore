-- Complete Manual Migration Script for TS-Lager Database
-- This script creates all necessary tables WITHOUT foreign key constraints
-- Run this manually to set up the database schema

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET FOREIGN_KEY_CHECKS = 0;
START TRANSACTION;
SET time_zone = "+00:00";

-- --------------------------------------------------------
-- Core Business Tables
-- --------------------------------------------------------

-- Categories table
CREATE TABLE IF NOT EXISTS `categories` (
  `categoryID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `abbreviation` varchar(10) NOT NULL,
  PRIMARY KEY (`categoryID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Subcategories table  
CREATE TABLE IF NOT EXISTS `subcategories` (
  `subcategoryID` varchar(50) NOT NULL,
  `name` varchar(255) NOT NULL,
  `abbreviation` varchar(10) DEFAULT NULL,
  `categoryID` int UNSIGNED DEFAULT NULL,
  PRIMARY KEY (`subcategoryID`),
  KEY `idx_subcategories_category` (`categoryID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Manufacturers table
CREATE TABLE IF NOT EXISTS `manufacturer` (
  `manufacturerID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `website` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`manufacturerID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Brands table
CREATE TABLE IF NOT EXISTS `brands` (
  `brandID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `manufacturerID` int UNSIGNED DEFAULT NULL,
  PRIMARY KEY (`brandID`),
  KEY `idx_brands_manufacturer` (`manufacturerID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Products table
CREATE TABLE IF NOT EXISTS `products` (
  `productID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `categoryID` int UNSIGNED DEFAULT NULL,
  `subcategoryID` varchar(50) DEFAULT NULL,
  `subbiercategoryID` varchar(50) DEFAULT NULL,
  `manufacturerID` int UNSIGNED DEFAULT NULL,
  `brandID` int UNSIGNED DEFAULT NULL,
  `description` text,
  `maintenanceInterval` int UNSIGNED DEFAULT NULL,
  `itemcostperday` decimal(10,2) DEFAULT NULL,
  `weight` decimal(8,2) DEFAULT NULL,
  `height` decimal(8,2) DEFAULT NULL,
  `width` decimal(8,2) DEFAULT NULL,
  `depth` decimal(8,2) DEFAULT NULL,
  `powerconsumption` decimal(8,2) DEFAULT NULL,
  `pos_in_category` int UNSIGNED DEFAULT NULL,
  PRIMARY KEY (`productID`),
  KEY `idx_products_category` (`categoryID`),
  KEY `idx_products_subcategory` (`subcategoryID`),
  KEY `idx_products_manufacturer` (`manufacturerID`),
  KEY `idx_products_brand` (`brandID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Customers table
CREATE TABLE IF NOT EXISTS `customers` (
  `customerID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `companyname` varchar(255) DEFAULT NULL,
  `lastname` varchar(255) DEFAULT NULL,
  `firstname` varchar(255) DEFAULT NULL,
  `street` varchar(255) DEFAULT NULL,
  `housenumber` varchar(20) DEFAULT NULL,
  `ZIP` varchar(20) DEFAULT NULL,
  `city` varchar(255) DEFAULT NULL,
  `federalstate` varchar(255) DEFAULT NULL,
  `country` varchar(255) DEFAULT NULL,
  `phonenumber` varchar(50) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `customertype` varchar(50) DEFAULT NULL,
  `notes` text,
  PRIMARY KEY (`customerID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Status table
CREATE TABLE IF NOT EXISTS `status` (
  `statusID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `status` varchar(255) NOT NULL,
  PRIMARY KEY (`statusID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Job Categories table
CREATE TABLE IF NOT EXISTS `jobCategory` (
  `jobcategoryID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `abbreviation` varchar(10) DEFAULT NULL,
  PRIMARY KEY (`jobcategoryID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Jobs table
CREATE TABLE IF NOT EXISTS `jobs` (
  `jobID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `customerID` int UNSIGNED NOT NULL,
  `statusID` int UNSIGNED NOT NULL,
  `jobcategoryID` int UNSIGNED DEFAULT NULL,
  `description` text,
  `discount` decimal(10,2) DEFAULT 0.00,
  `discount_type` varchar(20) DEFAULT 'amount',
  `revenue` decimal(10,2) DEFAULT 0.00,
  `final_revenue` decimal(10,2) DEFAULT NULL,
  `startDate` date DEFAULT NULL,
  `endDate` date DEFAULT NULL,
  `templateID` int UNSIGNED DEFAULT NULL,
  PRIMARY KEY (`jobID`),
  KEY `idx_jobs_customer` (`customerID`),
  KEY `idx_jobs_status` (`statusID`),
  KEY `idx_jobs_category` (`jobcategoryID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Devices table
CREATE TABLE IF NOT EXISTS `devices` (
  `deviceID` varchar(50) NOT NULL,
  `productID` int UNSIGNED DEFAULT NULL,
  `serialnumber` varchar(255) DEFAULT NULL,
  `purchaseDate` date DEFAULT NULL,
  `lastmaintenance` date DEFAULT NULL,
  `nextmaintenance` date DEFAULT NULL,
  `insurancenumber` varchar(255) DEFAULT NULL,
  `status` varchar(50) DEFAULT 'free',
  `insuranceID` int UNSIGNED DEFAULT NULL,
  `qr_code` varchar(255) DEFAULT NULL,
  `current_location` varchar(255) DEFAULT NULL,
  `gps_latitude` decimal(10,8) DEFAULT NULL,
  `gps_longitude` decimal(11,8) DEFAULT NULL,
  `condition_rating` decimal(3,1) DEFAULT 5.0,
  `usage_hours` decimal(10,2) DEFAULT 0.00,
  `total_revenue` decimal(12,2) DEFAULT 0.00,
  `last_maintenance_cost` decimal(10,2) DEFAULT NULL,
  `notes` text,
  `barcode` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`deviceID`),
  KEY `idx_devices_product` (`productID`),
  KEY `idx_devices_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Job-Device relationship table
CREATE TABLE IF NOT EXISTS `jobdevices` (
  `jobID` int UNSIGNED NOT NULL,
  `deviceID` varchar(50) NOT NULL,
  `custom_price` decimal(10,2) DEFAULT NULL,
  PRIMARY KEY (`jobID`, `deviceID`),
  KEY `idx_jobdevices_job` (`jobID`),
  KEY `idx_jobdevices_device` (`deviceID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Cases table
CREATE TABLE IF NOT EXISTS `cases` (
  `caseID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `description` text,
  `weight` decimal(8,2) DEFAULT NULL,
  `width` decimal(8,2) DEFAULT NULL,
  `height` decimal(8,2) DEFAULT NULL,
  `depth` decimal(8,2) DEFAULT NULL,
  `status` varchar(50) NOT NULL DEFAULT 'free',
  PRIMARY KEY (`caseID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Device-Case relationship table
CREATE TABLE IF NOT EXISTS `devicescases` (
  `caseID` int UNSIGNED NOT NULL,
  `deviceID` varchar(50) NOT NULL,
  PRIMARY KEY (`caseID`, `deviceID`),
  KEY `idx_devicescases_case` (`caseID`),
  KEY `idx_devicescases_device` (`deviceID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------
-- User Management & Authentication Tables
-- --------------------------------------------------------

-- Users table
CREATE TABLE IF NOT EXISTS `users` (
  `userID` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` varchar(100) NOT NULL UNIQUE,
  `email` varchar(255) NOT NULL UNIQUE,
  `password_hash` varchar(255) NOT NULL,
  `first_name` varchar(100) NOT NULL,
  `last_name` varchar(100) NOT NULL,
  `is_active` boolean NOT NULL DEFAULT true,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `last_login` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`userID`),
  KEY `idx_users_username` (`username`),
  KEY `idx_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Sessions table
CREATE TABLE IF NOT EXISTS `sessions` (
  `session_id` varchar(128) NOT NULL,
  `user_id` int UNSIGNED NOT NULL,
  `expires_at` timestamp NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`session_id`),
  KEY `idx_sessions_userid` (`user_id`),
  KEY `idx_sessions_expires` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- User preferences table
CREATE TABLE IF NOT EXISTS `user_preferences` (
  `preference_id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` int UNSIGNED NOT NULL UNIQUE,
  `language` varchar(10) NOT NULL DEFAULT 'de',
  `theme` varchar(20) NOT NULL DEFAULT 'dark',
  `time_zone` varchar(50) NOT NULL DEFAULT 'Europe/Berlin',
  `date_format` varchar(20) NOT NULL DEFAULT 'DD.MM.YYYY',
  `time_format` varchar(10) NOT NULL DEFAULT '24h',
  `email_notifications` boolean NOT NULL DEFAULT true,
  `system_notifications` boolean NOT NULL DEFAULT true,
  `job_status_notifications` boolean NOT NULL DEFAULT true,
  `device_alert_notifications` boolean NOT NULL DEFAULT true,
  `items_per_page` int NOT NULL DEFAULT 25,
  `default_view` varchar(20) NOT NULL DEFAULT 'list',
  `show_advanced_options` boolean NOT NULL DEFAULT false,
  `auto_save_enabled` boolean NOT NULL DEFAULT true,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`preference_id`),
  KEY `idx_user_preferences_userid` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------
-- Invoice & Financial Tables  
-- --------------------------------------------------------

-- Company settings table
CREATE TABLE IF NOT EXISTS `company_settings` (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `company_name` varchar(255) NOT NULL,
  `address_line1` varchar(255) DEFAULT NULL,
  `address_line2` varchar(255) DEFAULT NULL,
  `city` varchar(100) DEFAULT NULL,
  `state` varchar(100) DEFAULT NULL,
  `postal_code` varchar(20) DEFAULT NULL,
  `country` varchar(100) DEFAULT NULL,
  `phone` varchar(50) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `website` varchar(255) DEFAULT NULL,
  `tax_number` varchar(100) DEFAULT NULL,
  `vat_number` varchar(100) DEFAULT NULL,
  `logo_path` varchar(500) DEFAULT NULL,
  `bank_name` varchar(255) DEFAULT NULL,
  `iban` varchar(50) DEFAULT NULL,
  `bic` varchar(20) DEFAULT NULL,
  `account_holder` varchar(255) DEFAULT NULL,
  `ceo_name` varchar(255) DEFAULT NULL,
  `register_court` varchar(255) DEFAULT NULL,
  `register_number` varchar(100) DEFAULT NULL,
  `footer_text` text,
  `payment_terms_text` text,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Invoice templates table
CREATE TABLE IF NOT EXISTS `invoice_templates` (
  `template_id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `description` text,
  `html_template` longtext NOT NULL,
  `css_styles` longtext,
  `is_default` boolean NOT NULL DEFAULT false,
  `is_active` boolean NOT NULL DEFAULT true,
  `created_by` int UNSIGNED DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`template_id`),
  KEY `idx_invoice_templates_created_by` (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Email templates table  
CREATE TABLE IF NOT EXISTS `email_templates` (
  `template_id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `description` text,
  `template_type` enum('invoice','reminder','payment_confirmation','general') NOT NULL DEFAULT 'general',
  `subject` varchar(500) NOT NULL,
  `html_content` longtext NOT NULL,
  `text_content` longtext,
  `is_default` boolean NOT NULL DEFAULT false,
  `is_active` boolean NOT NULL DEFAULT true,
  `created_by` int UNSIGNED DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`template_id`),
  KEY `idx_email_templates_type` (`template_type`),
  KEY `idx_email_templates_created_by` (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Invoices table
CREATE TABLE IF NOT EXISTS `invoices` (
  `invoice_id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `invoice_number` varchar(100) NOT NULL UNIQUE,
  `customer_id` int UNSIGNED NOT NULL,
  `job_id` int UNSIGNED DEFAULT NULL,
  `template_id` int UNSIGNED DEFAULT NULL,
  `status` enum('draft','sent','paid','overdue','cancelled') NOT NULL DEFAULT 'draft',
  `issue_date` date NOT NULL,
  `due_date` date NOT NULL,
  `payment_terms` varchar(500) DEFAULT NULL,
  `subtotal` decimal(12,2) NOT NULL DEFAULT 0.00,
  `tax_rate` decimal(5,2) NOT NULL DEFAULT 0.00,
  `tax_amount` decimal(12,2) NOT NULL DEFAULT 0.00,
  `discount_amount` decimal(12,2) NOT NULL DEFAULT 0.00,
  `total_amount` decimal(12,2) NOT NULL DEFAULT 0.00,
  `paid_amount` decimal(12,2) NOT NULL DEFAULT 0.00,
  `balance_due` decimal(12,2) NOT NULL DEFAULT 0.00,
  `notes` text,
  `terms_conditions` text,
  `internal_notes` text,
  `sent_at` timestamp NULL DEFAULT NULL,
  `paid_at` timestamp NULL DEFAULT NULL,
  `created_by` int UNSIGNED DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`invoice_id`),
  UNIQUE KEY `uk_invoice_number` (`invoice_number`),
  KEY `idx_invoices_customer` (`customer_id`),
  KEY `idx_invoices_job` (`job_id`),
  KEY `idx_invoices_status` (`status`),
  KEY `idx_invoices_dates` (`issue_date`, `due_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Invoice line items table
CREATE TABLE IF NOT EXISTS `invoice_line_items` (
  `line_item_id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `invoice_id` bigint UNSIGNED NOT NULL,
  `item_type` enum('device','service','package','custom') NOT NULL DEFAULT 'custom',
  `device_id` varchar(50) DEFAULT NULL,
  `package_id` int UNSIGNED DEFAULT NULL,
  `description` text NOT NULL,
  `quantity` decimal(10,2) NOT NULL DEFAULT 1.00,
  `unit_price` decimal(12,2) NOT NULL DEFAULT 0.00,
  `total_price` decimal(12,2) NOT NULL DEFAULT 0.00,
  `rental_start_date` date DEFAULT NULL,
  `rental_end_date` date DEFAULT NULL,
  `rental_days` int DEFAULT NULL,
  `sort_order` int UNSIGNED DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`line_item_id`),
  KEY `idx_invoice_line_items_invoice` (`invoice_id`),
  KEY `idx_invoice_line_items_device` (`device_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Invoice payments table
CREATE TABLE IF NOT EXISTS `invoice_payments` (
  `payment_id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `invoice_id` bigint UNSIGNED NOT NULL,
  `amount` decimal(12,2) NOT NULL,
  `payment_method` varchar(100) DEFAULT NULL,
  `payment_date` date NOT NULL,
  `reference_number` varchar(255) DEFAULT NULL,
  `notes` text,
  `created_by` int UNSIGNED DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`payment_id`),
  KEY `idx_invoice_payments_invoice` (`invoice_id`),
  KEY `idx_invoice_payments_date` (`payment_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Invoice settings table
CREATE TABLE IF NOT EXISTS `invoice_settings` (
  `setting_id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `setting_key` varchar(100) NOT NULL UNIQUE,
  `setting_value` text,
  `setting_type` enum('text','number','boolean','json') NOT NULL DEFAULT 'text',
  `description` text,
  `updated_by` int UNSIGNED DEFAULT NULL,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`setting_id`),
  UNIQUE KEY `uk_setting_key` (`setting_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------
-- Additional System Tables
-- --------------------------------------------------------

-- Device assignment history table
CREATE TABLE IF NOT EXISTS `device_assignment_history` (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `deviceID` varchar(50) NOT NULL,
  `jobID` int UNSIGNED DEFAULT NULL,
  `customerID` int UNSIGNED DEFAULT NULL,
  `assignedAt` timestamp NOT NULL,
  `unassignedAt` timestamp NULL DEFAULT NULL,
  `duration` bigint DEFAULT NULL,
  `notes` text,
  `assignedBy` varchar(100) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_device_assignment_device` (`deviceID`),
  KEY `idx_device_assignment_job` (`jobID`),
  KEY `idx_device_assignment_customer` (`customerID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Analytics cache table
CREATE TABLE IF NOT EXISTS `analytics_cache` (
  `cacheID` int NOT NULL AUTO_INCREMENT,
  `metric_name` varchar(100) NOT NULL,
  `period_type` enum('daily','weekly','monthly','yearly') NOT NULL,
  `period_date` date NOT NULL,
  `value` decimal(15,4) DEFAULT NULL,
  `metadata` json DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`cacheID`),
  UNIQUE KEY `uk_analytics_cache` (`metric_name`, `period_type`, `period_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------
-- Insert Default Data
-- --------------------------------------------------------

-- Insert default admin user (password: admin123)
INSERT IGNORE INTO `users` (`userID`, `username`, `email`, `password_hash`, `first_name`, `last_name`, `is_active`) VALUES
(1, 'admin', 'admin@tslager.de', '$2a$10$rQJ8gTbNdCB5pzKS8M7YYeE8HGDQMv8x7vH2J5L6nGkY9vFfKsWsC', 'Admin', 'User', 1);

-- Insert default user preferences for admin
INSERT IGNORE INTO `user_preferences` (`user_id`) VALUES (1);

-- Insert default status values
INSERT IGNORE INTO `status` (`statusID`, `status`) VALUES
(1, 'Geplant'),
(2, 'In Bearbeitung'),
(3, 'Abgeschlossen'),
(4, 'Storniert');

-- Insert default company settings
INSERT IGNORE INTO `company_settings` (`id`, `company_name`) VALUES 
(1, 'TS-Lager GmbH');

-- Insert default invoice settings
INSERT IGNORE INTO `invoice_settings` (`setting_key`, `setting_value`, `setting_type`, `description`) VALUES
('invoice_number_prefix', 'INV-', 'text', 'Prefix for invoice numbers'),
('default_payment_terms', '30', 'number', 'Default payment terms in days'),
('default_tax_rate', '19.00', 'number', 'Default tax rate percentage'),
('currency_symbol', '€', 'text', 'Currency symbol'),
('currency_code', 'EUR', 'text', 'ISO currency code');

-- --------------------------------------------------------
-- Create Indexes for Performance
-- --------------------------------------------------------

-- Additional performance indexes
CREATE INDEX IF NOT EXISTS `idx_jobs_dates` ON `jobs` (`startDate`, `endDate`);
CREATE INDEX IF NOT EXISTS `idx_devices_location` ON `devices` (`current_location`);
CREATE INDEX IF NOT EXISTS `idx_customers_email` ON `customers` (`email`);
CREATE INDEX IF NOT EXISTS `idx_customers_company` ON `customers` (`companyname`);

COMMIT;
SET FOREIGN_KEY_CHECKS = 1;

-- --------------------------------------------------------
-- SCRIPT COMPLETION
-- --------------------------------------------------------

SELECT 'Migration script completed successfully!' as Status;
SELECT COUNT(*) as user_count FROM users;
SELECT COUNT(*) as session_count FROM sessions;
SELECT COUNT(*) as email_template_count FROM email_templates;