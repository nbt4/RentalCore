-- Enhanced Equipment Packages Migration
-- Add new fields for production-ready equipment packages

-- Add new columns to equipment_packages table (with error handling)
SET @sql = '';

-- Check and add max_rental_days column
SELECT COUNT(*) INTO @count FROM information_schema.columns 
WHERE table_schema = DATABASE() AND table_name = 'equipment_packages' AND column_name = 'max_rental_days';
SET @sql = IF(@count = 0, 'ALTER TABLE equipment_packages ADD COLUMN max_rental_days INT NULL;', '');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Check and add category column
SELECT COUNT(*) INTO @count FROM information_schema.columns 
WHERE table_schema = DATABASE() AND table_name = 'equipment_packages' AND column_name = 'category';
SET @sql = IF(@count = 0, 'ALTER TABLE equipment_packages ADD COLUMN category VARCHAR(50) NULL;', '');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Check and add tags column
SELECT COUNT(*) INTO @count FROM information_schema.columns 
WHERE table_schema = DATABASE() AND table_name = 'equipment_packages' AND column_name = 'tags';
SET @sql = IF(@count = 0, 'ALTER TABLE equipment_packages ADD COLUMN tags TEXT NULL;', '');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Check and add last_used_at column
SELECT COUNT(*) INTO @count FROM information_schema.columns 
WHERE table_schema = DATABASE() AND table_name = 'equipment_packages' AND column_name = 'last_used_at';
SET @sql = IF(@count = 0, 'ALTER TABLE equipment_packages ADD COLUMN last_used_at TIMESTAMP NULL;', '');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Check and add total_revenue column
SELECT COUNT(*) INTO @count FROM information_schema.columns 
WHERE table_schema = DATABASE() AND table_name = 'equipment_packages' AND column_name = 'total_revenue';
SET @sql = IF(@count = 0, 'ALTER TABLE equipment_packages ADD COLUMN total_revenue DECIMAL(12,2) DEFAULT 0.00;', '');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Add indexes for better performance (with error handling)
-- Equipment packages indexes
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM information_schema.statistics 
     WHERE table_schema = DATABASE() AND table_name = 'equipment_packages' AND index_name = 'idx_equipment_packages_category') = 0,
    'CREATE INDEX idx_equipment_packages_category ON equipment_packages(category);',
    'SELECT "Index idx_equipment_packages_category already exists";'
));
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM information_schema.statistics 
     WHERE table_schema = DATABASE() AND table_name = 'equipment_packages' AND index_name = 'idx_equipment_packages_active') = 0,
    'CREATE INDEX idx_equipment_packages_active ON equipment_packages(is_active);',
    'SELECT "Index idx_equipment_packages_active already exists";'
));
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM information_schema.statistics 
     WHERE table_schema = DATABASE() AND table_name = 'equipment_packages' AND index_name = 'idx_equipment_packages_usage') = 0,
    'CREATE INDEX idx_equipment_packages_usage ON equipment_packages(usage_count);',
    'SELECT "Index idx_equipment_packages_usage already exists";'
));
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Package devices indexes
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM information_schema.statistics 
     WHERE table_schema = DATABASE() AND table_name = 'package_devices' AND index_name = 'idx_package_devices_package_id') = 0,
    'CREATE INDEX idx_package_devices_package_id ON package_devices(package_id);',
    'SELECT "Index idx_package_devices_package_id already exists";'
));
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM information_schema.statistics 
     WHERE table_schema = DATABASE() AND table_name = 'package_devices' AND index_name = 'idx_package_devices_device_id') = 0,
    'CREATE INDEX idx_package_devices_device_id ON package_devices(device_id);',
    'SELECT "Index idx_package_devices_device_id already exists";'
));
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- Update existing package_items to be valid JSON if NULL
UPDATE equipment_packages 
SET package_items = '[]' 
WHERE package_items IS NULL OR package_items = '';

-- Add check constraints for data validation
ALTER TABLE equipment_packages 
ADD CONSTRAINT IF NOT EXISTS chk_package_price_positive 
    CHECK (package_price IS NULL OR package_price >= 0);

ALTER TABLE equipment_packages 
ADD CONSTRAINT IF NOT EXISTS chk_discount_range 
    CHECK (discount_percent >= 0 AND discount_percent <= 100);

ALTER TABLE equipment_packages 
ADD CONSTRAINT IF NOT EXISTS chk_rental_days_positive 
    CHECK (min_rental_days > 0);

