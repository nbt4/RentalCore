-- ============================================================================
-- Minimal Package-Device Relationship Table
-- Description: Core junction table for connecting equipment packages with devices
-- Date: 2025-06-12
-- ============================================================================

-- Create the package_devices junction table
CREATE TABLE IF NOT EXISTS `package_devices` (
    `packageID` INT UNSIGNED NOT NULL,
    `deviceID` VARCHAR(255) NOT NULL,
    `quantity` INT UNSIGNED NOT NULL DEFAULT 1,
    `custom_price` DECIMAL(12,2) NULL COMMENT 'Override price for this device in this package',
    `is_required` BOOLEAN NOT NULL DEFAULT FALSE COMMENT 'Whether this device is required or optional',
    `notes` TEXT NULL COMMENT 'Special notes about this device in this package',
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
    INDEX `idx_package_devices_device` (`deviceID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Quick verification queries (run after executing the above)
-- ============================================================================

-- Check if table was created successfully
-- DESCRIBE package_devices;

-- Check existing packages (should show empty initially)
-- SELECT packageID, name FROM equipment_packages;

-- Check existing devices (to see what devices are available)
-- SELECT deviceID, serialNumber, status FROM devices LIMIT 10;