-- Add reset_time column to line_accounts table
-- If account has reset_time set, use account's reset_time; otherwise use group's reset_time

ALTER TABLE line_accounts 
ADD COLUMN IF NOT EXISTS reset_time TIME;

COMMENT ON COLUMN line_accounts.reset_time IS 'Account reset time, use group reset_time if NULL';

