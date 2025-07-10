-- Performance optimization for devices table
-- Create indexes for commonly queried fields

-- Index for device ID lookups (primary key should already be indexed)
CREATE INDEX IF NOT EXISTS idx_devices_device_id ON devices(deviceID);

-- Index for product ID joins
CREATE INDEX IF NOT EXISTS idx_devices_product_id ON devices(productID);

-- Index for serial number searches
CREATE INDEX IF NOT EXISTS idx_devices_serial_number ON devices(serialnumber);

-- Index for status filtering
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status);

-- Composite index for common queries (device ID + product ID)
CREATE INDEX IF NOT EXISTS idx_devices_device_product ON devices(deviceID, productID);

-- Index for products table name searches
CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);

-- Index for product categories
CREATE INDEX IF NOT EXISTS idx_products_category ON products(categoryID);

-- Index for job devices assignments
CREATE INDEX IF NOT EXISTS idx_jobdevices_device_id ON jobdevices(deviceID);

-- Composite index for job devices (device + job)
CREATE INDEX IF NOT EXISTS idx_jobdevices_device_job ON jobdevices(deviceID, jobID);

-- Analyze tables to update statistics
ANALYZE TABLE devices;
ANALYZE TABLE products;
ANALYZE TABLE jobdevices;

-- Show current table sizes
SELECT 
    'devices' as table_name,
    COUNT(*) as row_count,
    ROUND(((data_length + index_length) / 1024 / 1024), 2) as size_mb
FROM information_schema.tables t
JOIN devices d ON 1=1
WHERE t.table_schema = DATABASE() AND t.table_name = 'devices'
GROUP BY t.table_name

UNION ALL

SELECT 
    'products' as table_name,
    COUNT(*) as row_count,
    ROUND(((data_length + index_length) / 1024 / 1024), 2) as size_mb
FROM information_schema.tables t
JOIN products p ON 1=1
WHERE t.table_schema = DATABASE() AND t.table_name = 'products'
GROUP BY t.table_name

UNION ALL

SELECT 
    'jobdevices' as table_name,
    COUNT(*) as row_count,
    ROUND(((data_length + index_length) / 1024 / 1024), 2) as size_mb
FROM information_schema.tables t
JOIN jobdevices jd ON 1=1
WHERE t.table_schema = DATABASE() AND t.table_name = 'jobdevices'
GROUP BY t.table_name;