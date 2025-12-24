# Line账号管理系统 - 数据库文档

## 📋 概述

本文档详细介绍Line账号分组管理与进线统计系统的数据库设计，包括表结构、索引策略、分区方案和设计原则。

## 🏗️ 数据库架构

### 技术栈
- **数据库**: PostgreSQL 15+
- **字符集**: UTF-8
- **时区**: Asia/Shanghai
- **连接池**: pgx

### 设计原则
1. **规范化设计**: 遵循第三范式，减少数据冗余
2. **分区策略**: 大表采用时间分区，提高查询性能
3. **索引优化**: 为常用查询字段建立合适的索引
4. **软删除**: 重要数据采用软删除策略
5. **审计字段**: 记录创建、更新、删除的审计信息

## 📊 表结构详解

### 核心表关系图

```
users (管理员/用户)
├── groups (分组) ──┬── group_stats (分组统计)
│                  ├── line_accounts (Line账号) ──┬── line_account_stats (账号统计)
│                  │                             └── customers (客户)
│                  │                             └── follow_up_records (跟进记录)
│                  ├── import_batches (导入批次) ──┘
│                  └── contact_pool (底库)
│
└── llm_configs (大模型配置)
    └── llm_templates (Prompt模板)
        └── llm_call_logs (调用日志)

incoming_logs (进线日志) - 分区表
account_status_logs (状态日志) - 分区表
```

## 📋 详细表结构

### 1. users - 用户表

**用途**: 存储系统用户（管理员和普通用户）

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | SERIAL | PRIMARY KEY | 用户ID |
| username | VARCHAR(50) | UNIQUE NOT NULL | 用户名 |
| password_hash | VARCHAR(255) | NOT NULL | 密码哈希 |
| email | VARCHAR(100) | - | 邮箱 |
| role | VARCHAR(20) | NOT NULL DEFAULT 'user' | 角色（admin/user） |
| max_groups | INTEGER | - | 最大分组数限制 |
| is_active | BOOLEAN | DEFAULT TRUE | 是否激活 |
| created_by | INTEGER | FK→users.id | 创建者ID |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 更新时间 |
| deleted_at | TIMESTAMP | - | 软删除时间 |

**索引**:
- `idx_users_username` - 用户名索引
- `idx_users_role` - 角色索引
- `idx_users_deleted` - 软删除索引

---

### 2. groups - 分组表

**用途**: 存储分组信息，每个分组对应一个激活码

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | SERIAL | PRIMARY KEY | 分组ID |
| user_id | INTEGER | NOT NULL FK→users.id | 所属用户ID |
| activation_code | VARCHAR(32) | UNIQUE NOT NULL | 激活码 |
| account_limit | INTEGER | - | 账号数量限制 |
| is_active | BOOLEAN | DEFAULT TRUE | 是否激活 |
| remark | VARCHAR(255) | - | 备注 |
| description | TEXT | - | 描述 |
| category | VARCHAR(50) | DEFAULT 'default' | 分类 |
| dedup_scope | VARCHAR(20) | DEFAULT 'current' | 去重范围（current/global） |
| reset_time | TIME | DEFAULT '09:00:00' | 每日重置时间 |
| login_password | VARCHAR(255) | - | 子账号登录密码 |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 更新时间 |
| last_login_at | TIMESTAMP | - | 最后登录时间 |
| deleted_at | TIMESTAMP | - | 软删除时间 |

**索引**:
- `idx_groups_activation_code` - 激活码唯一索引
- `idx_groups_user_id` - 用户ID索引
- `idx_groups_category` - 分类索引
- `idx_groups_is_active` - 激活状态索引
- `idx_groups_deleted` - 软删除索引

---

### 3. group_stats - 分组统计表

