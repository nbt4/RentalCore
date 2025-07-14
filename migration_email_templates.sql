-- Migration script for email_templates table
-- Run this manually on your MySQL database

USE `TS-Lager`;

-- Create email_templates table
CREATE TABLE IF NOT EXISTS `email_templates` (
    `template_id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(255) NOT NULL,
    `description` TEXT NULL,
    `template_type` ENUM('invoice', 'reminder', 'payment_confirmation', 'general') NOT NULL DEFAULT 'general',
    `subject` VARCHAR(500) NOT NULL,
    `html_content` LONGTEXT NOT NULL,
    `text_content` LONGTEXT NULL,
    `is_default` BOOLEAN NOT NULL DEFAULT FALSE,
    `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
    `created_by` INT UNSIGNED NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Indexes for better performance
    INDEX `idx_email_templates_type` (`template_type`),
    INDEX `idx_email_templates_default` (`is_default`),
    INDEX `idx_email_templates_active` (`is_active`),
    INDEX `idx_email_templates_created_by` (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Optional: Add foreign key constraint if you have REFERENCES permission
-- If this fails, just skip it - the table will work without the constraint
-- ALTER TABLE `email_templates` 
-- ADD CONSTRAINT `fk_email_templates_created_by` 
-- FOREIGN KEY (`created_by`) REFERENCES `users`(`userID`) ON DELETE SET NULL;

-- Insert some sample email templates (optional)
INSERT INTO `email_templates` (`name`, `description`, `template_type`, `subject`, `html_content`, `text_content`, `is_default`, `is_active`) VALUES
('Default Invoice Template', 'Standard invoice email template', 'invoice', 'Invoice {{.invoice.InvoiceNumber}} - {{.company.CompanyName}}', 
'<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Invoice</title>
</head>
<body>
    <h1>Invoice {{.invoice.InvoiceNumber}}</h1>
    <p>Dear {{.customer.GetDisplayName}},</p>
    <p>Please find attached your invoice for the amount of {{.settings.CurrencySymbol}}{{.invoice.TotalAmount}}.</p>
    <p>Due date: {{.invoice.DueDate.Format "02.01.2006"}}</p>
    <p>Best regards,<br>{{.company.CompanyName}}</p>
</body>
</html>', 
'Invoice {{.invoice.InvoiceNumber}}

Dear {{.customer.GetDisplayName}},

Please find attached your invoice for the amount of {{.settings.CurrencySymbol}}{{.invoice.TotalAmount}}.

Due date: {{.invoice.DueDate.Format "02.01.2006"}}

Best regards,
{{.company.CompanyName}}', 
TRUE, TRUE),

('Default Payment Reminder', 'Standard payment reminder template', 'reminder', 'Payment Reminder - Invoice {{.invoice.InvoiceNumber}}', 
'<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Payment Reminder</title>
</head>
<body>
    <h1>Payment Reminder</h1>
    <p>Dear {{.customer.GetDisplayName}},</p>
    <p>This is a friendly reminder that invoice {{.invoice.InvoiceNumber}} with an outstanding balance of {{.settings.CurrencySymbol}}{{.invoice.BalanceDue}} is now overdue.</p>
    <p>Original due date: {{.invoice.DueDate.Format "02.01.2006"}}</p>
    <p>Please process payment at your earliest convenience.</p>
    <p>Best regards,<br>{{.company.CompanyName}}</p>
</body>
</html>', 
'Payment Reminder - Invoice {{.invoice.InvoiceNumber}}

Dear {{.customer.GetDisplayName}},

This is a friendly reminder that invoice {{.invoice.InvoiceNumber}} with an outstanding balance of {{.settings.CurrencySymbol}}{{.invoice.BalanceDue}} is now overdue.

Original due date: {{.invoice.DueDate.Format "02.01.2006"}}

Please process payment at your earliest convenience.

Best regards,
{{.company.CompanyName}}', 
TRUE, TRUE);

-- Verify the table was created successfully
DESCRIBE `email_templates`;

-- Show sample data
SELECT `template_id`, `name`, `template_type`, `is_default`, `is_active` FROM `email_templates`;