-- ============================================================================
-- Simple Package-Device Relationship Migration  
-- Description: Creates package_devices junction table (no dynamic SQL)
-- Date: 2025-06-12
-- ============================================================================

-- Create package_devices table
-- Using BIGINT UNSIGNED to match what GORM typically creates for uint in Go
CREATE TABLE IF NOT EXISTS `package_devices` (
    `packageID` BIGINT UNSIGNED NOT NULL,
    `deviceID` VARCHAR(255) NOT NULL,
    `quantity` INT UNSIGNED NOT NULL DEFAULT 1,
    `custom_price` DECIMAL(12,2) NULL COMMENT 'Override price for this device in package',
    `is_required` BOOLEAN NOT NULL DEFAULT FALSE COMMENT 'Whether device is required or optional',
    `notes` TEXT NULL COMMENT 'Special notes about this device in package',
    `sort_order` INT UNSIGNED NULL COMMENT 'Display order within package',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Composite primary key (package + device combination must be unique)
    PRIMARY KEY (`packageID`, `deviceID`),
    
    -- Indexes for performance
    INDEX `idx_package_devices_package` (`packageID`),
    INDEX `idx_package_devices_device` (`deviceID`),
    INDEX `idx_package_devices_required` (`is_required`),
    INDEX `idx_package_devices_sort` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add foreign key constraints
-- Package foreign key
ALTER TABLE `package_devices` 
ADD CONSTRAINT `fk_package_devices_package`
    FOREIGN KEY (`packageID`) 
    REFERENCES `equipment_packages`(`packageID`)
    ON DELETE CASCADE
    ON UPDATE CASCADE;

-- Device foreign key  
ALTER TABLE `package_devices`
ADD CONSTRAINT `fk_package_devices_device`
    FOREIGN KEY (`deviceID`) 
    REFERENCES `devices`(`deviceID`)
    ON DELETE CASCADE
    ON UPDATE CASCADE;

-- ============================================================================
-- Package Categories Table (optional but useful)
-- ============================================================================

CREATE TABLE IF NOT EXISTS `package_categories` (
    `categoryID` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL,
    `description` TEXT NULL,
    `color` VARCHAR(7) NULL COMMENT 'Hex color code for UI (#007bff)',
    `sort_order` INT UNSIGNED NULL,
    `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Constraints
    UNIQUE KEY `uk_package_categories_name` (`name`),
    
    -- Indexes
    INDEX `idx_package_categories_active` (`is_active`),
    INDEX `idx_package_categories_sort` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default package categories
INSERT IGNORE INTO `package_categories` (`name`, `description`, `color`, `sort_order`) VALUES
('Basic', 'Basic equipment packages for standard events', '#007bff', 1),
('Premium', 'Premium packages with high-end equipment', '#28a745', 2), 
('Specialized', 'Specialized packages for specific event types', '#ffc107', 3),
('Seasonal', 'Seasonal packages for holidays and occasions', '#17a2b8', 4),
('Custom', 'Custom packages created for specific customers', '#6c757d', 5);

-- ============================================================================
-- Add category relationship to equipment_packages (if column doesn't exist)
-- Note: This might fail if column already exists - that's OK
-- ============================================================================

-- Try to add categoryID column (ignore error if exists)
ALTER TABLE `equipment_packages` 
ADD COLUMN `categoryID` INT UNSIGNED NULL AFTER `description`;

-- Try to add foreign key (ignore error if exists)
ALTER TABLE `equipment_packages`
ADD CONSTRAINT `fk_equipment_packages_category`
    FOREIGN KEY (`categoryID`) 
    REFERENCES `package_categories`(`categoryID`)
    ON DELETE SET NULL
    ON UPDATE CASCADE;

-- Try to add index (ignore error if exists)  
ALTER TABLE `equipment_packages`
ADD INDEX `idx_equipment_packages_category` (`categoryID`);

-- ============================================================================
-- Create useful views
-- ============================================================================

-- Package summary view
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
    SUM(CASE WHEN pd.is_required = 1 THEN pd.quantity ELSE 0 END) as requiredDevices,
    SUM(CASE WHEN pd.is_required = 0 THEN pd.quantity ELSE 0 END) as optionalDevices,
    ep.createdAt,
    ep.updatedAt
FROM equipment_packages ep
LEFT JOIN package_categories pc ON ep.categoryID = pc.categoryID
LEFT JOIN package_devices pd ON ep.packageID = pd.packageID
GROUP BY ep.packageID, ep.name, ep.description, ep.packagePrice, 
         ep.discountPercent, ep.minRentalDays, ep.isActive, 
         ep.usageCount, pc.name, ep.createdAt, ep.updatedAt;

-- Package devices detail view
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
-- Show results (for verification)
-- ============================================================================

-- Show what was created
SELECT 'Migration completed successfully!' as status;
SELECT COUNT(*) as packages_count FROM equipment_packages;
SELECT COUNT(*) as devices_count FROM devices LIMIT 1;
SELECT COUNT(*) as categories_count FROM package_categories;

-- Show table structure
DESCRIBE package_devices;