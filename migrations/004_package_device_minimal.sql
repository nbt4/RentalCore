-- ============================================================================
-- MINIMAL Package-Device Relationship Migration
-- Description: Creates package_devices table with exact types from your schema
-- Date: 2025-06-12
-- ============================================================================

-- Create package_devices junction table
CREATE TABLE IF NOT EXISTS `package_devices` (
    `packageID` INT NOT NULL,
    `deviceID` VARCHAR(50) NOT NULL,
    `quantity` INT UNSIGNED NOT NULL DEFAULT 1,
    `custom_price` DECIMAL(12,2) NULL,
    `is_required` TINYINT(1) NOT NULL DEFAULT 0,
    `notes` TEXT NULL,
    `sort_order` INT UNSIGNED NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`packageID`, `deviceID`)
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

-- Show completion
SELECT 'Migration completed successfully!' as status;