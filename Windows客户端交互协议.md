# Windows客户端交互协议

> **版本**: v1.0
> **更新日期**: 2025-12-21
> **说明**: Windows客户端与服务器的完整交互协议设计

---

## 📋 协议概述

### 通信方式
- **WebSocket**: 实时数据上报、心跳保持连接
- **HTTP/HTTPS**: 登录认证、批量数据上传

### 认证方式
- 使用激活码进行认证
- 支持一个客户端登录多个激活码

---

## 🔐 HTTP 认证接口

### 1. 激活码登录

**接口**: `POST /api/client/login`

**请求**:
```json
{
  "activation_code": "ABC123"
}
```

**响应**:
```json
{
  "success": true,
  "group_id": 1,
  "group_name": "分组1",
  "remark": "德华2",
  "ws_url": "wss://yourdomain.com/api/ws/client",
  "token": "eyJhbGc..."
}
```

**说明**:
- 验证激活码是否有效且未被禁用
- 返回WebSocket连接地址和临时token
- 客户端使用token建立WebSocket连接

---

## 🔌 WebSocket 连接

### 1. 建立连接

**连接URL**: 
```
wss://yourdomain.com/api/ws/client?activation_code={code}&token={token}
```

**连接参数**:
- `activation_code`: 激活码
- `token`: 登录时获取的临时token

**连接成功响应**:
```json
{
  "type": "auth_success",
  "data": {
    "group_id": 1,
    "activation_code": "ABC123",
    "message": "认证成功，请同步Line账号列表"
  }
}
```

### 2. 多激活码连接

**客户端可以同时建立多个WebSocket连接**:
```
ws1: wss://domain.com/api/ws/client?activation_code=ABC123&token=xxx
ws2: wss://domain.com/api/ws/client?activation_code=DEF456&token=yyy
ws3: wss://domain.com/api/ws/client?activation_code=GHI789&token=zzz
```

每个连接独立管理，互不影响。

---

## 📤 客户端 → 服务器消息

### 1. 心跳包（每60秒）

**消息类型**: `heartbeat`

```json
{
  "type": "heartbeat",
  "activation_code": "ABC123",
  "timestamp": 1703123456
}
```

**说明**:
- 每60秒发送一次
- 服务器更新该激活码的最后活跃时间
- 超过65秒未收到心跳，标记为离线

---

### 2. 同步Line账号列表

**消息类型**: `sync_line_accounts`

**触发时机**: 
- WebSocket连接成功后
- 检测到本地Line账号有变化时

```json
{
  "type": "sync_line_accounts",
  "activation_code": "ABC123",
  "data": [
    {
      "line_id": "@line001",
      "display_name": "张三",
      "phone_number": "+886123456789",
      "platform_type": "line",
      "profile_url": "https://line.me/R/ti/p/@line001",
      "avatar_url": "https://...",
      "online_status": "online"
    },
    {
      "line_id": "@line002",
      "display_name": "李四",
      "platform_type": "line_business",
      "profile_url": "https://line.me/R/ti/p/@line002"
    }
  ]
}
```

**必填字段**:
- `line_id`: Line账号的唯一标识
- `platform_type`: 平台类型（line / line_business）

**可选字段**:
- `display_name`: 显示名称
- `phone_number`: 手机号
- `profile_url`: 主页地址
- `avatar_url`: 头像URL
- `online_status`: 在线状态

**服务器处理**:
1. 根据 `activation_code` 找到对应的分组（group_id）
2. 对每个Line账号：
   - 如果 `line_id` 已存在 → 更新账号信息
   - 如果不存在 → 创建新账号记录，关联到该分组
3. 自动生成二维码（根据profile_url）
4. 返回同步结果

---

### 3. 上报进线数据

**消息类型**: `incoming`

**触发时机**: 检测到有人加好友（进线）

```json
{
  "type": "incoming",
  "activation_code": "ABC123",
  "data": {
    "line_account_id": "@line001",         // Line账号的line_id（不是数据库ID）
    "incoming_line_id": "U123456789",      // 进线客户的Line User ID
    "timestamp": "2025-12-21 10:30:00",
    
    // 以下为可选字段
    "display_name": "王五",
    "avatar_url": "https://...",
    "phone_number": "+886999888777"
  }
}
```

**必填字段**:
- `line_account_id`: 哪个Line账号收到的进线（使用line_id标识）
- `incoming_line_id`: 进线客户的Line User ID
- `timestamp`: 进线时间

**可选字段**（尽量上报）:
- `display_name`: 客户显示名称
- `avatar_url`: 客户头像
- `phone_number`: 客户手机号

**服务器处理**:
1. 通过 `activation_code` + `line_account_id` 找到对应的Line账号记录
2. 检查去重范围，判断是否重复
3. 记录进线日志
4. 更新统计数据
5. 如果不重复，添加到底库
6. 推送实时更新到前端