**用途**: 存储分组级别的统计数据

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | SERIAL | PRIMARY KEY | 统计ID |
| group_id | INTEGER | UNIQUE NOT NULL FK→groups.id | 分组ID |
| total_accounts | INTEGER | DEFAULT 0 | 总账号数 |
| online_accounts | INTEGER | DEFAULT 0 | 在线账号数 |
| line_accounts | INTEGER | DEFAULT 0 | Line账号数 |
| line_business_accounts | INTEGER | DEFAULT 0 | Line商务账号数 |
| today_incoming | INTEGER | DEFAULT 0 | 今日进线数 |
| total_incoming | INTEGER | DEFAULT 0 | 总进线数 |
| duplicate_incoming | INTEGER | DEFAULT 0 | 重复进线数 |
| today_duplicate | INTEGER | DEFAULT 0 | 今日重复进线数 |
| last_reset_date | DATE | - | 最后重置日期 |
| last_reset_time | TIMESTAMP | - | 最后重置时间 |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 更新时间 |

**索引**:
- `idx_group_stats_unique` - 分组ID唯一索引

---

### 4. line_accounts - Line账号表

**用途**: 存储Line账号信息

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | SERIAL | PRIMARY KEY | 账号ID |
| group_id | INTEGER | NOT NULL FK→groups.id | 分组ID |
| activation_code | VARCHAR(32) | NOT NULL | 激活码 |
| platform_type | VARCHAR(20) | NOT NULL DEFAULT 'line' | 平台类型 |
| line_id | VARCHAR(100) | NOT NULL | Line ID |
| display_name | VARCHAR(100) | - | 显示名称 |
| phone_number | VARCHAR(20) | - | 手机号 |
| profile_url | VARCHAR(500) | - | 个人资料链接 |
| avatar_url | VARCHAR(500) | - | 头像链接 |
| bio | TEXT | - | 个人简介 |
| status_message | VARCHAR(255) | - | 状态消息 |
| add_friend_link | VARCHAR(500) | - | 添加好友链接 |
| qr_code_path | VARCHAR(255) | - | 二维码路径 |
| online_status | VARCHAR(20) | DEFAULT 'offline' | 在线状态 |
| reset_time | TIME | - | 账号独立重置时间 |
| last_active_at | TIMESTAMP | - | 最后活跃时间 |
| last_online_time | TIMESTAMP | - | 最后在线时间 |
| first_login_at | TIMESTAMP | - | 首次登录时间 |
| account_remark | TEXT | - | 账号备注 |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 更新时间 |
| deleted_at | TIMESTAMP | - | 软删除时间 |
| deleted_by | INTEGER | FK→users.id | 删除者ID |

**约束**:
- `check_platform_type`: 平台类型必须为 'line' 或 'line_business'
- `check_online_status`: 在线状态必须为指定值之一

**索引**:
- `idx_line_accounts_unique` - 分组+LineID唯一索引
- `idx_line_accounts_group_id` - 分组ID索引
- `idx_line_accounts_activation_code` - 激活码索引
- `idx_line_accounts_platform_type` - 平台类型索引
- `idx_line_accounts_online_status` - 在线状态索引
- `idx_line_accounts_line_id` - LineID索引
- `idx_line_accounts_deleted` - 软删除索引

---

### 5. line_account_stats - Line账号统计表

**用途**: 存储账号级别的统计数据

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | SERIAL | PRIMARY KEY | 统计ID |
| line_account_id | INTEGER | UNIQUE NOT NULL FK→line_accounts.id | 账号ID |
| today_incoming | INTEGER | DEFAULT 0 | 今日进线数 |
| total_incoming | INTEGER | DEFAULT 0 | 总进线数 |
| duplicate_incoming | INTEGER | DEFAULT 0 | 重复进线数 |
| today_duplicate | INTEGER | DEFAULT 0 | 今日重复进线数 |
| last_reset_date | DATE | - | 最后重置日期 |
| last_reset_time | TIMESTAMP | - | 最后重置时间 |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 更新时间 |

**索引**:
- `idx_line_account_stats_unique` - 账号ID唯一索引

---

### 6. incoming_logs - 进线日志表（分区表）

