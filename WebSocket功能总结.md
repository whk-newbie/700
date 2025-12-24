# WebSocket功能总结

**最后更新时间：2025-12-23**

## 概述

本系统的WebSocket模块提供了实时通信功能，支持两种客户端类型（Windows客户端和前端看板）的连接，并实现了多种消息类型的处理和广播机制。

## 客户端类型

| 客户端类型 | 描述 | 主要功能 |
|-----------|------|----------|
| Windows客户端 | Windows桌面应用程序 | 发送数据同步消息（账号、客户、进线等） |
| 前端看板 | Web前端界面 | 接收实时数据更新和统计信息 |

## 消息类型

### 客户端发送消息

| 消息类型 | 发送者 | 描述 | 数据结构 |
|---------|--------|------|----------|
| `heartbeat` | 所有客户端 | 心跳检测 | `{type: "heartbeat", activation_code: "...", timestamp: 1234567890}` |
| `sync_line_accounts` | Windows客户端 | 同步Line账号信息 | `{type: "sync_line_accounts", activation_code: "...", data: [...]}` |
| `incoming` | Windows客户端 | 进线消息通知 | `{type: "incoming", activation_code: "...", data: {...}}` |
| `customer_sync` | Windows客户端 | 客户数据同步 | `{type: "customer_sync", activation_code: "...", data: {...}}` |
| `follow_up_sync` | Windows客户端 | 跟进记录同步 | `{type: "follow_up_sync", activation_code: "...", data: {...}}` |
| `account_status_change` | Windows客户端 | 账号状态变化 | `{type: "account_status_change", activation_code: "...", data: {...}}` |

### 服务器推送消息

| 消息类型 | 接收者 | 描述 | 数据结构 |
|---------|--------|------|----------|
| `heartbeat_ack` | 所有客户端 | 心跳响应确认 | `{type: "heartbeat_ack", activation_code: "...", timestamp: 1234567890, data: {status: "ok", message: "心跳正常"}}` |
| `auth_success` | Windows客户端 | 认证成功 | `{type: "auth_success", data: {group_id: 1, activation_code: "...", message: "认证成功，请同步Line账号列表"}}` |
| `connected` | 前端看板 | 连接成功 | `{type: "connected", data: {user_id: 1, group_id: 1, message: "WebSocket连接成功"}}` |
| `sync_result` | Windows客户端 | 同步结果反馈 | `{type: "sync_result", data: {success: true, created_count: 1, updated_count: 2, accounts: [...]}}` |
| `incoming_received` | Windows客户端 | 进线消息已接收 | `{type: "incoming_received", data: {line_account_id: "...", incoming_line_id: "...", status: "processed"}}` |
| `customer_sync_received` | Windows客户端 | 客户同步已接收 | `{type: "customer_sync_received", data: {customer_id: "...", customer_db_id: 1, status: "processed"}}` |
| `follow_up_sync_received` | Windows客户端 | 跟进记录同步已接收 | `{type: "follow_up_sync_received", data: {customer_id: "...", follow_up_id: 1, status: "processed"}}` |
| `account_status_updated` | Windows客户端 | 账号状态更新确认 | `{type: "account_status_updated", data: {line_account_id: "...", online_status: "online", status: "updated"}}` |
| `account_status_change` | 前端看板 | 账号状态变化广播 | `{type: "account_status_change", data: {line_account_id: "...", online_status: "online", group_id: 1, timestamp: 1234567890}}` |
| `group_stats_update` | 前端看板 | 分组统计更新 | `{type: "group_stats_update", data: {activation_code: "HG66OP88", total_accounts: 10, online_accounts: 8, total_incoming: 100, today_incoming: 50, duplicate_incoming: 20, today_duplicate: 5, timestamp: 1234567890}}` |
| `account_stats_update` | 前端看板 | 账号统计更新 | `{type: "account_stats_update", data: {line_id: "line_account_123", total_incoming: 50, today_incoming: 25, duplicate_incoming: 10, today_duplicate: 3, timestamp: 1234567890}}` |
| `error` | 所有客户端 | 错误消息 | `{type: "error", error: "错误描述信息"}` |

## 核心功能模块

### 1. 连接管理 (Manager)
- **连接池管理**: 分别管理Windows客户端和前端看板连接
- **心跳检测**: 自动检测客户端连接状态，超时自动断开
- **分组广播**: 支持按分组ID广播消息到前端看板
- **全局广播**: 广播消息到所有前端看板

### 2. 消息处理 (MessageHandler)
- **消息路由**: 根据消息类型分发到对应处理函数
- **数据验证**: 验证激活码和用户权限
- **业务处理**: 调用相应的业务服务处理数据
- **响应反馈**: 向客户端发送处理结果

### 3. 广播中心 (Hub)
- **进线更新广播**: 实时推送新进线客户信息
- **账号状态广播**: 实时推送Line账号在线状态变化
- **统计更新广播**: 实时推送分组统计数据更新

### 4. 前端集成
- **自动重连**: 连接断开后自动重连机制
- **心跳保活**: 定期发送心跳保持连接
- **消息处理器**: 支持注册多个消息处理器
- **状态管理**: 集成Vuex/Pinia进行状态管理

## 技术特性

### 连接特性
- **双向通信**: 支持客户端到服务器和服务器到客户端的消息
- **并发处理**: 支持多个客户端同时连接
- **连接超时**: 自动检测和处理超时连接
- **优雅关闭**: 支持正常关闭和异常关闭处理

### 安全性
- **认证验证**: Windows客户端通过激活码验证，前端看板通过JWT token验证
- **权限控制**: 按用户和分组进行权限控制
- **数据隔离**: 不同分组的数据相互隔离

### 性能优化
- **异步处理**: 使用goroutine异步处理消息
- **缓冲通道**: 使用缓冲通道避免阻塞
- **批量发送**: 支持批量发送消息优化性能
- **连接复用**: WebSocket连接复用减少开销

### 可靠性
- **消息确认**: 重要消息需要客户端确认
- **错误处理**: 完善的错误处理和日志记录
- **数据校验**: 对接收的数据进行完整性校验
- **状态同步**: 实时同步各种业务状态

## 使用场景

1. **实时监控**: 前端看板实时显示在线账号数、进线统计等
2. **即时通知**: 新客户进线时立即通知相关人员
3. **状态同步**: Line账号在线状态变化实时反映到界面
4. **数据同步**: Windows客户端自动同步客户数据到服务器
5. **统计更新**: 分组统计数据实时更新，无需手动刷新
