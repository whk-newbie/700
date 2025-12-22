-- Add deleted_at column to contact_pool table
-- Version: v1.2
-- Created: 2025-12-22

ALTER TABLE contact_pool 
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_contact_pool_deleted ON contact_pool(deleted_at);

COMMENT ON COLUMN contact_pool.deleted_at IS 'Soft delete timestamp';