**用途**: 记录所有的进线数据，采用月度分区

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 日志ID |
| line_account_id | INTEGER | NOT NULL FK→line_accounts.id | 账号ID |
| group_id | INTEGER | NOT NULL FK→groups.id | 分组ID |
| activation_code | VARCHAR(32) | NOT NULL | 激活码 |
| platform_type | VARCHAR(20) | NOT NULL | 平台类型 |
| incoming_line_id | VARCHAR(100) | NOT NULL | 进线Line ID |
| display_name | VARCHAR(100) | - | 显示名称 |
| avatar_url | VARCHAR(500) | - | 头像链接 |
| phone_number | VARCHAR(20) | - | 手机号 |
| incoming_time | TIMESTAMP | NOT NULL | 进线时间 |
| is_duplicate | BOOLEAN | DEFAULT FALSE | 是否重复 |
| duplicate_type | VARCHAR(20) | - | 重复类型（current/global） |
| source_type | VARCHAR(20) | DEFAULT 'platform' | 来源类型 |
| remark | TEXT | - | 备注 |
| metadata | JSONB | - | 元数据 |

**分区策略**: 按月分区（incoming_logs_yyyy_mm）

---

### 7. contact_pool - 底库表

**用途**: 存储导入的联系人信息，作为去重依据

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 联系人ID |
| source_type | VARCHAR(20) | NOT NULL | 来源类型（import/platform） |
| import_batch_id | INTEGER | FK→import_batches.id | 导入批次ID |
| group_id | INTEGER | NOT NULL FK→groups.id | 分组ID |
| activation_code | VARCHAR(32) | NOT NULL | 激活码 |
| line_account_id | INTEGER | FK→line_accounts.id | 账号ID |
| platform_type | VARCHAR(20) | NOT NULL | 平台类型 |
| line_id | VARCHAR(100) | NOT NULL | Line ID |
| display_name | VARCHAR(100) | - | 显示名称 |
| phone_number | VARCHAR(20) | - | 手机号 |
| avatar_url | VARCHAR(500) | - | 头像链接 |
| dedup_scope | VARCHAR(20) | - | 去重范围 |
| first_seen_at | TIMESTAMP | - | 首次发现时间 |
| remark | TEXT | - | 备注 |
| metadata | JSONB | - | 元数据 |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 更新时间 |
| deleted_at | TIMESTAMP | - | 软删除时间 |

**索引**:
- `idx_contact_pool_global_unique` - 全局唯一索引（line_id + platform_type）
- `idx_contact_pool_group_id` - 分组ID索引
- `idx_contact_pool_activation_code` - 激活码索引
- `idx_contact_pool_line_id` - LineID索引
- `idx_contact_pool_deleted` - 软删除索引

---

### 8. customers - 客户表

**用途**: 存储客户详细信息

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 客户ID |
| group_id | INTEGER | NOT NULL FK→groups.id | 分组ID |
| activation_code | VARCHAR(32) | NOT NULL | 激活码 |
| line_account_id | INTEGER | FK→line_accounts.id | 账号ID |
| platform_type | VARCHAR(20) | NOT NULL | 平台类型 |
| customer_id | VARCHAR(100) | NOT NULL | 客户ID |
| display_name | VARCHAR(100) | - | 显示名称 |
| avatar_url | VARCHAR(500) | - | 头像链接 |
| phone_number | VARCHAR(20) | - | 手机号 |
| customer_type | VARCHAR(50) | - | 客户类型 |
| gender | VARCHAR(10) | - | 性别 |
| country | VARCHAR(50) | - | 国家 |
| birthday | DATE | - | 生日 |
| address | TEXT | - | 地址 |
| nickname_remark | VARCHAR(20) | - | 昵称备注 |
| remark | TEXT | - | 备注 |
| tags | JSONB | - | 标签 |
| metadata | JSONB | - | 元数据 |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 更新时间 |
| deleted_at | TIMESTAMP | - | 软删除时间 |

---

### 9. follow_up_records - 跟进记录表

