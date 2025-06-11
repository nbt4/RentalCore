# 📊 DATABASE MIGRATION MANUAL
## Step-by-Step Guide for Applying Enhancement Features

### ⚠️ **IMPORTANT: BACKUP FIRST!**
Before running any migration, create a complete database backup:

```bash
mysqldump -h mysql -u root -p TS-Lager > backup_$(date +%Y%m%d_%H%M%S).sql
```

---

## 🗂️ **MIGRATION FILES TO RUN**

### **File Location:** `/opt/dev/go-barcode-webapp/migrations/002_enhancement_features.sql`

This migration will add:
- ✅ **Analytics & Tracking Tables** - Equipment usage logs, financial transactions, analytics cache
- ✅ **Document Management** - File storage, digital signatures  
- ✅ **Advanced Search** - Saved searches, search history
- ✅ **Workflow Features** - Job templates, equipment packages
- ✅ **Security & Permissions** - Roles, user roles, audit logs
- ✅ **Mobile & PWA** - Push subscriptions, offline sync queue
- ✅ **Enhanced Existing Tables** - Add new columns to users, jobs, devices, customers

---

## 🚀 **STEP-BY-STEP EXECUTION**

### **Step 1: Connect to Database**
```bash
mysql -h mysql -u root -p
```

### **Step 2: Select Database**
```sql
USE `TS-Lager`;
```

### **Step 3: Run the Migration**
```sql
SOURCE /opt/dev/go-barcode-webapp/migrations/002_enhancement_features.sql;
```

**OR** if you prefer to run from command line:
```bash
mysql -h mysql -u root -p TS-Lager < /opt/dev/go-barcode-webapp/migrations/002_enhancement_features.sql
```

---

## ✅ **VERIFICATION STEPS**

After running the migration, verify the changes:

### **1. Check New Tables Were Created:**
```sql
SHOW TABLES LIKE '%usage_logs%';
SHOW TABLES LIKE '%financial_transactions%';
SHOW TABLES LIKE '%documents%';
SHOW TABLES LIKE '%roles%';
SHOW TABLES LIKE '%analytics_cache%';
```

### **2. Verify New Columns Added:**
```sql
DESCRIBE users;
DESCRIBE jobs;
DESCRIBE devices;
DESCRIBE customers;
```

### **3. Check Default Data:**
```sql
SELECT * FROM roles;
SELECT * FROM job_templates;
```

### **4. Verify Views Created:**
```sql
SHOW FULL TABLES WHERE table_type = 'VIEW';
```

---

## 🔧 **ROLLBACK PLAN** (If Something Goes Wrong)

### **Option 1: Restore from Backup**
```bash
mysql -h mysql -u root -p TS-Lager < backup_YYYYMMDD_HHMMSS.sql
```

### **Option 2: Manual Rollback** (Advanced)
```sql
-- Drop new tables (be very careful!)
DROP TABLE IF EXISTS equipment_usage_logs;
DROP TABLE IF EXISTS financial_transactions;
DROP TABLE IF EXISTS analytics_cache;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS digital_signatures;
DROP TABLE IF EXISTS saved_searches;
DROP TABLE IF EXISTS search_history;
DROP TABLE IF EXISTS job_templates;
DROP TABLE IF EXISTS equipment_packages;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS push_subscriptions;
DROP TABLE IF EXISTS offline_sync_queue;

-- Drop views
DROP VIEW IF EXISTS equipment_utilization;
DROP VIEW IF EXISTS customer_performance;
DROP VIEW IF EXISTS monthly_revenue;

-- Remove added columns (be extremely careful!)
ALTER TABLE users 
DROP COLUMN timezone,
DROP COLUMN language,
DROP COLUMN avatar_path,
DROP COLUMN notification_preferences,
DROP COLUMN last_active,
DROP COLUMN login_attempts,
DROP COLUMN locked_until,
DROP COLUMN two_factor_enabled,
DROP COLUMN two_factor_secret;
```

---

## 📝 **POST-MIGRATION TASKS**

### **1. Update Application Code**
The migration is already compatible with the new handlers I created:
- ✅ `analytics_handler.go` 
- ✅ `search_handler.go`
- ✅ Updated `main.go` with new routes

### **2. Restart Application**
```bash
cd /opt/dev/go-barcode-webapp
go build -o server cmd/server/main.go
./server -config=config.json
```

### **3. Test New Features**
- **Analytics Dashboard:** `http://localhost:9000/analytics`
- **Advanced Search:** `http://localhost:9000/search/global?q=test`
- **User Management:** Should work as before

### **4. Create First Admin Role Assignment**
```sql
-- Assign super_admin role to your main user (replace 1 with your user ID)
INSERT INTO user_roles (userID, roleID, assigned_by) 
VALUES (1, 1, 1);
```

---

## 📊 **MIGRATION IMPACT**

### **Database Size Impact:**
- **~15 new tables** added
- **~20 new columns** added to existing tables
- **3 new views** for analytics
- **~10 new indexes** for performance

### **Performance Impact:**
- **Minimal** - All new tables start empty
- **Improved** search performance with new indexes
- **Analytics** queries are optimized with views

### **Data Safety:**
- **No existing data modified** - only additions
- **Backward compatible** - existing app functionality unchanged
- **Transactional migration** - rolls back completely if any step fails

---

## 🆘 **TROUBLESHOOTING**

### **Error: "Table already exists"**
```sql
-- Check if migration was partially run
SHOW TABLES LIKE '%enhancement%';
-- If yes, you may need manual cleanup before re-running
```

### **Error: "Unknown column"**
```sql
-- Check if columns were added correctly
DESCRIBE users;
DESCRIBE devices;
```

### **Error: "Syntax error"**
- Ensure you're running MySQL 5.7+ (for JSON support)
- Check that database collation is utf8mb4

### **Performance Issues After Migration**
```sql
-- Run these optimization commands
ANALYZE TABLE equipment_usage_logs;
ANALYZE TABLE financial_transactions;
OPTIMIZE TABLE jobs;
OPTIMIZE TABLE devices;
```

---

## ✅ **SUCCESS CONFIRMATION**

You'll know the migration was successful when:

1. **✅ Analytics Dashboard loads:** `http://localhost:9000/analytics`
2. **✅ Search works:** Global search returns results
3. **✅ No errors in application logs**
4. **✅ All existing functionality still works**
5. **✅ New tables visible in database**

---

## 📞 **NEED HELP?**

If you encounter any issues:
1. **Check the backup** is complete before proceeding
2. **Review error messages** carefully
3. **Test on development environment** first if possible
4. **Take screenshots** of any errors for troubleshooting

**The migration is designed to be safe and reversible!** 🚀