ALTER TABLE equipment_packages 
ADD CONSTRAINT IF NOT EXISTS chk_max_rental_days 
    CHECK (max_rental_days IS NULL OR max_rental_days >= min_rental_days);

ALTER TABLE package_devices 
ADD CONSTRAINT IF NOT EXISTS chk_quantity_positive 
    CHECK (quantity > 0);

ALTER TABLE package_devices 
ADD CONSTRAINT IF NOT EXISTS chk_custom_price_positive 
    CHECK (custom_price IS NULL OR custom_price >= 0);

-- Create view for package statistics
CREATE OR REPLACE VIEW package_stats AS
SELECT 
    ep.package_id,
    ep.name,
    ep.category,
    ep.is_active,
    ep.usage_count,
    ep.total_revenue,
    ep.last_used_at,
    COUNT(pd.device_id) as device_count,
    SUM(pd.quantity) as total_quantity,
    SUM(CASE WHEN pd.is_required THEN 1 ELSE 0 END) as required_devices,
    COALESCE(
        CASE 
            WHEN ep.package_price IS NOT NULL THEN ep.package_price
            ELSE (
                SELECT SUM(
                    COALESCE(pd2.custom_price, p.item_cost_per_day, 0) * pd2.quantity
                )
                FROM package_devices pd2
                LEFT JOIN devices d ON pd2.device_id = d.device_id
                LEFT JOIN products p ON d.product_id = p.product_id
                WHERE pd2.package_id = ep.package_id
            )
        END, 0
    ) as calculated_price,
    CASE 
        WHEN ep.package_price IS NOT NULL THEN ep.package_price * (1 - ep.discount_percent / 100)
        ELSE (
            SELECT SUM(
                COALESCE(pd2.custom_price, p.item_cost_per_day, 0) * pd2.quantity
            ) * (1 - ep.discount_percent / 100)
            FROM package_devices pd2
            LEFT JOIN devices d ON pd2.device_id = d.device_id
            LEFT JOIN products p ON d.product_id = p.product_id
            WHERE pd2.package_id = ep.package_id
        )
    END as final_price
FROM equipment_packages ep
LEFT JOIN package_devices pd ON ep.package_id = pd.package_id
GROUP BY ep.package_id;

-- Create function to update package revenue
DELIMITER $$
CREATE OR REPLACE FUNCTION update_package_revenue(
    p_package_id INT,
    p_revenue DECIMAL(12,2)
) RETURNS BOOLEAN
READS SQL DATA
DETERMINISTIC
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RETURN FALSE;
    END;
    
    START TRANSACTION;
    
    UPDATE equipment_packages 
    SET 
        total_revenue = total_revenue + p_revenue,
        updated_at = NOW()
    WHERE package_id = p_package_id;
    
    COMMIT;
    RETURN TRUE;
END$$
DELIMITER ;

-- Create function to increment package usage
DELIMITER $$
CREATE OR REPLACE FUNCTION increment_package_usage(
    p_package_id INT
) RETURNS BOOLEAN
READS SQL DATA
DETERMINISTIC
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        RETURN FALSE;
    END;
    
    START TRANSACTION;
    
    UPDATE equipment_packages 
    SET 
        usage_count = usage_count + 1,
        last_used_at = NOW(),
        updated_at = NOW()
    WHERE package_id = p_package_id;
    
    COMMIT;
    RETURN TRUE;
END$$
DELIMITER ;

-- Add sample categories if packages table is empty
INSERT IGNORE INTO equipment_packages (name, description, category, is_active, min_rental_days, created_at, updated_at)
VALUES 
('Audio Basic Package', 'Basic audio equipment for small events', 'audio', true, 1, NOW(), NOW()),
('Lighting Standard Package', 'Standard lighting setup for medium venues', 'lighting', true, 2, NOW(), NOW()),
('Video Production Package', 'Complete video production equipment', 'video', true, 1, NOW(), NOW());

-- Update package statistics for existing packages
UPDATE equipment_packages ep 
SET total_revenue = COALESCE(
    (SELECT SUM(ft.amount) 
     FROM financial_transactions ft 
     INNER JOIN jobs j ON ft.job_id = j.job_id 
     WHERE j.description LIKE CONCAT('%', ep.name, '%') 
     AND ft.type = 'rental' 
     AND ft.status = 'completed'), 
    0
)
WHERE ep.package_id > 0;