**用途**: 存储客户跟进记录

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 记录ID |
| group_id | INTEGER | NOT NULL FK→groups.id | 分组ID |
| activation_code | VARCHAR(32) | NOT NULL | 激活码 |
| line_account_id | INTEGER | FK→line_accounts.id | 账号ID |
| customer_id | INTEGER | FK→customers.id | 客户ID |
| platform_type | VARCHAR(20) | NOT NULL | 平台类型 |
| follow_up_type | VARCHAR(50) | - | 跟进类型 |
| content | TEXT | NOT NULL | 跟进内容 |
| contact_method | VARCHAR(20) | - | 联系方式 |
| next_follow_up_time | TIMESTAMP | - | 下次跟进时间 |
| status | VARCHAR(20) | DEFAULT 'pending' | 状态 |
| remark | TEXT | - | 备注 |
| metadata | JSONB | - | 元数据 |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 更新时间 |
| deleted_at | TIMESTAMP | - | 软删除时间 |

---

### 10. import_batches - 导入批次表

**用途**: 记录联系人导入批次信息

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | SERIAL | PRIMARY KEY | 批次ID |
| batch_name | VARCHAR(100) | - | 批次名称 |
| platform_type | VARCHAR(20) | NOT NULL | 平台类型 |
| total_count | INTEGER | DEFAULT 0 | 总数 |
| success_count | INTEGER | DEFAULT 0 | 成功数 |
| duplicate_count | INTEGER | DEFAULT 0 | 重复数 |
| error_count | INTEGER | DEFAULT 0 | 错误数 |
| dedup_scope | VARCHAR(20) | - | 去重范围 |
| file_name | VARCHAR(255) | - | 文件名 |
| file_path | VARCHAR(500) | - | 文件路径 |
| file_size | BIGINT | - | 文件大小 |
| imported_by | INTEGER | FK→users.id | 导入者ID |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| completed_at | TIMESTAMP | - | 完成时间 |

---

### 11. 大模型相关表

#### llm_configs - 大模型配置表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | SERIAL | PRIMARY KEY | 配置ID |
| user_id | INTEGER | NOT NULL FK→users.id | 用户ID |
| provider | VARCHAR(50) | NOT NULL | 提供商 |
| name | VARCHAR(100) | NOT NULL | 配置名称 |
| api_key | TEXT | NOT NULL | API Key（加密存储） |
| base_url | VARCHAR(500) | - | 基础URL |
| model | VARCHAR(100) | - | 模型名称 |
| max_tokens | INTEGER | - | 最大Token数 |
| temperature | DECIMAL(3,2) | - | 温度参数 |
| is_active | BOOLEAN | DEFAULT TRUE | 是否激活 |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 更新时间 |

#### llm_templates - Prompt模板表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | SERIAL | PRIMARY KEY | 模板ID |
| user_id | INTEGER | NOT NULL FK→users.id | 用户ID |
| config_id | INTEGER | FK→llm_configs.id | 配置ID |
| name | VARCHAR(100) | NOT NULL | 模板名称 |
| template | TEXT | NOT NULL | 模板内容 |
| variables | JSONB | - | 变量定义 |
| description | TEXT | - | 描述 |
| is_active | BOOLEAN | DEFAULT TRUE | 是否激活 |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 更新时间 |

#### llm_call_logs - 调用日志表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 日志ID |
| user_id | INTEGER | NOT NULL FK→users.id | 用户ID |
| config_id | INTEGER | FK→llm_configs.id | 配置ID |
| template_id | INTEGER | FK→llm_templates.id | 模板ID |
| activation_code | VARCHAR(32) | - | 激活码 |
| provider | VARCHAR(50) | NOT NULL | 提供商 |
| model | VARCHAR(100) | - | 模型 |
| status | VARCHAR(20) | NOT NULL | 状态 |
| prompt_tokens | INTEGER | - | Prompt Token数 |
| completion_tokens | INTEGER | - | Completion Token数 |
| total_tokens | INTEGER | - | 总Token数 |
| duration_ms | INTEGER | - | 调用耗时（毫秒） |
| error_message | TEXT | - | 错误信息 |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

