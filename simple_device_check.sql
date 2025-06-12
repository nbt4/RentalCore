-- Simple check to see what devices exist
SELECT COUNT(*) as total_devices FROM devices;

-- Show first 5 devices with their status
SELECT deviceID, status, productID FROM devices LIMIT 5;

-- Show all unique status values
SELECT DISTINCT status, COUNT(*) as count FROM devices GROUP BY status;