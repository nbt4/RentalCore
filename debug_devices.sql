-- Debug query to check device data for package creation
SELECT 
    'DEVICE STATUS ANALYSIS:' as info;

-- Check all unique status values
SELECT 
    status, 
    COUNT(*) as count 
FROM devices 
GROUP BY status 
ORDER BY count DESC;

-- Check sample devices with status info
SELECT 
    deviceID,
    status,
    serialnumber,
    productID
FROM devices 
LIMIT 10;

-- Check if any devices have status 'free'
SELECT COUNT(*) as free_devices_count 
FROM devices 
WHERE status = 'free';

-- Check devices that could be available (different status values)
SELECT 
    status, 
    COUNT(*) as count,
    GROUP_CONCAT(deviceID LIMIT 5) as sample_device_ids
FROM devices 
GROUP BY status;