---

### 12. account_status_logs - 账号状态日志表（分区表）

**用途**: 记录账号状态变化历史，采用月度分区

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGSERIAL | PRIMARY KEY | 日志ID |
| line_account_id | INTEGER | NOT NULL FK→line_accounts.id | 账号ID |
| group_id | INTEGER | NOT NULL FK→groups.id | 分组ID |
| activation_code | VARCHAR(32) | NOT NULL | 激活码 |
| platform_type | VARCHAR(20) | NOT NULL | 平台类型 |
| line_id | VARCHAR(100) | NOT NULL | Line ID |
| old_status | VARCHAR(20) | - | 旧状态 |
| new_status | VARCHAR(20) | NOT NULL | 新状态 |
| change_reason | VARCHAR(100) | - | 变更原因 |
| change_time | TIMESTAMP | NOT NULL | 变更时间 |
| metadata | JSONB | - | 元数据 |

**分区策略**: 按月分区（account_status_logs_yyyy_mm）

## 🗂️ 分区策略

### 分区表
1. **incoming_logs**: 按月分区（incoming_logs_yyyy_mm）
2. **account_status_logs**: 按月分区（account_status_logs_yyyy_mm）

### 分区管理
- **自动创建**: 每月1号凌晨2点创建下月分区
- **数据归档**: 每晚4点归档12个月前的数据
- **分区函数**: `create_next_month_partitions()`

### 分区优势
- **查询性能**: 大幅提升历史数据查询速度
- **维护效率**: 删除过期数据只需删除整个分区
- **存储管理**: 可以针对不同分区设置不同的存储策略

## 🔍 索引策略

### 索引类型
1. **唯一索引**: 确保数据唯一性
2. **普通索引**: 加速查询
3. **复合索引**: 多字段查询优化
4. **部分索引**: 只对特定条件建立索引

### 关键索引说明
- **激活码索引**: 支持快速查找分组和账号
- **软删除索引**: 只索引未删除的记录
- **时间索引**: 支持时间范围查询
- **分组索引**: 支持按分组过滤数据

## 🔒 数据安全

### 加密策略
- **API Key**: 使用AES-256-GCM加密存储
- **密码**: 使用bcrypt哈希存储
- **敏感数据**: 在传输和存储过程中加密

### 访问控制
- **行级安全**: 根据用户角色过滤数据
- **字段级权限**: 敏感字段只对特定角色可见
- **审计日志**: 记录所有重要操作

## 📊 性能优化

### 查询优化
- **分页查询**: 使用OFFSET + LIMIT，支持大结果集
- **预编译查询**: 避免SQL注入，提高执行效率
- **连接池**: 复用数据库连接，减少开销

### 统计优化
- **增量更新**: 实时更新统计数据，避免全量计算
- **缓存策略**: 使用Redis缓存热点数据
- **异步处理**: 统计计算异步执行，不阻塞主流程

## 🛠️ 维护脚本

### 定期任务
1. **每日重置**: 根据分组重置时间重置统计数据
2. **全量校准**: 每周校准所有统计数据
3. **离线检测**: 每5分钟检测离线账号
4. **分区管理**: 每月创建新分区
5. **数据归档**: 每月归档过期数据

### 监控指标
- **连接数**: 监控数据库连接使用情况
- **查询性能**: 监控慢查询
- **存储使用**: 监控表和索引大小
- **分区状态**: 监控分区创建和使用情况

## 📝 使用建议

### 开发环境
```sql
-- 创建数据库
CREATE DATABASE line_management WITH ENCODING 'UTF8';

-- 执行初始化脚本
\i migrations/001_init_schema.sql
\i migrations/002_init_admin.sql
```

### 生产环境
- 定期备份重要数据
- 监控数据库性能指标
- 及时清理过期日志
- 维护合适的索引

---

最后更新：2025-12-24
