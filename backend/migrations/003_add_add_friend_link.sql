-- 添加 add_friend_link 字段到 line_accounts 表
-- Version: v1.1
-- Created: 2025-01-XX

ALTER TABLE line_accounts 
ADD COLUMN IF NOT EXISTS add_friend_link VARCHAR(500);

COMMENT ON COLUMN line_accounts.add_friend_link IS 'Line添加好友链接';