---

### 4. 上报客户信息（客户画像）

**消息类型**: `customer_sync`

**触发时机**: 在Line上为客户添加了画像信息

```json
{
  "type": "customer_sync",
  "activation_code": "ABC123",
  "data": {
    "line_account_id": "@line001",
    "customer_id": "U123456789",
    "display_name": "张三",
    "avatar_url": "https://...",
    "phone_number": "+886123456789",
    "gender": "male",
    "country": "Taiwan",
    "birthday": "1990-01-01",
    "address": "台北市...",
    "remark": "VIP客户"
  }
}
```

**服务器处理**:
1. 通过 `activation_code` + `line_account_id` 找到Line账号
2. 通过 `customer_id` 查找或创建客户记录
3. 更新客户信息
4. 客户类型标记为"新增线索-实时"

---

### 5. 上报跟进记录

**消息类型**: `follow_up_sync`

**触发时机**: 在Line上为客户添加了跟进记录

```json
{
  "type": "follow_up_sync",
  "activation_code": "ABC123",
  "data": {
    "line_account_id": "@line001",
    "customer_id": "U123456789",
    "content": "已联系客户，客户表示有兴趣",
    "timestamp": "2025-12-21 11:00:00"
  }
}
```

**服务器处理**:
1. 通过 `activation_code` + `line_account_id` 找到Line账号
2. 通过 `customer_id` 找到客户记录
3. 创建跟进记录

---

### 6. Line账号状态变化

**消息类型**: `account_status_change`

**触发时机**: Line账号登录或退出

```json
{
  "type": "account_status_change",
  "activation_code": "ABC123",
  "data": {
    "line_account_id": "@line001",
    "online_status": "online",           // 'online' | 'user_logout' | 'abnormal_offline'
    "timestamp": "2025-12-21 12:00:00"
  }
}
```

**服务器处理**:
1. 更新Line账号的在线状态
2. 记录状态变化日志（account_status_logs表）
3. 推送实时更新到前端

---

## 📥 服务器 → 客户端消息

### 1. 认证成功
```json
{
  "type": "auth_success",
  "data": {
    "group_id": 1,
    "activation_code": "ABC123",
    "message": "认证成功，请同步Line账号列表"
  }
}
```

### 2. 认证失败
```json
{
  "type": "auth_error",
  "message": "激活码无效或已被禁用"
}
```

### 3. 账号同步结果
```json
{
  "type": "sync_result",
  "data": {
    "success": true,
    "created_count": 2,
    "updated_count": 1,
    "accounts": [
      {
        "line_id": "@line001",
        "account_id": 123,
        "status": "created"
      }
    ]
  }
}
```

### 4. 强制下线指令
```json
{
  "type": "force_offline",
  "data": {
    "line_account_id": "@line001",
    "reason": "管理员操作"
  }
}
```

### 5. 配置更新通知
```json
{
  "type": "config_update",
  "action": "reload_settings",
  "message": "分组配置已更新"
}
```

---

## 🔄 数据归属与隔离

### 归属规则

**每条数据的归属通过以下方式确定**:

```
激活码 + Line ID → 确定数据归属

示例：
- 激活码ABC123 + Line账号@line001 → 分组1的Line账号1
- 激活码DEF456 + Line账号@line002 → 分组2的Line账号2
- 激活码ABC123 + 进线UVW789 + Line账号@line001 → 分组1的进线记录
```

### 数据表关联

```sql
-- Line账号归属
line_accounts.group_id = 通过activation_code查询得到的group_id
line_accounts.activation_code = 上报时的activation_code

-- 进线记录归属
incoming_logs.line_account_id = 通过activation_code + line_id查询得到的line_account_id

-- 客户归属
customers.group_id = 通过activation_code查询得到的group_id
customers.line_account_id = 通过activation_code + line_id查询得到的line_account_id

-- 跟进记录归属
follow_up_records.group_id = 通过activation_code查询得到的group_id
follow_up_records.line_account_id = 通过activation_code + line_id查询得到的line_account_id
```

---

## 🔄 完整交互流程示例

### 场景：客户端添加新激活码并上报数据

