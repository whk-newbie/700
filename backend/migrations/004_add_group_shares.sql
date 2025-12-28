-- 004_add_group_shares.sql
-- 创建分组分享表

-- 创建 group_shares 表
CREATE TABLE IF NOT EXISTS group_shares (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    share_code VARCHAR(16) NOT NULL,
    password VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    view_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT unique_share_code UNIQUE (share_code)
);

-- 创建索引
CREATE INDEX idx_group_shares_group_id ON group_shares(group_id);
CREATE INDEX idx_group_shares_share_code ON group_shares(share_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_group_shares_deleted_at ON group_shares(deleted_at);

-- 添加注释
COMMENT ON TABLE group_shares IS '分组分享表';
COMMENT ON COLUMN group_shares.id IS '主键ID';
COMMENT ON COLUMN group_shares.group_id IS '分组ID';
COMMENT ON COLUMN group_shares.share_code IS '分享码';
COMMENT ON COLUMN group_shares.password IS '访问密码（默认为分享码）';
COMMENT ON COLUMN group_shares.expires_at IS '过期时间，为空表示永久有效';
COMMENT ON COLUMN group_shares.is_active IS '是否激活';
COMMENT ON COLUMN group_shares.view_count IS '访问次数统计';
COMMENT ON COLUMN group_shares.created_at IS '创建时间';
COMMENT ON COLUMN group_shares.updated_at IS '更新时间';
COMMENT ON COLUMN group_shares.deleted_at IS '软删除时间';

