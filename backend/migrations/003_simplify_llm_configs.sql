-- 简化llm_configs表结构，只保留OpenAI API Key
-- 迁移说明：删除不需要的字段，只保留id、api_key和updated_at

-- 1. 备份现有数据（如果需要）
-- CREATE TABLE llm_configs_backup AS SELECT * FROM llm_configs;

-- 2. 删除外键约束（如果有）
ALTER TABLE llm_prompt_templates DROP CONSTRAINT IF EXISTS llm_prompt_templates_config_id_fkey;
ALTER TABLE llm_call_logs DROP CONSTRAINT IF EXISTS llm_call_logs_config_id_fkey;

-- 3. 删除不需要的索引
DROP INDEX IF EXISTS idx_llm_configs_provider;
DROP INDEX IF EXISTS idx_llm_configs_is_active;

-- 4. 删除约束
ALTER TABLE llm_configs DROP CONSTRAINT IF EXISTS check_provider;

-- 5. 删除不需要的列
ALTER TABLE llm_configs 
    DROP COLUMN IF EXISTS name,
    DROP COLUMN IF EXISTS provider,
    DROP COLUMN IF EXISTS api_url,
    DROP COLUMN IF EXISTS model,
    DROP COLUMN IF EXISTS max_tokens,
    DROP COLUMN IF EXISTS temperature,
    DROP COLUMN IF EXISTS top_p,
    DROP COLUMN IF EXISTS frequency_penalty,
    DROP COLUMN IF EXISTS presence_penalty,
    DROP COLUMN IF EXISTS system_prompt,
    DROP COLUMN IF EXISTS timeout_seconds,
    DROP COLUMN IF EXISTS max_retries,
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS created_by;

-- 6. 确保api_key和updated_at字段存在
ALTER TABLE llm_configs 
    ALTER COLUMN api_key SET NOT NULL,
    ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;

-- 7. 只保留一条记录（如果有多条，保留最新的）
DO $$
DECLARE
    record_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO record_count FROM llm_configs;
    IF record_count > 1 THEN
        -- 删除除ID最小的记录外的所有记录
        DELETE FROM llm_configs 
        WHERE id NOT IN (
            SELECT MIN(id) FROM llm_configs
        );
    END IF;
END $$;

-- 8. 更新触发器（如果存在）
DROP TRIGGER IF EXISTS trigger_llm_configs_updated_at ON llm_configs;
CREATE TRIGGER trigger_llm_configs_updated_at
    BEFORE UPDATE ON llm_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 注意：llm_prompt_templates和llm_call_logs表中的config_id字段仍然存在
-- 但由于不再使用这些功能，可以保留字段但不使用，或者后续删除这些表

