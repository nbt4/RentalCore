-- ============================================================================
-- Package-Device Relationship Migration (SIMPLE VERSION)
-- Description: Creates package_devices junction table - no dynamic SQL
-- Date: 2025-06-12
-- ============================================================================

-- Create package_devices junction table
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
    
    -- Composite primary key
    PRIMARY KEY (`packageID`, `deviceID`),
    
    -- Indexes for performance
    INDEX `idx_package_devices_package` (`packageID`),
    INDEX `idx_package_devices_device` (`deviceID`),
    INDEX `idx_package_devices_required` (`is_required`),
    INDEX `idx_package_devices_sort` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Add foreign key constraints
ALTER TABLE `package_devices` 
ADD CONSTRAINT `fk_package_devices_package`
    FOREIGN KEY (`packageID`) 
    REFERENCES `equipment_packages`(`packageID`)
    ON DELETE CASCADE
    ON UPDATE CASCADE;

ALTER TABLE `package_devices`
ADD CONSTRAINT `fk_package_devices_device`
    FOREIGN KEY (`deviceID`) 
    REFERENCES `devices`(`deviceID`)
    ON DELETE CASCADE
    ON UPDATE CASCADE;

-- Create package_categories table
CREATE TABLE IF NOT EXISTS `package_categories` (
    `categoryID` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL,
    `description` TEXT NULL,
    `color` VARCHAR(7) NULL COMMENT 'Hex color code for UI',
    `sort_order` INT UNSIGNED NULL,
    `is_active` TINYINT(1) NOT NULL DEFAULT 1,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY `uk_package_categories_name` (`name`),
    INDEX `idx_package_categories_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Insert default categories
INSERT IGNORE INTO `package_categories` (`name`, `description`, `color`, `sort_order`) VALUES
('Basic', 'Basic equipment packages', '#007bff', 1),
('Premium', 'Premium equipment packages', '#28a745', 2), 
('Specialized', 'Specialized packages', '#ffc107', 3),
('Custom', 'Custom packages', '#6c757d', 4);

-- Add categoryID to equipment_packages (ignore error if exists)
ALTER TABLE `equipment_packages` ADD COLUMN `categoryID` INT NULL AFTER `description`;

-- Add foreign key (ignore error if exists)
ALTER TABLE `equipment_packages`
ADD CONSTRAINT `fk_equipment_packages_category`
    FOREIGN KEY (`categoryID`) 
    REFERENCES `package_categories`(`categoryID`)
    ON DELETE SET NULL ON UPDATE CASCADE;

-- Add index (ignore error if exists)
ALTER TABLE `equipment_packages` ADD INDEX `idx_equipment_packages_category` (`categoryID`);

-- Create views for easy querying
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
    COUNT(pd.deviceID) as deviceCount,
    SUM(pd.quantity) as totalDevices,
    ep.created_at as createdAt
FROM equipment_packages ep
LEFT JOIN package_categories pc ON ep.categoryID = pc.categoryID
LEFT JOIN package_devices pd ON ep.packageID = pd.packageID
GROUP BY ep.packageID, ep.name, ep.description, ep.package_price, 
         ep.discount_percent, ep.min_rental_days, ep.is_active, 
         ep.usage_count, pc.name, ep.created_at;

-- Show completion status
SELECT 'Migration completed successfully!' as status;
SELECT COUNT(*) as package_devices_count FROM package_devices;
SELECT COUNT(*) as package_categories_count FROM package_categories;