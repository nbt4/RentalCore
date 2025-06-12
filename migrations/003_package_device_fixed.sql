-- ============================================================================
-- Migration: Equipment Packages and Package-Device Relationships (FIXED)
-- Description: Creates equipment packages table and junction table for devices
-- Date: 2025-06-12
-- ============================================================================

-- First, create the equipment_packages table if it doesn't exist
CREATE TABLE IF NOT EXISTS `equipment_packages` (
    `packageID` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(255) NOT NULL,
    `description` TEXT NULL,
    `packageItems` JSON NOT NULL DEFAULT ('[]'),
    `packagePrice` DECIMAL(12,2) NULL,
    `discountPercent` DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    `minRentalDays` INT NOT NULL DEFAULT 1,
    `isActive` BOOLEAN NOT NULL DEFAULT TRUE,
    `createdBy` INT UNSIGNED NULL,
    `createdAt` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updatedAt` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `usageCount` INT NOT NULL DEFAULT 0,
    
    -- Indexes
    INDEX `idx_equipment_packages_active` (`isActive`),
    INDEX `idx_equipment_packages_created_by` (`createdBy`),
    INDEX `idx_equipment_packages_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create the package_devices junction table
CREATE TABLE IF NOT EXISTS `package_devices` (
    `packageID` INT UNSIGNED NOT NULL,
    `deviceID` VARCHAR(255) NOT NULL,
    `quantity` INT UNSIGNED NOT NULL DEFAULT 1,
    `custom_price` DECIMAL(12,2) NULL COMMENT 'Override price for this device in this package',
    `is_required` BOOLEAN NOT NULL DEFAULT FALSE COMMENT 'Whether this device is required or optional',
    `notes` TEXT NULL COMMENT 'Special notes about this device in this package',
    `sort_order` INT UNSIGNED NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Composite primary key
    PRIMARY KEY (`packageID`, `deviceID`),
    
    -- Foreign key constraints
    CONSTRAINT `fk_package_devices_package`
        FOREIGN KEY (`packageID`) 
        REFERENCES `equipment_packages`(`packageID`)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
        
    CONSTRAINT `fk_package_devices_device`
        FOREIGN KEY (`deviceID`) 
        REFERENCES `devices`(`deviceID`)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    
    -- Indexes for performance
    INDEX `idx_package_devices_package` (`packageID`),
    INDEX `idx_package_devices_device` (`deviceID`),
    INDEX `idx_package_devices_required` (`is_required`),
    INDEX `idx_package_devices_sort` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Package Categories Table (optional but useful)
-- ============================================================================

CREATE TABLE IF NOT EXISTS `package_categories` (
    `categoryID` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL,
    `description` TEXT NULL,
    `color` VARCHAR(7) NULL COMMENT 'Hex color code for UI',
    `sort_order` INT UNSIGNED NULL,
    `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Unique constraint
    UNIQUE KEY `uk_package_categories_name` (`name`),
    
    -- Indexes
    INDEX `idx_package_categories_active` (`is_active`),
    INDEX `idx_package_categories_sort` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add category relationship to equipment_packages table (safely)
SET @sql = CONCAT('ALTER TABLE equipment_packages ADD COLUMN IF NOT EXISTS categoryID INT UNSIGNED NULL AFTER description');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add foreign key for category (only if column exists and constraint doesn't exist)
SET @constraint_exists = (
    SELECT COUNT(*) 
    FROM information_schema.TABLE_CONSTRAINTS 
    WHERE CONSTRAINT_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'equipment_packages' 
    AND CONSTRAINT_NAME = 'fk_equipment_packages_category'
);

SET @sql = IF(@constraint_exists = 0,
    'ALTER TABLE equipment_packages ADD CONSTRAINT fk_equipment_packages_category FOREIGN KEY (categoryID) REFERENCES package_categories(categoryID) ON DELETE SET NULL ON UPDATE CASCADE',
    'SELECT "Foreign key constraint already exists" as message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add index for category if it doesn't exist
SET @index_exists = (
    SELECT COUNT(*) 
    FROM information_schema.STATISTICS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'equipment_packages' 
    AND INDEX_NAME = 'idx_equipment_packages_category'
);

SET @sql = IF(@index_exists = 0,
    'ALTER TABLE equipment_packages ADD INDEX idx_equipment_packages_category (categoryID)',
    'SELECT "Index already exists" as message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================================================
-- Insert default package categories
-- ============================================================================

INSERT IGNORE INTO `package_categories` (`name`, `description`, `color`, `sort_order`) VALUES
('Basic', 'Basic equipment packages for standard events', '#007bff', 1),
('Premium', 'Premium packages with high-end equipment', '#28a745', 2),
('Specialized', 'Specialized packages for specific event types', '#ffc107', 3),
('Seasonal', 'Seasonal packages for holidays and special occasions', '#17a2b8', 4),
('Custom', 'Custom packages created for specific customers', '#6c757d', 5);

-- ============================================================================
-- Useful Views for querying package-device relationships
-- ============================================================================

-- View: Package details with device count
CREATE OR REPLACE VIEW `vw_package_summary` AS
SELECT 
    ep.packageID,
    ep.name as packageName,
    ep.description,
    ep.packagePrice,
    ep.discountPercent,
    ep.minRentalDays,
    ep.isActive,
    ep.usageCount,
    pc.name as categoryName,
    COUNT(pd.deviceID) as deviceCount,
    SUM(pd.quantity) as totalDevices,
    COUNT(CASE WHEN pd.is_required = 1 THEN 1 END) as requiredDevices,
    COUNT(CASE WHEN pd.is_required = 0 THEN 1 END) as optionalDevices,
    ep.createdAt,
    ep.updatedAt
FROM equipment_packages ep
LEFT JOIN package_categories pc ON ep.categoryID = pc.categoryID
LEFT JOIN package_devices pd ON ep.packageID = pd.packageID
GROUP BY ep.packageID, ep.name, ep.description, ep.packagePrice, 
         ep.discountPercent, ep.minRentalDays, ep.isActive, 
         ep.usageCount, pc.name, ep.createdAt, ep.updatedAt;

-- View: Package devices with product details
CREATE OR REPLACE VIEW `vw_package_devices_detail` AS
SELECT 
    pd.packageID,
    ep.name as packageName,
    pd.deviceID,
    d.serialNumber,
    d.status as deviceStatus,
    p.name as productName,
    p.category as productCategory,
    p.subcategory as productSubcategory,
    p.itemCostPerDay as defaultPrice,
    pd.custom_price,
    COALESCE(pd.custom_price, p.itemCostPerDay) as effectivePrice,
    pd.quantity,
    pd.is_required,
    pd.notes,
    pd.sort_order,
    (COALESCE(pd.custom_price, p.itemCostPerDay) * pd.quantity) as lineTotal
FROM package_devices pd
INNER JOIN equipment_packages ep ON pd.packageID = ep.packageID
INNER JOIN devices d ON pd.deviceID = d.deviceID
LEFT JOIN products p ON d.productID = p.productID
ORDER BY pd.packageID, pd.sort_order, pd.deviceID;

-- ============================================================================
-- Verification queries (uncomment to run after execution)
-- ============================================================================

-- Check if tables were created successfully
-- DESCRIBE equipment_packages;
-- DESCRIBE package_devices;
-- DESCRIBE package_categories;

-- Show foreign key constraints
-- SELECT 
--     CONSTRAINT_NAME,
--     TABLE_NAME,
--     COLUMN_NAME,
--     REFERENCED_TABLE_NAME,
--     REFERENCED_COLUMN_NAME
-- FROM information_schema.KEY_COLUMN_USAGE 
-- WHERE CONSTRAINT_SCHEMA = DATABASE() 
-- AND TABLE_NAME IN ('package_devices', 'equipment_packages');

-- Show sample data
-- SELECT COUNT(*) as package_count FROM equipment_packages;
-- SELECT COUNT(*) as device_count FROM devices;
-- SELECT * FROM package_categories;