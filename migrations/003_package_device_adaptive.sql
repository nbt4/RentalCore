-- ============================================================================
-- Adaptive Package-Device Relationship Migration
-- Description: Automatically detects packageID data type and creates matching tables
-- Date: 2025-06-12
-- ============================================================================

-- First, let's check what data type the existing equipment_packages.packageID has
SET @packageid_type = (
    SELECT COLUMN_TYPE 
    FROM information_schema.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'equipment_packages' 
    AND COLUMN_NAME = 'packageID'
);

-- Show the detected type for debugging
SELECT CONCAT('Detected packageID type: ', COALESCE(@packageid_type, 'TABLE NOT FOUND')) AS debug_info;

-- Check if equipment_packages table exists
SET @table_exists = (
    SELECT COUNT(*) 
    FROM information_schema.TABLES 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'equipment_packages'
);

-- If table doesn't exist, create it first
SET @create_packages_sql = IF(@table_exists = 0,
    'CREATE TABLE equipment_packages (
        packageID BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        description TEXT NULL,
        packageItems JSON NOT NULL DEFAULT (JSON_ARRAY()),
        packagePrice DECIMAL(12,2) NULL,
        discountPercent DECIMAL(5,2) NOT NULL DEFAULT 0.00,
        minRentalDays INT NOT NULL DEFAULT 1,
        isActive BOOLEAN NOT NULL DEFAULT TRUE,
        createdBy BIGINT UNSIGNED NULL,
        createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updatedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        usageCount INT NOT NULL DEFAULT 0,
        INDEX idx_equipment_packages_active (isActive),
        INDEX idx_equipment_packages_name (name)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci',
    'SELECT "Equipment packages table already exists" as message'
);

PREPARE stmt FROM @create_packages_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Re-detect the packageID type after potential table creation
SET @packageid_type = (
    SELECT COLUMN_TYPE 
    FROM information_schema.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'equipment_packages' 
    AND COLUMN_NAME = 'packageID'
);

-- Create package_devices table with matching packageID type
-- Use BIGINT UNSIGNED if the original is bigint, otherwise use INT UNSIGNED
SET @package_devices_sql = CONCAT(
    'CREATE TABLE IF NOT EXISTS package_devices (',
    'packageID ', COALESCE(@packageid_type, 'BIGINT UNSIGNED'), ' NOT NULL,',
    'deviceID VARCHAR(255) NOT NULL,',
    'quantity INT UNSIGNED NOT NULL DEFAULT 1,',
    'custom_price DECIMAL(12,2) NULL COMMENT ''Override price for this device'',',
    'is_required BOOLEAN NOT NULL DEFAULT FALSE COMMENT ''Required vs optional device'',',
    'notes TEXT NULL COMMENT ''Special notes about this device'',',
    'sort_order INT UNSIGNED NULL,',
    'created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,',
    'updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,',
    'PRIMARY KEY (packageID, deviceID),',
    'INDEX idx_package_devices_package (packageID),',
    'INDEX idx_package_devices_device (deviceID),',
    'INDEX idx_package_devices_required (is_required)',
    ') ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci'
);

PREPARE stmt FROM @package_devices_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add foreign key constraints (after tables exist)
-- Package foreign key
SET @fk_package_exists = (
    SELECT COUNT(*) 
    FROM information_schema.TABLE_CONSTRAINTS 
    WHERE CONSTRAINT_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'package_devices' 
    AND CONSTRAINT_NAME = 'fk_package_devices_package'
);

SET @add_package_fk = IF(@fk_package_exists = 0,
    'ALTER TABLE package_devices ADD CONSTRAINT fk_package_devices_package 
     FOREIGN KEY (packageID) REFERENCES equipment_packages(packageID) 
     ON DELETE CASCADE ON UPDATE CASCADE',
    'SELECT "Package foreign key already exists" as message'
);

PREPARE stmt FROM @add_package_fk;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Device foreign key
SET @fk_device_exists = (
    SELECT COUNT(*) 
    FROM information_schema.TABLE_CONSTRAINTS 
    WHERE CONSTRAINT_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'package_devices' 
    AND CONSTRAINT_NAME = 'fk_package_devices_device'
);

SET @add_device_fk = IF(@fk_device_exists = 0,
    'ALTER TABLE package_devices ADD CONSTRAINT fk_package_devices_device 
     FOREIGN KEY (deviceID) REFERENCES devices(deviceID) 
     ON DELETE CASCADE ON UPDATE CASCADE',
    'SELECT "Device foreign key already exists" as message'
);

PREPARE stmt FROM @add_device_fk;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================================================
-- Package Categories Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS package_categories (
    categoryID INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    color VARCHAR(7) NULL COMMENT 'Hex color code',
    sort_order INT UNSIGNED NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_package_categories_name (name),
    INDEX idx_package_categories_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default categories
INSERT IGNORE INTO package_categories (name, description, color, sort_order) VALUES
('Basic', 'Basic equipment packages', '#007bff', 1),
('Premium', 'Premium equipment packages', '#28a745', 2),
('Specialized', 'Specialized packages', '#ffc107', 3),
('Custom', 'Custom packages', '#6c757d', 4);

-- ============================================================================
-- Verification
-- ============================================================================

-- Show final table structures
SELECT 'EQUIPMENT_PACKAGES TABLE:' as info;
DESCRIBE equipment_packages;

SELECT 'PACKAGE_DEVICES TABLE:' as info;
DESCRIBE package_devices;

-- Show foreign key constraints
SELECT 
    CONSTRAINT_NAME,
    TABLE_NAME,
    COLUMN_NAME,
    REFERENCED_TABLE_NAME,
    REFERENCED_COLUMN_NAME
FROM information_schema.KEY_COLUMN_USAGE 
WHERE CONSTRAINT_SCHEMA = DATABASE() 
AND TABLE_NAME = 'package_devices'
AND REFERENCED_TABLE_NAME IS NOT NULL;