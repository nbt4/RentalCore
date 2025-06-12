-- ============================================================================
-- Package-Device Relationship Migration (CORRECTED FOR EXISTING SCHEMA)
-- Description: Creates package_devices junction table with exact matching types
-- Date: 2025-06-12
-- ============================================================================

-- Create package_devices junction table
-- Using exact types from existing schema:
-- equipment_packages.packageID = INT (not BIGINT)
-- devices.deviceID = VARCHAR(50) (not VARCHAR(255))
CREATE TABLE IF NOT EXISTS `package_devices` (
    `packageID` INT NOT NULL,
    `deviceID` VARCHAR(50) NOT NULL,
    `quantity` INT UNSIGNED NOT NULL DEFAULT 1,
    `custom_price` DECIMAL(12,2) NULL COMMENT 'Override price for this device in package',
    `is_required` TINYINT(1) NOT NULL DEFAULT 0 COMMENT 'Whether device is required (1) or optional (0)',
    `notes` TEXT NULL COMMENT 'Special notes about this device in package',
    `sort_order` INT UNSIGNED NULL COMMENT 'Display order within package',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Composite primary key (package + device combination must be unique)
    PRIMARY KEY (`packageID`, `deviceID`),
    
    -- Foreign key constraints with exact matching types
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================================================
-- Package Categories Table (to organize packages)
-- ============================================================================

CREATE TABLE IF NOT EXISTS `package_categories` (
    `categoryID` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL,
    `description` TEXT NULL,
    `color` VARCHAR(7) NULL COMMENT 'Hex color code for UI (#007bff)',
    `sort_order` INT UNSIGNED NULL,
    `is_active` TINYINT(1) NOT NULL DEFAULT 1,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Constraints
    UNIQUE KEY `uk_package_categories_name` (`name`),
    
    -- Indexes
    INDEX `idx_package_categories_active` (`is_active`),
    INDEX `idx_package_categories_sort` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================================================
-- Add category relationship to equipment_packages table
-- ============================================================================

-- Check if categoryID column already exists before adding
SET @col_exists = (
    SELECT COUNT(*) 
    FROM information_schema.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'equipment_packages' 
    AND COLUMN_NAME = 'categoryID'
);

-- Add categoryID column if it doesn't exist
SET @sql = IF(@col_exists = 0,
    'ALTER TABLE equipment_packages ADD COLUMN categoryID INT NULL AFTER description',
    'SELECT "categoryID column already exists" as message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add foreign key constraint if it doesn't exist
SET @fk_exists = (
    SELECT COUNT(*) 
    FROM information_schema.TABLE_CONSTRAINTS 
    WHERE CONSTRAINT_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'equipment_packages' 
    AND CONSTRAINT_NAME = 'fk_equipment_packages_category'
);

SET @sql = IF(@fk_exists = 0,
    'ALTER TABLE equipment_packages ADD CONSTRAINT fk_equipment_packages_category 
     FOREIGN KEY (categoryID) REFERENCES package_categories(categoryID) 
     ON DELETE SET NULL ON UPDATE CASCADE',
    'SELECT "Category foreign key already exists" as message'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add index if it doesn't exist
SET @idx_exists = (
    SELECT COUNT(*) 
    FROM information_schema.STATISTICS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'equipment_packages' 
    AND INDEX_NAME = 'idx_equipment_packages_category'
);

SET @sql = IF(@idx_exists = 0,
    'ALTER TABLE equipment_packages ADD INDEX idx_equipment_packages_category (categoryID)',
    'SELECT "Category index already exists" as message'
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
-- Create useful views for package management
-- ============================================================================

-- Package summary view with device counts
CREATE OR REPLACE VIEW `vw_package_summary` AS
SELECT 
    ep.packageID,
    ep.name as packageName,
    ep.description,
    ep.package_price as packagePrice,
    ep.discount_percent as discountPercent,
    ep.min_rental_days as minRentalDays,
    ep.is_active as isActive,
    ep.usage_count as usageCount,
    pc.name as categoryName,
    pc.color as categoryColor,
    COUNT(pd.deviceID) as deviceCount,
    SUM(pd.quantity) as totalDevices,
    SUM(CASE WHEN pd.is_required = 1 THEN pd.quantity ELSE 0 END) as requiredDevices,
    SUM(CASE WHEN pd.is_required = 0 THEN pd.quantity ELSE 0 END) as optionalDevices,
    ep.created_at as createdAt,
    ep.updated_at as updatedAt
FROM equipment_packages ep
LEFT JOIN package_categories pc ON ep.categoryID = pc.categoryID
LEFT JOIN package_devices pd ON ep.packageID = pd.packageID
GROUP BY ep.packageID, ep.name, ep.description, ep.package_price, 
         ep.discount_percent, ep.min_rental_days, ep.is_active, 
         ep.usage_count, pc.name, pc.color, ep.created_at, ep.updated_at;

-- Package devices detail view with product information
CREATE OR REPLACE VIEW `vw_package_devices_detail` AS
SELECT 
    pd.packageID,
    ep.name as packageName,
    pd.deviceID,
    d.serialnumber as serialNumber,
    d.status as deviceStatus,
    p.name as productName,
    CONCAT(sc.name, ' > ', sbc.name) as productCategory,
    p.itemcostperday as defaultPrice,
    pd.custom_price,
    COALESCE(pd.custom_price, p.itemcostperday) as effectivePrice,
    pd.quantity,
    pd.is_required,
    pd.notes,
    pd.sort_order,
    (COALESCE(pd.custom_price, p.itemcostperday) * pd.quantity) as lineTotal
FROM package_devices pd
INNER JOIN equipment_packages ep ON pd.packageID = ep.packageID
INNER JOIN devices d ON pd.deviceID = d.deviceID
LEFT JOIN products p ON d.productID = p.productID
LEFT JOIN subcategories sc ON p.subcategoryID = sc.subcategoryID
LEFT JOIN subbiercategories sbc ON p.subbiercategoryID = sbc.subbiercategoryID
ORDER BY pd.packageID, pd.sort_order, pd.deviceID;

-- ============================================================================
-- Verification and information
-- ============================================================================

-- Show successful completion
SELECT 'Package-Device relationship migration completed successfully!' as status;

-- Show current state
SELECT 'CURRENT DATA COUNTS:' as info;
SELECT COUNT(*) as equipment_packages FROM equipment_packages;
SELECT COUNT(*) as devices FROM devices;
SELECT COUNT(*) as package_categories FROM package_categories;
SELECT COUNT(*) as package_devices FROM package_devices;

-- Show table structure for verification
SELECT 'PACKAGE_DEVICES TABLE STRUCTURE:' as info;
DESCRIBE package_devices;