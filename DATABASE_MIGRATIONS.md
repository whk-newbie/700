# 数据库迁移说明

## 📋 概述

本项目使用**自动迁移机制**，在应用启动时自动检测并执行未执行的数据库迁移文件。

## 🔄 迁移机制

### 自动迁移流程

1. **应用启动时**：在数据库连接成功后，自动执行 `database.RunMigrations()`
2. **迁移检测**：扫描 `backend/migrations/` 目录下的所有 `.sql` 文件
3. **迁移记录**：在数据库中维护 `migration_records` 表，记录已执行的迁移
4. **自动执行**：只执行未记录的迁移文件，按文件名顺序执行

### 迁移文件命名规范

迁移文件必须遵循以下命名规范：

```
001_xxx.sql
002_xxx.sql
003_xxx.sql
...
```

- 文件名必须以数字前缀开头（用于排序）
- 数字前缀后跟下划线和描述性名称
- 文件扩展名必须是 `.sql`

### 迁移记录表

系统会自动创建 `migration_records` 表来跟踪已执行的迁移：

```sql
CREATE TABLE migration_records (
    id SERIAL PRIMARY KEY,
    filename VARCHAR(255) UNIQUE NOT NULL,
    applied_at BIGINT NOT NULL
);
```

## 📁 迁移文件列表

### 001_init_schema.sql
- **描述**：初始化数据库表结构
- **执行时机**：首次部署时
- **内容**：创建所有基础表、索引、触发器、函数等

### 002_init_admin.sql
- **描述**：初始化管理员账号
- **执行时机**：首次部署时
- **内容**：创建默认管理员账号（用户名: admin, 密码: admin123）

### 003_simplify_llm_configs.sql
- **描述**：简化 LLM 配置表结构
- **执行时机**：更新到支持 OpenAI API Key 加密的版本时
- **内容**：
  - 删除 `llm_configs` 表中不需要的字段
  - 只保留 `id`、`api_key`、`updated_at` 字段
  - 删除相关外键约束和索引
  - 删除 `llm_prompt_templates` 表

## 🚀 部署场景

### 场景1：全新部署

1. **首次启动**：
   - `001_init_schema.sql` 自动执行（通过 `docker-entrypoint-initdb.d`）
   - `002_init_admin.sql` 自动执行（通过 `docker-entrypoint-initdb.d`）
   - `003_simplify_llm_configs.sql` 自动执行（通过应用启动时的迁移机制）

2. **迁移记录**：
   - 所有迁移都会被记录到 `migration_records` 表
   - 后续启动不会重复执行

### 场景2：已有数据库升级

1. **已有数据库**：
   - `docker-entrypoint-initdb.d` 不会执行（数据库已存在）
   - 应用启动时，迁移机制会检测未执行的迁移

2. **自动执行**：
   - 只执行未记录的迁移文件（如 `003_simplify_llm_configs.sql`）
   - 已执行的迁移会被跳过

### 场景3：Docker 部署

#### 首次部署

```bash
# 1. 启动服务
docker-compose up -d

# 2. 查看迁移日志
docker-compose logs backend | grep -i migration
```

#### 更新部署

```bash
# 1. 拉取最新代码
git pull

# 2. 重新构建镜像
docker-compose build backend

# 3. 重启服务（迁移会自动执行）
docker-compose up -d backend

# 4. 查看迁移日志
docker-compose logs backend | grep -i migration
```

## 🔍 验证迁移状态

### 查看迁移记录

```sql
-- 连接到数据库
psql -U lineuser -d line_management

-- 查看已执行的迁移
SELECT * FROM migration_records ORDER BY applied_at;
```

### 查看应用日志

```bash
# Docker 环境
docker-compose logs backend | grep -i migration

# 本地环境
tail -f backend/logs/app.log | grep -i migration
```

## ⚠️ 注意事项

### 1. 迁移文件顺序

- 迁移文件必须按数字顺序命名
- 系统会按文件名排序执行
- 不要跳过数字（如：001, 002, 004 会出错）

### 2. 迁移文件幂等性

- 迁移 SQL 应该使用 `IF EXISTS`、`IF NOT EXISTS` 等语句
- 确保迁移可以安全地重复执行
- 示例：
  ```sql
  DROP TABLE IF EXISTS old_table;
  CREATE TABLE IF NOT EXISTS new_table (...);
  ALTER TABLE table_name DROP COLUMN IF EXISTS old_column;
  ```

### 3. 数据备份

- **重要**：在执行迁移前，建议备份数据库
- 特别是对于删除列、删除表等操作

### 4. 迁移失败处理

- 如果迁移失败，应用会记录错误并退出
- 需要手动修复问题后重新启动
- 检查 `migration_records` 表，确认哪些迁移已执行

## 🛠️ 手动执行迁移

如果需要手动执行迁移（不推荐），可以：

### 方法1：使用 psql

```bash
# 连接到数据库
psql -U lineuser -d line_management

# 执行迁移文件
\i backend/migrations/003_simplify_llm_configs.sql

# 手动记录迁移（如果需要）
INSERT INTO migration_records (filename, applied_at)
VALUES ('003_simplify_llm_configs.sql', EXTRACT(EPOCH FROM NOW()));
```

### 方法2：使用 Docker

```bash
# 复制迁移文件到容器
docker cp backend/migrations/003_simplify_llm_configs.sql line-mgmt-postgres:/tmp/

# 进入容器
docker exec -it line-mgmt-postgres psql -U lineuser -d line_management

# 执行迁移
\i /tmp/003_simplify_llm_configs.sql
```

## 📝 创建新迁移

1. **创建迁移文件**：
   ```bash
   # 在 backend/migrations/ 目录下创建新文件
   # 文件名格式：004_description.sql
   touch backend/migrations/004_add_new_feature.sql
   ```

2. **编写迁移 SQL**：
   ```sql
   -- 004_add_new_feature.sql
   -- 描述：添加新功能相关的表或字段
   
   -- 使用 IF NOT EXISTS 确保幂等性
   CREATE TABLE IF NOT EXISTS new_table (
       id SERIAL PRIMARY KEY,
       ...
   );
   ```

3. **测试迁移**：
   - 在开发环境测试
   - 确保迁移可以安全地重复执行
   - 验证迁移后的数据库结构

4. **部署**：
   - 提交代码
   - 部署到生产环境
   - 应用启动时会自动执行新迁移

## 🔗 相关文档

- [DATABASE.md](./DATABASE.md) - 数据库设计文档
- [DEPLOYMENT.md](./DEPLOYMENT.md) - 部署文档
- [docker-compose.yml](./docker-compose.yml) - Docker 配置

---

最后更新：2025-12-24

