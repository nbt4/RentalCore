-- ============================================================================
-- Check Existing Table Structure
-- Run this first to see what packageID data type exists
-- ============================================================================

-- Check if equipment_packages table exists and show structure
SELECT 'EQUIPMENT_PACKAGES TABLE STRUCTURE:' as info;
DESCRIBE equipment_packages;

-- Check devices table structure 
SELECT 'DEVICES TABLE STRUCTURE:' as info;  
DESCRIBE devices;

-- Show specific column info
SELECT 
    COLUMN_NAME,
    COLUMN_TYPE,
    IS_NULLABLE,
    COLUMN_KEY,
    COLUMN_DEFAULT
FROM information_schema.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
AND TABLE_NAME = 'equipment_packages'
AND COLUMN_NAME = 'packageID';

-- Show devices deviceID column
SELECT 
    COLUMN_NAME,
    COLUMN_TYPE,
    IS_NULLABLE,
    COLUMN_KEY
FROM information_schema.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
AND TABLE_NAME = 'devices'
AND COLUMN_NAME = 'deviceID';

-- Count records
SELECT 'DATA COUNTS:' as info;
SELECT COUNT(*) as equipment_packages_count FROM equipment_packages;
SELECT COUNT(*) as devices_count FROM devices;