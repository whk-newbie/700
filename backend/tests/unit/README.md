# 单元测试说明

## 测试环境准备

### 1. 创建测试数据库

在运行测试之前，需要创建一个独立的测试数据库：

```sql
CREATE DATABASE line_management_test;
```

### 2. 配置数据库连接

测试使用的数据库连接配置在 `helper.go` 中：

```go
cfg := &config.Config{
    Database: config.DatabaseConfig{
        Host:     "localhost",
        Port:     5432,
        User:     "postgres",
        Password: "postgres",
        DBName:   "line_management_test",
        SSLMode:  "disable",
    },
}
```

如果你的数据库配置不同，请修改 `helper.go` 中的配置。

### 3. 初始化测试数据库表结构

运行测试前，需要在测试数据库中创建表结构：

```bash
cd backend
psql -U postgres -d line_management_test -f migrations/001_init_schema.sql
```

## 运行测试

### 运行所有单元测试

```bash
cd backend
go test ./tests/unit/... -v
```

### 运行特定测试文件

```bash
cd backend
go test ./tests/unit/dedup_service_test.go ./tests/unit/helper.go -v
go test ./tests/unit/stats_service_test.go ./tests/unit/helper.go -v
go test ./tests/unit/incoming_service_test.go ./tests/unit/helper.go -v
go test ./tests/unit/auth_service_test.go ./tests/unit/helper.go -v
go test ./tests/unit/group_service_test.go ./tests/unit/helper.go -v
```

### 运行特定测试套件

```bash
cd backend
go test ./tests/unit/... -v -run TestDedupServiceTestSuite
go test ./tests/unit/... -v -run TestStatsServiceTestSuite
go test ./tests/unit/... -v -run TestIncomingServiceTestSuite
go test ./tests/unit/... -v -run TestAuthServiceTestSuite
go test ./tests/unit/... -v -run TestGroupServiceTestSuite
```

### 运行特定测试用例

```bash
cd backend
go test ./tests/unit/... -v -run TestDedupServiceTestSuite/TestCheckDuplicateCurrent_NoDuplicate
```

### 查看测试覆盖率

```bash
cd backend
go test ./tests/unit/... -cover
go test ./tests/unit/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 测试文件说明

### helper.go
测试辅助工具，包含：
- 测试数据库初始化
- 测试数据清理
- 测试数据创建辅助函数

### dedup_service_test.go
去重逻辑单元测试，覆盖：
- 当前分组去重
- 全局去重
- 底库去重
- 不同场景的去重判断

### stats_service_test.go
统计计算单元测试，覆盖：
- 分组统计获取
- 账号统计获取
- 总览统计
- 进线趋势计算

### incoming_service_test.go
进线处理单元测试，覆盖：
- 进线数据处理
- 去重判断
- 统计更新
- 底库添加
- 事务处理

### auth_service_test.go
认证服务单元测试，覆盖：
- 用户认证
- 子账号认证
- 密码验证
- 用户状态检查

### group_service_test.go
分组服务单元测试，覆盖：
- 分组创建
- 分组查询
- 分组更新
- 分组删除
- 激活码生成
- 批量操作

## 测试数据清理

每个测试用例执行前都会自动清理测试数据，确保测试的独立性和可重复性。

## 注意事项

1. **不要在生产数据库上运行测试**：测试会清理所有数据
2. **测试数据库独立**：使用专门的测试数据库 `line_management_test`
3. **测试隔离**：每个测试用例都是独立的，不依赖其他测试
4. **并发安全**：测试可以并发运行（使用 `-p` 参数）

## 常见问题

### 1. 数据库连接失败

检查 `helper.go` 中的数据库配置是否正确。

### 2. 表不存在

运行测试前需要先创建表结构：
```bash
psql -U postgres -d line_management_test -f migrations/001_init_schema.sql
```

### 3. 测试失败

查看详细错误信息：
```bash
go test ./tests/unit/... -v
```

### 4. 清理测试数据

测试会自动清理数据，如果需要手动清理：
```sql
DROP DATABASE line_management_test;
CREATE DATABASE line_management_test;
```

