# 📊 PHASE 1: DATABASE SCHEMA EXTENSIONS
## Production-Ready Database Design

### 🎯 **OVERVIEW**
Extend existing database to support all new features while maintaining backward compatibility.

---

## 📋 **NEW TABLES TO CREATE**

### 1. **Analytics & Tracking**
```sql
-- Equipment utilization tracking
CREATE TABLE equipment_usage_logs (
    logID int AUTO_INCREMENT PRIMARY KEY,
    deviceID varchar(50) NOT NULL,
    jobID int,
    action enum('assigned', 'returned', 'maintenance', 'available') NOT NULL,
    timestamp datetime NOT NULL,
    duration_hours decimal(10,2),
    revenue_generated decimal(12,2),
    notes text,
    FOREIGN KEY (deviceID) REFERENCES devices(deviceID),
    FOREIGN KEY (jobID) REFERENCES jobs(jobID)
);

-- Financial transactions
CREATE TABLE financial_transactions (
    transactionID int AUTO_INCREMENT PRIMARY KEY,
    jobID int,
    customerID int,
    type enum('rental', 'deposit', 'payment', 'refund', 'fee') NOT NULL,
    amount decimal(12,2) NOT NULL,
    currency varchar(3) DEFAULT 'EUR',
    status enum('pending', 'completed', 'failed', 'cancelled') NOT NULL,
    payment_method varchar(50),
    transaction_date datetime NOT NULL,
    due_date date,
    notes text,
    FOREIGN KEY (jobID) REFERENCES jobs(jobID),
    FOREIGN KEY (customerID) REFERENCES customers(customerID)
);

-- Analytics cache for performance
CREATE TABLE analytics_cache (
    cacheID int AUTO_INCREMENT PRIMARY KEY,
    metric_name varchar(100) NOT NULL,
    period_type enum('daily', 'weekly', 'monthly', 'yearly') NOT NULL,
    period_date date NOT NULL,
    value decimal(15,4),
    metadata json,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_metric (metric_name, period_type, period_date)
);
```

### 2. **Document Management**
```sql
-- Document storage
CREATE TABLE documents (
    documentID int AUTO_INCREMENT PRIMARY KEY,
    entity_type enum('job', 'device', 'customer', 'user') NOT NULL,
    entity_id varchar(50) NOT NULL,
    filename varchar(255) NOT NULL,
    original_filename varchar(255) NOT NULL,
    file_path varchar(500) NOT NULL,
    file_size bigint NOT NULL,
    mime_type varchar(100) NOT NULL,
    document_type enum('contract', 'manual', 'photo', 'invoice', 'receipt', 'other') NOT NULL,
    description text,
    uploaded_by bigint UNSIGNED,
    uploaded_at timestamp DEFAULT CURRENT_TIMESTAMP,
    is_public boolean DEFAULT false,
    version int DEFAULT 1,
    FOREIGN KEY (uploaded_by) REFERENCES users(userID)
);

-- Digital signatures
CREATE TABLE digital_signatures (
    signatureID int AUTO_INCREMENT PRIMARY KEY,
    documentID int NOT NULL,
    signer_name varchar(100) NOT NULL,
    signer_email varchar(100),
    signature_data longtext NOT NULL, -- Base64 encoded signature
    signed_at timestamp DEFAULT CURRENT_TIMESTAMP,
    ip_address varchar(45),
    FOREIGN KEY (documentID) REFERENCES documents(documentID)
);
```

### 3. **Advanced Search & Saved Filters**
```sql
-- Saved searches
CREATE TABLE saved_searches (
    searchID int AUTO_INCREMENT PRIMARY KEY,
    userID bigint UNSIGNED NOT NULL,
    name varchar(100) NOT NULL,
    search_type enum('global', 'jobs', 'devices', 'customers', 'cases') NOT NULL,
    filters json NOT NULL,
    is_default boolean DEFAULT false,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (userID) REFERENCES users(userID)
);

-- Search history for analytics
CREATE TABLE search_history (
    historyID int AUTO_INCREMENT PRIMARY KEY,
    userID bigint UNSIGNED,
    search_term varchar(500),
    search_type varchar(50),
    filters json,
    results_count int,
    searched_at timestamp DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (userID) REFERENCES users(userID)
);
```

### 4. **Workflow & Templates**
```sql
-- Job templates
CREATE TABLE job_templates (
    templateID int AUTO_INCREMENT PRIMARY KEY,
    name varchar(100) NOT NULL,
    description text,
    jobcategoryID int,
    default_duration_days int,
    equipment_list json, -- Array of productIDs or deviceIDs
    notes_template text,
    created_by bigint UNSIGNED,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    is_active boolean DEFAULT true,
    FOREIGN KEY (jobcategoryID) REFERENCES jobCategory(jobcategoryID),
    FOREIGN KEY (created_by) REFERENCES users(userID)
);

-- Equipment packages
CREATE TABLE equipment_packages (
    packageID int AUTO_INCREMENT PRIMARY KEY,
    name varchar(100) NOT NULL,
    description text,
    package_items json, -- Array of {productID, quantity, notes}
    package_price decimal(12,2),
    discount_percent decimal(5,2) DEFAULT 0.00,
    is_active boolean DEFAULT true,
    created_by bigint UNSIGNED,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(userID)
);
```