```
第1步：客户端HTTP登录
POST /api/client/login
{
  "activation_code": "ABC123"
}

响应：
{
  "success": true,
  "group_id": 1,
  "token": "xxx"
}

第2步：建立WebSocket连接
wss://domain.com/api/ws/client?activation_code=ABC123&token=xxx

服务器返回：
{
  "type": "auth_success",
  "data": { "group_id": 1, "activation_code": "ABC123" }
}

第3步：客户端同步Line账号列表
{
  "type": "sync_line_accounts",
  "activation_code": "ABC123",
  "data": [
    { "line_id": "@line001", "display_name": "张三", "platform_type": "line" },
    { "line_id": "@line002", "display_name": "李四", "platform_type": "line_business" }
  ]
}

服务器处理：
- 创建或更新Line账号记录
- 关联到group_id=1
- 生成二维码

服务器返回：
{
  "type": "sync_result",
  "data": { "success": true, "created_count": 2 }
}

第4步：客户端发送心跳（每60秒）
{
  "type": "heartbeat",
  "activation_code": "ABC123",
  "timestamp": 1703123456
}

第5步：检测到进线，上报进线数据
{
  "type": "incoming",
  "activation_code": "ABC123",
  "data": {
    "line_account_id": "@line001",
    "incoming_line_id": "U123456789",
    "display_name": "王五",
    "timestamp": "2025-12-21 10:30:00"
  }
}

服务器处理：
- 通过ABC123 + @line001找到对应的line_account记录
- 检查去重范围判断是否重复
- 记录进线日志
- 更新统计数据
- 推送更新到前端

第6步：客户端上报客户画像
{
  "type": "customer_sync",
  "activation_code": "ABC123",
  "data": {
    "line_account_id": "@line001",
    "customer_id": "U123456789",
    "display_name": "王五",
    "phone_number": "+886999888777",
    "gender": "male",
    "country": "Taiwan"
  }
}

第7步：客户端上报跟进记录
{
  "type": "follow_up_sync",
  "activation_code": "ABC123",
  "data": {
    "line_account_id": "@line001",
    "customer_id": "U123456789",
    "content": "已联系客户，客户表示有兴趣"
  }
}
```

---

## 🎯 数据归属确定逻辑（关键）

### 服务器端处理逻辑

```python
# 伪代码示例

def handle_incoming(message):
    activation_code = message['activation_code']
    line_account_line_id = message['data']['line_account_id']
    incoming_line_id = message['data']['incoming_line_id']
    
    # 1. 通过激活码找到分组
    group = Group.query.filter_by(activation_code=activation_code).first()
    
    # 2. 通过激活码 + line_id 找到Line账号记录
    line_account = LineAccount.query.filter_by(
        group_id=group.id,
        line_id=line_account_line_id
    ).first()
    
    # 3. 检查去重
    is_duplicate = check_duplicate(
        group_id=group.id,
        incoming_line_id=incoming_line_id,
        dedup_scope=group.dedup_scope
    )
    
    # 4. 记录进线
    log = IncomingLog(
        line_account_id=line_account.id,
        group_id=group.id,
        incoming_line_id=incoming_line_id,
        is_duplicate=is_duplicate
    )
    db.session.add(log)
    
    # 5. 更新统计
    update_stats(line_account.id, group.id)
    
    # 6. 添加到底库（如果不重复）
    if not is_duplicate:
        add_to_contact_pool(group.id, incoming_line_id, ...)
```

---

## 📊 去重范围说明

### 两种去重范围

#### 1. 当前激活码（本分组去重）
- 只检查该分组下的进线历史
- 只检查该分组的底库数据

```python
def check_duplicate_current(group_id, incoming_line_id):
    # 检查该分组的底库
    exists_in_pool = ContactPool.query.filter_by(
        group_id=group_id,
        line_id=incoming_line_id
    ).first()
    
    return exists_in_pool is not None
```

#### 2. 全局去重
- 检查所有分组的进线历史
- 检查所有分组的底库数据

```python
def check_duplicate_global(incoming_line_id):
    # 检查全局底库
    exists_in_pool = ContactPool.query.filter_by(
        line_id=incoming_line_id
    ).first()
    
    return exists_in_pool is not None
```

---

## 🚀 客户端开发要点

### 需要实现的功能

1. **多激活码管理**
   - 支持添加/删除激活码
   - 每个激活码独立建立WebSocket连接
   - 管理多个连接的状态

2. **Line账号自动发现**
   - 扫描本地Line客户端
   - 检测登录的Line账号
   - 定期检查账号变化
   - 上报到服务器

3. **进线监听**
   - 监听Line的好友添加事件
   - 获取新好友的Line User ID
   - 上报到服务器

4. **心跳机制**
   - 每60秒发送心跳包
   - 保持连接活跃

5. **数据上报队列**
   - 异步上报数据
   - 上报失败重试机制
   - 离线缓存机制

---

## ✅ 已确认内容

- [x] 通信方式：WebSocket + HTTP
- [x] 认证方式：激活码认证
- [x] 多激活码支持：一个客户端可登录多个激活码
- [x] 数据归属：通过激活码 + Line ID确定
- [x] 上报内容：Line账号、进线、客户信息、跟进记录
- [x] 在线判断：WebSocket连接 = 在线
- [x] 去重范围：本分组、全局两种
- [x] 进线数据：最小必填 + 可选扩展


