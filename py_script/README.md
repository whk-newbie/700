# WebSocket测试脚本

这个文件夹包含用于测试WebSocket实时数据推送功能的Python脚本。

## 文件说明

- `websocket_test_client.py` - WebSocket测试客户端脚本
- `.venv/` - uv创建的虚拟环境
- `README.md` - 使用说明

## 环境设置

### 1. 激活虚拟环境

```bash
# Windows
py_script\.venv\Scripts\activate

# Linux/Mac
source py_script/.venv/bin/activate
```

### 2. 或者直接使用uv运行

```bash
uv run python py_script/websocket_test_client.py
```

## 传统方式（如果没有uv）

确保已安装Python 3.7+ 和 websockets库：

```bash
pip install websockets
```

### 2. 修改配置

打开 `websocket_test_client.py`，修改激活码：

```python
ACTIVATION_CODE = "你的激活码"  # 修改为实际激活码
```

### 3. 运行测试

#### 方式一：激活虚拟环境后运行
```bash
# 激活虚拟环境
py_script\.venv\Scripts\activate  # Windows
source py_script/.venv/bin/activate  # Linux/Mac

# 运行脚本
python websocket_test_client.py
```

#### 方式二：使用uv直接运行
```bash
uv run python websocket_test_client.py
```

#### 方式三：指定Python路径运行
```bash
py_script\.venv\Scripts\python.exe websocket_test_client.py  # Windows
py_script/.venv/bin/python websocket_test_client.py  # Linux/Mac
```

## 测试内容

脚本会自动执行以下测试：

1. **WebSocket连接** - 连接到服务器并验证激活码
2. **账号同步** - 创建测试Line账号
3. **进线模拟** - 发送客户进线消息
4. **客户同步** - 同步客户详细信息
5. **状态更新** - 测试账号在线/离线状态变化
6. **心跳测试** - 发送心跳并接收响应

## 观察结果

运行脚本时，在前端看板页面观察：

- ✅ 账号列表实时更新
- ✅ 进线统计增加
- ✅ 客户列表更新
- ✅ 账号状态变化
- ✅ 分组统计刷新

## 注意事项

- 确保后端服务器正在运行
- 激活码必须正确
- 前端页面需要打开才能看到实时更新效果
