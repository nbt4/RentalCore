-- ================================================================
-- Migration: Remove Job Templates
-- Description: Remove job template functionality completely
-- ================================================================

-- Remove foreign key constraints that reference job_templates table
ALTER TABLE jobs DROP FOREIGN KEY jobs_ibfk_4;

-- Remove templateID column from jobs table
ALTER TABLE jobs DROP COLUMN templateID;

-- Drop job_templates table completely
DROP TABLE IF EXISTS job_templates;

-- ================================================================
-- Clean up any remaining references
-- ================================================================

-- No other tables should reference job_templates at this point
-- The migration is complete