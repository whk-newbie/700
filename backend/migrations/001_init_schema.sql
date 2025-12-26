-- Line Account Management System - Database Initialization Script
-- Version: v2.0 (Merged all migrations)
-- Created: 2025-12-21
-- Updated: 2025-01-XX (Merged migrations 001-005)

-- ============================================
-- 1. Create Base Tables
-- ============================================

-- 1.1 users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100),
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    max_groups INTEGER DEFAULT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT check_role CHECK (role IN ('admin', 'user'))
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_deleted ON users(deleted_at);

-- 1.2 groups table
CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activation_code VARCHAR(32) UNIQUE NOT NULL,
    account_limit INTEGER DEFAULT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    remark VARCHAR(255),
    description TEXT,
    category VARCHAR(50) DEFAULT 'default',
    dedup_scope VARCHAR(20) DEFAULT 'current',
    reset_time TIME DEFAULT '09:00:00',
    login_password VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT check_dedup_scope CHECK (dedup_scope IN ('current', 'global'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_groups_activation_code ON groups(activation_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_groups_user_id ON groups(user_id);
CREATE INDEX IF NOT EXISTS idx_groups_category ON groups(category);
CREATE INDEX IF NOT EXISTS idx_groups_is_active ON groups(is_active);
CREATE INDEX IF NOT EXISTS idx_groups_deleted ON groups(deleted_at);

-- 1.3 group_stats table
CREATE TABLE IF NOT EXISTS group_stats (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    total_accounts INTEGER DEFAULT 0,
    online_accounts INTEGER DEFAULT 0,
    line_accounts INTEGER DEFAULT 0,
    line_business_accounts INTEGER DEFAULT 0,
    today_incoming INTEGER DEFAULT 0,
    total_incoming INTEGER DEFAULT 0,
    duplicate_incoming INTEGER DEFAULT 0,
    today_duplicate INTEGER DEFAULT 0,
    last_reset_date DATE,
    last_reset_time TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_group_stats UNIQUE(group_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_stats_unique ON group_stats(group_id);

-- 1.4 line_accounts table
CREATE TABLE IF NOT EXISTS line_accounts (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    activation_code VARCHAR(32) NOT NULL,
    platform_type VARCHAR(20) NOT NULL DEFAULT 'line',
    line_id VARCHAR(100) NOT NULL,
    display_name VARCHAR(100),
    phone_number VARCHAR(20),
    profile_url VARCHAR(500),
    avatar_url VARCHAR(500),
    bio TEXT,
    status_message VARCHAR(255),
    add_friend_link VARCHAR(500),
    qr_code_path VARCHAR(255),
    online_status VARCHAR(20) DEFAULT 'offline',
    reset_time TIME,
    last_active_at TIMESTAMP,
    last_online_time TIMESTAMP,
    first_login_at TIMESTAMP,
    account_remark TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    deleted_by INTEGER REFERENCES users(id),
    CONSTRAINT check_platform_type CHECK (platform_type IN ('line', 'line_business')),
    CONSTRAINT check_online_status CHECK (online_status IN ('online', 'offline', 'user_logout', 'abnormal_offline'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_line_accounts_unique ON line_accounts(group_id, line_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_line_accounts_group_id ON line_accounts(group_id);
CREATE INDEX IF NOT EXISTS idx_line_accounts_activation_code ON line_accounts(activation_code);
CREATE INDEX IF NOT EXISTS idx_line_accounts_platform_type ON line_accounts(platform_type);
CREATE INDEX IF NOT EXISTS idx_line_accounts_online_status ON line_accounts(online_status);
CREATE INDEX IF NOT EXISTS idx_line_accounts_line_id ON line_accounts(line_id);
CREATE INDEX IF NOT EXISTS idx_line_accounts_deleted ON line_accounts(deleted_at);

-- 1.5 line_account_stats table
CREATE TABLE IF NOT EXISTS line_account_stats (
    id SERIAL PRIMARY KEY,
    line_account_id INTEGER NOT NULL REFERENCES line_accounts(id) ON DELETE CASCADE,
    today_incoming INTEGER DEFAULT 0,
    total_incoming INTEGER DEFAULT 0,
    duplicate_incoming INTEGER DEFAULT 0,
    today_duplicate INTEGER DEFAULT 0,
    last_reset_date DATE,
    last_reset_time TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_line_account_stats UNIQUE(line_account_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_line_account_stats_unique ON line_account_stats(line_account_id);

-- 1.6 import_batches table
CREATE TABLE IF NOT EXISTS import_batches (
    id SERIAL PRIMARY KEY,
    batch_name VARCHAR(100),
    platform_type VARCHAR(20) NOT NULL,
    total_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    duplicate_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    dedup_scope VARCHAR(20),
    file_name VARCHAR(255),
    file_path VARCHAR(500),
    file_size BIGINT,
    imported_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    CONSTRAINT check_platform_type_batch CHECK (platform_type IN ('line', 'line_business')),
    CONSTRAINT check_dedup_scope_batch CHECK (dedup_scope IN ('current', 'global'))
);

CREATE INDEX IF NOT EXISTS idx_import_batches_platform ON import_batches(platform_type);
CREATE INDEX IF NOT EXISTS idx_import_batches_created ON import_batches(created_at DESC);

-- 1.7 contact_pool table
CREATE TABLE IF NOT EXISTS contact_pool (
    id BIGSERIAL PRIMARY KEY,
    source_type VARCHAR(20) NOT NULL,
    import_batch_id INTEGER REFERENCES import_batches(id),
    group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    activation_code VARCHAR(32) NOT NULL,
    line_account_id INTEGER REFERENCES line_accounts(id),
    platform_type VARCHAR(20) NOT NULL,
    line_id VARCHAR(100) NOT NULL,
    display_name VARCHAR(100),
    phone_number VARCHAR(20),
    avatar_url VARCHAR(500),
    dedup_scope VARCHAR(20),
    first_seen_at TIMESTAMP,
    remark TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT check_source_type CHECK (source_type IN ('import', 'platform')),
    CONSTRAINT check_platform_type_cp CHECK (platform_type IN ('line', 'line_business'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_contact_pool_global_unique ON contact_pool(line_id, platform_type);
CREATE INDEX IF NOT EXISTS idx_contact_pool_group_id ON contact_pool(group_id);
CREATE INDEX IF NOT EXISTS idx_contact_pool_activation_code ON contact_pool(activation_code);
CREATE INDEX IF NOT EXISTS idx_contact_pool_line_id ON contact_pool(line_id);
CREATE INDEX IF NOT EXISTS idx_contact_pool_deleted ON contact_pool(deleted_at);

-- 1.8 customers table
CREATE TABLE IF NOT EXISTS customers (
    id BIGSERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    activation_code VARCHAR(32) NOT NULL,
    line_account_id INTEGER REFERENCES line_accounts(id),
    platform_type VARCHAR(20) NOT NULL,
    customer_id VARCHAR(100) NOT NULL,
    display_name VARCHAR(100),
    avatar_url VARCHAR(500),
    phone_number VARCHAR(20),
    customer_type VARCHAR(50),
    gender VARCHAR(10),
    country VARCHAR(50),
    birthday DATE,
    address TEXT,
    nickname_remark VARCHAR(20),
    remark TEXT,
    tags JSONB,
    profile_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT check_platform_type_cust CHECK (platform_type IN ('line', 'line_business')),
    CONSTRAINT check_gender CHECK (gender IN ('male', 'female', 'unknown'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_unique ON customers(customer_id, platform_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_customers_group_id ON customers(group_id);
CREATE INDEX IF NOT EXISTS idx_customers_activation_code ON customers(activation_code);
CREATE INDEX IF NOT EXISTS idx_customers_line_account ON customers(line_account_id);
CREATE INDEX IF NOT EXISTS idx_customers_platform_type ON customers(platform_type);

-- 1.9 follow_up_records table
CREATE TABLE IF NOT EXISTS follow_up_records (
    id BIGSERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    activation_code VARCHAR(32) NOT NULL,
    line_account_id INTEGER REFERENCES line_accounts(id),
    customer_id BIGINT REFERENCES customers(id),
    platform_type VARCHAR(20) NOT NULL,
    line_account_display_name VARCHAR(100),
    line_account_line_id VARCHAR(100),
    line_account_avatar_url VARCHAR(500),
    customer_display_name VARCHAR(100),
    customer_line_id VARCHAR(100),
    customer_avatar_url VARCHAR(500),
    content TEXT NOT NULL,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT check_platform_type_fup CHECK (platform_type IN ('line', 'line_business'))
);

CREATE INDEX IF NOT EXISTS idx_follow_up_group_id ON follow_up_records(group_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_follow_up_line_account ON follow_up_records(line_account_id);
CREATE INDEX IF NOT EXISTS idx_follow_up_customer ON follow_up_records(customer_id);

-- 1.10 llm_configs table
CREATE TABLE IF NOT EXISTS llm_configs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    api_url VARCHAR(500) NOT NULL,
    api_key TEXT NOT NULL,
    model VARCHAR(100) NOT NULL,
    max_tokens INTEGER DEFAULT 2000,
    temperature DECIMAL(3,2) DEFAULT 0.7,
    top_p DECIMAL(3,2) DEFAULT 1.0,
    frequency_penalty DECIMAL(3,2) DEFAULT 0.0,
    presence_penalty DECIMAL(3,2) DEFAULT 0.0,
    system_prompt TEXT,
    timeout_seconds INTEGER DEFAULT 30,
    max_retries INTEGER DEFAULT 3,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id),
    CONSTRAINT check_provider CHECK (provider IN ('openai', 'anthropic', 'aliyun', 'xunfei', 'baidu', 'zhipu', 'custom'))
);

CREATE INDEX IF NOT EXISTS idx_llm_configs_provider ON llm_configs(provider);
CREATE INDEX IF NOT EXISTS idx_llm_configs_is_active ON llm_configs(is_active);

-- 1.11 llm_prompt_templates table
-- 注意：此表已废弃，保留用于历史数据兼容，不再使用外键约束
CREATE TABLE IF NOT EXISTS llm_prompt_templates (
    id SERIAL PRIMARY KEY,
    config_id INTEGER,
    template_name VARCHAR(100) NOT NULL,
    template_content TEXT NOT NULL,
    variables JSONB,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_prompt_templates_config ON llm_prompt_templates(config_id);
CREATE INDEX IF NOT EXISTS idx_prompt_templates_active ON llm_prompt_templates(is_active);

-- 1.12 llm_call_logs table
-- 注意：config_id 和 template_id 不使用外键约束，以支持简化的配置管理
CREATE TABLE IF NOT EXISTS llm_call_logs (
    id BIGSERIAL PRIMARY KEY,
    config_id INTEGER,
    template_id INTEGER,
    group_id INTEGER REFERENCES groups(id),
    activation_code VARCHAR(32),
    request_messages JSONB NOT NULL,
    request_params JSONB,
    response_content TEXT,
    response_data JSONB,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    tokens_used INTEGER,
    prompt_tokens INTEGER,
    completion_tokens INTEGER,
    call_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    duration_ms INTEGER,
    CONSTRAINT check_llm_status CHECK (status IN ('success', 'error'))
);

CREATE INDEX IF NOT EXISTS idx_llm_call_logs_config ON llm_call_logs(config_id, call_time DESC);
CREATE INDEX IF NOT EXISTS idx_llm_call_logs_group ON llm_call_logs(group_id, call_time DESC);
CREATE INDEX IF NOT EXISTS idx_llm_call_logs_time ON llm_call_logs(call_time DESC);

-- ============================================
-- 2. Create Partition Tables
-- ============================================

-- 2.1 incoming_logs partition table
CREATE TABLE IF NOT EXISTS incoming_logs (
    id BIGSERIAL,
    line_account_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    incoming_line_id VARCHAR(100) NOT NULL,
    incoming_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    display_name VARCHAR(100),
    avatar_url VARCHAR(500),
    phone_number VARCHAR(20),
    is_duplicate BOOLEAN DEFAULT FALSE,
    duplicate_scope VARCHAR(20),
    customer_type VARCHAR(50),
    raw_data JSONB,
    PRIMARY KEY (id, incoming_time)
) PARTITION BY RANGE (incoming_time);

-- Create 2025 partitions
CREATE TABLE IF NOT EXISTS incoming_logs_2025_01 PARTITION OF incoming_logs
    FOR VALUES FROM ('2025-01-01 00:00:00') TO ('2025-02-01 00:00:00');
CREATE TABLE IF NOT EXISTS incoming_logs_2025_02 PARTITION OF incoming_logs
    FOR VALUES FROM ('2025-02-01 00:00:00') TO ('2025-03-01 00:00:00');
CREATE TABLE IF NOT EXISTS incoming_logs_2025_03 PARTITION OF incoming_logs
    FOR VALUES FROM ('2025-03-01 00:00:00') TO ('2025-04-01 00:00:00');
CREATE TABLE IF NOT EXISTS incoming_logs_2025_04 PARTITION OF incoming_logs
    FOR VALUES FROM ('2025-04-01 00:00:00') TO ('2025-05-01 00:00:00');
CREATE TABLE IF NOT EXISTS incoming_logs_2025_05 PARTITION OF incoming_logs
    FOR VALUES FROM ('2025-05-01 00:00:00') TO ('2025-06-01 00:00:00');
CREATE TABLE IF NOT EXISTS incoming_logs_2025_06 PARTITION OF incoming_logs
    FOR VALUES FROM ('2025-06-01 00:00:00') TO ('2025-07-01 00:00:00');
CREATE TABLE IF NOT EXISTS incoming_logs_2025_07 PARTITION OF incoming_logs
    FOR VALUES FROM ('2025-07-01 00:00:00') TO ('2025-08-01 00:00:00');
CREATE TABLE IF NOT EXISTS incoming_logs_2025_08 PARTITION OF incoming_logs
    FOR VALUES FROM ('2025-08-01 00:00:00') TO ('2025-09-01 00:00:00');
CREATE TABLE IF NOT EXISTS incoming_logs_2025_09 PARTITION OF incoming_logs
    FOR VALUES FROM ('2025-09-01 00:00:00') TO ('2025-10-01 00:00:00');
CREATE TABLE IF NOT EXISTS incoming_logs_2025_10 PARTITION OF incoming_logs
    FOR VALUES FROM ('2025-10-01 00:00:00') TO ('2025-11-01 00:00:00');
CREATE TABLE IF NOT EXISTS incoming_logs_2025_11 PARTITION OF incoming_logs
    FOR VALUES FROM ('2025-11-01 00:00:00') TO ('2025-12-01 00:00:00');
CREATE TABLE IF NOT EXISTS incoming_logs_2025_12 PARTITION OF incoming_logs
    FOR VALUES FROM ('2025-12-01 00:00:00') TO ('2026-01-01 00:00:00');

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_incoming_logs_line_account ON incoming_logs(line_account_id, incoming_time DESC);
CREATE INDEX IF NOT EXISTS idx_incoming_logs_group_id ON incoming_logs(group_id, incoming_time DESC);
CREATE INDEX IF NOT EXISTS idx_incoming_logs_incoming_line_id ON incoming_logs(incoming_line_id);
CREATE INDEX IF NOT EXISTS idx_incoming_logs_duplicate ON incoming_logs(is_duplicate);
CREATE INDEX IF NOT EXISTS idx_incoming_logs_time ON incoming_logs(incoming_time DESC);

-- 2.2 account_status_logs partition table
CREATE TABLE IF NOT EXISTS account_status_logs (
    id BIGSERIAL,
    line_account_id INTEGER NOT NULL,
    from_status VARCHAR(20) NOT NULL,
    to_status VARCHAR(20) NOT NULL,
    reason VARCHAR(50) NOT NULL,
    ip_address VARCHAR(50),
    occurred_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, occurred_at),
    CONSTRAINT check_from_status CHECK (from_status IN ('online', 'offline', 'user_logout', 'abnormal_offline')),
    CONSTRAINT check_to_status CHECK (to_status IN ('online', 'offline', 'user_logout', 'abnormal_offline')),
    CONSTRAINT check_reason CHECK (reason IN ('user_login', 'user_logout', 'abnormal_offline', 'force_offline'))
) PARTITION BY RANGE (occurred_at);

-- Create 2025 partitions
CREATE TABLE IF NOT EXISTS account_status_logs_2025_01 PARTITION OF account_status_logs
    FOR VALUES FROM ('2025-01-01 00:00:00') TO ('2025-02-01 00:00:00');
CREATE TABLE IF NOT EXISTS account_status_logs_2025_02 PARTITION OF account_status_logs
    FOR VALUES FROM ('2025-02-01 00:00:00') TO ('2025-03-01 00:00:00');
CREATE TABLE IF NOT EXISTS account_status_logs_2025_03 PARTITION OF account_status_logs
    FOR VALUES FROM ('2025-03-01 00:00:00') TO ('2025-04-01 00:00:00');
CREATE TABLE IF NOT EXISTS account_status_logs_2025_04 PARTITION OF account_status_logs
    FOR VALUES FROM ('2025-04-01 00:00:00') TO ('2025-05-01 00:00:00');
CREATE TABLE IF NOT EXISTS account_status_logs_2025_05 PARTITION OF account_status_logs
    FOR VALUES FROM ('2025-05-01 00:00:00') TO ('2025-06-01 00:00:00');
CREATE TABLE IF NOT EXISTS account_status_logs_2025_06 PARTITION OF account_status_logs
    FOR VALUES FROM ('2025-06-01 00:00:00') TO ('2025-07-01 00:00:00');
CREATE TABLE IF NOT EXISTS account_status_logs_2025_07 PARTITION OF account_status_logs
    FOR VALUES FROM ('2025-07-01 00:00:00') TO ('2025-08-01 00:00:00');
CREATE TABLE IF NOT EXISTS account_status_logs_2025_08 PARTITION OF account_status_logs
    FOR VALUES FROM ('2025-08-01 00:00:00') TO ('2025-09-01 00:00:00');
CREATE TABLE IF NOT EXISTS account_status_logs_2025_09 PARTITION OF account_status_logs
    FOR VALUES FROM ('2025-09-01 00:00:00') TO ('2025-10-01 00:00:00');
CREATE TABLE IF NOT EXISTS account_status_logs_2025_10 PARTITION OF account_status_logs
    FOR VALUES FROM ('2025-10-01 00:00:00') TO ('2025-11-01 00:00:00');
CREATE TABLE IF NOT EXISTS account_status_logs_2025_11 PARTITION OF account_status_logs
    FOR VALUES FROM ('2025-11-01 00:00:00') TO ('2025-12-01 00:00:00');
CREATE TABLE IF NOT EXISTS account_status_logs_2025_12 PARTITION OF account_status_logs
    FOR VALUES FROM ('2025-12-01 00:00:00') TO ('2026-01-01 00:00:00');

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_account_status_logs_account ON account_status_logs(line_account_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_account_status_logs_time ON account_status_logs(occurred_at DESC);

-- ============================================
-- 3. Create Triggers
-- ============================================

-- Create common trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to tables
DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
CREATE TRIGGER trigger_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_groups_updated_at ON groups;
CREATE TRIGGER trigger_groups_updated_at
    BEFORE UPDATE ON groups
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_line_accounts_updated_at ON line_accounts;
CREATE TRIGGER trigger_line_accounts_updated_at
    BEFORE UPDATE ON line_accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_customers_updated_at ON customers;
CREATE TRIGGER trigger_customers_updated_at
    BEFORE UPDATE ON customers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_follow_up_records_updated_at ON follow_up_records;
CREATE TRIGGER trigger_follow_up_records_updated_at
    BEFORE UPDATE ON follow_up_records
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_llm_configs_updated_at ON llm_configs;
CREATE TRIGGER trigger_llm_configs_updated_at
    BEFORE UPDATE ON llm_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_llm_prompt_templates_updated_at ON llm_prompt_templates;
CREATE TRIGGER trigger_llm_prompt_templates_updated_at
    BEFORE UPDATE ON llm_prompt_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_contact_pool_updated_at ON contact_pool;
CREATE TRIGGER trigger_contact_pool_updated_at
    BEFORE UPDATE ON contact_pool
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- 4. Create Views
-- ============================================

-- Leads list view
CREATE OR REPLACE VIEW v_leads_list AS
SELECT 
    g.id as group_id,
    g.activation_code,
    g.user_id,
    g.is_active,
    g.remark,
    g.description,
    g.category,
    g.dedup_scope,
    g.reset_time,
    g.created_at,
    g.last_login_at,
    g.account_limit,
    COALESCE(gs.total_accounts, 0) as total_accounts,
    COALESCE(gs.online_accounts, 0) as online_accounts,
    COALESCE(gs.line_accounts, 0) as line_accounts,
    COALESCE(gs.line_business_accounts, 0) as line_business_accounts,
    COALESCE(gs.today_incoming, 0) as today_incoming,
    COALESCE(gs.total_incoming, 0) as total_incoming,
    COALESCE(gs.duplicate_incoming, 0) as duplicate_incoming,
    COALESCE(gs.today_duplicate, 0) as today_duplicate
FROM groups g
LEFT JOIN group_stats gs ON gs.group_id = g.id
WHERE g.deleted_at IS NULL;

-- ============================================
-- 5. Create Partition Management Function
-- ============================================

CREATE OR REPLACE FUNCTION create_next_month_partitions()
RETURNS void AS $$
DECLARE
    next_month_start DATE;
    next_month_end DATE;
    partition_name TEXT;
BEGIN
    -- Calculate next month start date
    next_month_start := DATE_TRUNC('month', CURRENT_DATE + INTERVAL '1 month');
    next_month_end := next_month_start + INTERVAL '1 month';
    
    -- Create incoming_logs partition
    partition_name := 'incoming_logs_' || TO_CHAR(next_month_start, 'YYYY_MM');
    
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF incoming_logs
         FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        next_month_start,
        next_month_end
    );
    
    -- Create indexes
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%I_line_account ON %I(line_account_id, incoming_time DESC)', partition_name, partition_name);
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%I_group_id ON %I(group_id, incoming_time DESC)', partition_name, partition_name);
    
    -- Create account_status_logs partition
    partition_name := 'account_status_logs_' || TO_CHAR(next_month_start, 'YYYY_MM');
    
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF account_status_logs
         FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        next_month_start,
        next_month_end
    );
    
    -- Create indexes
    EXECUTE format('CREATE INDEX IF NOT EXISTS idx_%I_account ON %I(line_account_id, occurred_at DESC)', partition_name, partition_name);
    
    RAISE NOTICE 'Created partitions for %', TO_CHAR(next_month_start, 'YYYY-MM');
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 6. Initialize Data
-- ============================================

-- Initialize admin account
-- Default username: admin
-- Default password: admin123 (Please change in production)
-- Password hash is bcrypt encrypted, cost=10
--
-- Note: This password hash is for 'admin123'
-- In production, use Go's golang.org/x/crypto/bcrypt to generate
-- Example: bcrypt.GenerateFromPassword([]byte("admin123"), 10)

INSERT INTO users (username, password_hash, role, is_active, created_at, updated_at)
VALUES (
    'admin',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'admin',
    TRUE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (username) DO NOTHING;

-- Add column comments
COMMENT ON COLUMN line_accounts.add_friend_link IS 'Line添加好友链接';
COMMENT ON COLUMN line_accounts.reset_time IS 'Account reset time, use group reset_time if NULL';
COMMENT ON COLUMN contact_pool.deleted_at IS 'Soft delete timestamp';