### 5. **Security & Permissions**
```sql
-- Roles and permissions
CREATE TABLE roles (
    roleID int AUTO_INCREMENT PRIMARY KEY,
    name varchar(50) NOT NULL UNIQUE,
    description text,
    permissions json NOT NULL, -- Array of permission strings
    is_system_role boolean DEFAULT false,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

-- User roles (many-to-many)
CREATE TABLE user_roles (
    userID bigint UNSIGNED NOT NULL,
    roleID int NOT NULL,
    assigned_at timestamp DEFAULT CURRENT_TIMESTAMP,
    assigned_by bigint UNSIGNED,
    PRIMARY KEY (userID, roleID),
    FOREIGN KEY (userID) REFERENCES users(userID),
    FOREIGN KEY (roleID) REFERENCES roles(roleID),
    FOREIGN KEY (assigned_by) REFERENCES users(userID)
);

-- Audit log
CREATE TABLE audit_log (
    auditID bigint AUTO_INCREMENT PRIMARY KEY,
    userID bigint UNSIGNED,
    action varchar(100) NOT NULL,
    entity_type varchar(50) NOT NULL,
    entity_id varchar(50) NOT NULL,
    old_values json,
    new_values json,
    ip_address varchar(45),
    user_agent text,
    timestamp timestamp DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (userID) REFERENCES users(userID),
    INDEX idx_entity (entity_type, entity_id),
    INDEX idx_user_time (userID, timestamp),
    INDEX idx_action_time (action, timestamp)
);
```

### 6. **Mobile & PWA Features**
```sql
-- Push notification subscriptions
CREATE TABLE push_subscriptions (
    subscriptionID int AUTO_INCREMENT PRIMARY KEY,
    userID bigint UNSIGNED NOT NULL,
    endpoint text NOT NULL,
    keys_p256dh text NOT NULL,
    keys_auth text NOT NULL,
    user_agent text,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    last_used timestamp DEFAULT CURRENT_TIMESTAMP,
    is_active boolean DEFAULT true,
    FOREIGN KEY (userID) REFERENCES users(userID)
);

-- Offline sync queue
CREATE TABLE offline_sync_queue (
    queueID int AUTO_INCREMENT PRIMARY KEY,
    userID bigint UNSIGNED NOT NULL,
    action enum('create', 'update', 'delete') NOT NULL,
    entity_type varchar(50) NOT NULL,
    entity_data json NOT NULL,
    timestamp timestamp DEFAULT CURRENT_TIMESTAMP,
    synced boolean DEFAULT false,
    synced_at timestamp NULL,
    FOREIGN KEY (userID) REFERENCES users(userID)
);
```

---

## 📈 **EXISTING TABLE EXTENSIONS**

### Extend `users` table:
```sql
ALTER TABLE users 
ADD COLUMN timezone varchar(50) DEFAULT 'Europe/Berlin',
ADD COLUMN language varchar(5) DEFAULT 'en',
ADD COLUMN avatar_path varchar(500),
ADD COLUMN notification_preferences json,
ADD COLUMN last_active timestamp NULL,
ADD COLUMN login_attempts int DEFAULT 0,
ADD COLUMN locked_until timestamp NULL;
```

### Extend `jobs` table:
```sql
ALTER TABLE jobs
ADD COLUMN templateID int NULL,
ADD COLUMN priority enum('low', 'normal', 'high', 'urgent') DEFAULT 'normal',
ADD COLUMN internal_notes text,
ADD COLUMN customer_notes text,
ADD COLUMN estimated_revenue decimal(12,2),
ADD COLUMN actual_cost decimal(12,2) DEFAULT 0.00,
ADD COLUMN profit_margin decimal(5,2),
ADD COLUMN contract_signed boolean DEFAULT false,
ADD COLUMN contract_documentID int NULL,
ADD FOREIGN KEY (templateID) REFERENCES job_templates(templateID),
ADD FOREIGN KEY (contract_documentID) REFERENCES documents(documentID);
```

### Extend `devices` table:
```sql
ALTER TABLE devices
ADD COLUMN qr_code varchar(255) UNIQUE,
ADD COLUMN current_location varchar(100),
ADD COLUMN gps_coordinates point,
ADD COLUMN condition_rating decimal(3,1) DEFAULT 5.0,
ADD COLUMN usage_hours decimal(10,2) DEFAULT 0.00,
ADD COLUMN total_revenue decimal(12,2) DEFAULT 0.00,
ADD COLUMN last_maintenance_cost decimal(10,2),
ADD COLUMN notes text;
```

### Extend `customers` table:
```sql
ALTER TABLE customers
ADD COLUMN tax_number varchar(50),
ADD COLUMN credit_limit decimal(12,2) DEFAULT 0.00,
ADD COLUMN payment_terms int DEFAULT 30,
ADD COLUMN preferred_payment_method varchar(50),
ADD COLUMN customer_since date,
ADD COLUMN total_lifetime_value decimal(12,2) DEFAULT 0.00,
ADD COLUMN last_job_date date,
ADD COLUMN rating decimal(3,1) DEFAULT 5.0;
```

---

## 🔧 **INDEXES FOR PERFORMANCE**
```sql
-- Analytics performance
CREATE INDEX idx_usage_logs_device_date ON equipment_usage_logs(deviceID, timestamp);
CREATE INDEX idx_transactions_customer_date ON financial_transactions(customerID, transaction_date);
CREATE INDEX idx_transactions_status ON financial_transactions(status, due_date);

-- Search performance  
CREATE FULLTEXT INDEX idx_customers_search ON customers(companyname, firstname, lastname, email);
CREATE FULLTEXT INDEX idx_jobs_search ON jobs(description);

-- Document management
CREATE INDEX idx_documents_entity ON documents(entity_type, entity_id, document_type);
CREATE INDEX idx_documents_date ON documents(uploaded_at, document_type);
```

---

## 🚀 **MIGRATION SCRIPT**
Ready to create the migration SQL file that will safely update your production database while preserving all existing data.

**Next Step: Create the complete migration script?** ✅