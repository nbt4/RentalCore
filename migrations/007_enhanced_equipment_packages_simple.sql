-- Enhanced Equipment Packages Migration (MySQL Compatible)
-- Add new fields for production-ready equipment packages

-- Add new columns to equipment_packages table
-- Note: Run each ALTER TABLE separately if you get syntax errors

-- Add max_rental_days column
ALTER TABLE equipment_packages ADD COLUMN max_rental_days INT NULL;

-- Add category column  
ALTER TABLE equipment_packages ADD COLUMN category VARCHAR(50) NULL;

-- Add tags column
ALTER TABLE equipment_packages ADD COLUMN tags TEXT NULL;

-- Add last_used_at column
ALTER TABLE equipment_packages ADD COLUMN last_used_at TIMESTAMP NULL;

-- Add total_revenue column
ALTER TABLE equipment_packages ADD COLUMN total_revenue DECIMAL(12,2) DEFAULT 0.00;

-- Add indexes for better performance
CREATE INDEX idx_equipment_packages_category ON equipment_packages(category);
CREATE INDEX idx_equipment_packages_active ON equipment_packages(is_active);
CREATE INDEX idx_equipment_packages_usage ON equipment_packages(usage_count);

-- Add indexes for package_devices table
CREATE INDEX idx_package_devices_package_id ON package_devices(package_id);
CREATE INDEX idx_package_devices_device_id ON package_devices(device_id);

-- Update existing package_items to be valid JSON if NULL
UPDATE equipment_packages 
SET package_items = '[]' 
WHERE package_items IS NULL OR package_items = '';

-- Add sample categories if packages table is empty
INSERT IGNORE INTO equipment_packages (name, description, category, is_active, min_rental_days, created_at, updated_at, package_items)
VALUES 
('Audio Basic Package', 'Basic audio equipment for small events', 'audio', true, 1, NOW(), NOW(), '[]'),
('Lighting Standard Package', 'Standard lighting setup for medium venues', 'lighting', true, 2, NOW(), NOW(), '[]'),
('Video Production Package', 'Complete video production equipment', 'video', true, 1, NOW(), NOW(), '[]');