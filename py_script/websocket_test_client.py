#!/usr/bin/env python3
"""
WebSocket测试客户端
用于测试WebSocket连接、账号同步和实时数据更新功能

运行方法:
python py_script/websocket_test_client.py

需要修改ACTIVATION_CODE为实际的激活码
"""

import asyncio
import websockets
import json
import time
import random
import string
import threading
from datetime import datetime
import requests

# ===== 配置区域 =====
WS_URL = "ws://localhost:8080/api/ws/client"  # WebSocket服务器地址
API_BASE_URL = "http://localhost:8080/api/v1"  # REST API基础地址（包含版本号）
ACTIVATION_CODE = "HG66OP88"  # ⚠️ 需要修改为实际激活码
ADMIN_USERNAME = "admin"  # 管理员用户名
ADMIN_PASSWORD = "admin123"  # 管理员密码

class WebSocketTestClient:
    """WebSocket测试客户端"""

    def __init__(self, activation_code):
        self.activation_code = activation_code
        self.websocket = None
        self.connected = False
        self.admin_token = None

    async def connect(self):
        """连接到WebSocket服务器"""
        try:
            uri = f"{WS_URL}?activation_code={self.activation_code}"
            print(f"🔌 正在连接到: {uri}")
            self.websocket = await websockets.connect(uri)
            self.connected = True
            print("✅ WebSocket连接成功")

            # 启动消息监听任务
            asyncio.create_task(self.listen_messages())
            return True

        except Exception as e:
            print(f"❌ WebSocket连接失败: {e}")
            return False

    async def listen_messages(self):
        """监听服务器消息"""
        try:
            while self.connected:
                message = await self.websocket.recv()
                data = json.loads(message)
                msg_type = data.get('type', 'unknown')

                # 处理不同类型的消息
                if msg_type == 'sync_result':
                    created = data['data']['created_count']
                    updated = data['data']['updated_count']
                    print(f"📊 账号同步完成: 新建 {created} 个, 更新 {updated} 个")

                elif msg_type == 'incoming_received':
                    line_id = data['data']['incoming_line_id']
                    print(f"👋 进线消息已处理: {line_id}")

                elif msg_type == 'customer_sync_received':
                    customer_id = data['data']['customer_id']
                    print(f"👤 客户同步成功: {customer_id}")

                elif msg_type == 'follow_up_sync_received':
                    follow_up_id = data['data']['follow_up_id']
                    print(f"📝 跟进记录同步成功: ID {follow_up_id}")

                elif msg_type == 'account_status_updated':
                    account_id = data['data']['line_account_id']
                    status = data['data']['online_status']
                    print(f"🔄 账号状态已更新: {account_id} -> {status}")

                elif msg_type == 'heartbeat_ack':
                    print("💓 心跳正常")

                else:
                    print(f"📨 收到消息: {msg_type}")

        except websockets.exceptions.ConnectionClosed:
            print("🔌 连接已关闭")
            self.connected = False
        except Exception as e:
            print(f"❌ 监听消息时出错: {e}")

    async def send_heartbeat(self):
        """发送心跳"""
        if not self.connected:
            return

        message = {
            "type": "heartbeat",
            "activation_code": self.activation_code,
            "timestamp": int(time.time())
        }

        try:
            await self.websocket.send(json.dumps(message))
            print("💓 发送心跳")
        except Exception as e:
            print(f"❌ 发送心跳失败: {e}")

    async def sync_accounts(self, accounts):
        """同步账号信息"""
        if not self.connected:
            print("❌ 未连接，无法同步账号")
            return

        message = {
            "type": "sync_line_accounts",
            "activation_code": self.activation_code,
            "data": accounts
        }

        try:
            await self.websocket.send(json.dumps(message))
            print(f"📤 同步 {len(accounts)} 个账号")
        except Exception as e:
            print(f"❌ 同步账号失败: {e}")

    async def send_incoming(self, account_id, incoming_data):
        """发送进线消息"""
        if not self.connected:
            return

        message = {
            "type": "incoming",
            "activation_code": self.activation_code,
            "data": {
                "line_account_id": account_id,
                "incoming_line_id": incoming_data["incoming_line_id"],
                "timestamp": datetime.now().isoformat(),
                "display_name": incoming_data.get("display_name", "测试客户"),
                "avatar_url": incoming_data.get("avatar_url", ""),
                "phone_number": incoming_data.get("phone_number", "")
            }
        }

        try:
            await self.websocket.send(json.dumps(message))
            print(f"📥 进线: {incoming_data['display_name']}")
        except Exception as e:
            print(f"❌ 发送进线消息失败: {e}")

    async def sync_customer(self, customer_data):
        """同步客户"""
        if not self.connected:
            return

        message = {
            "type": "customer_sync",
            "activation_code": self.activation_code,
            "data": customer_data
        }

        try:
            await self.websocket.send(json.dumps(message))
            print(f"👤 同步客户: {customer_data['display_name']}")
        except Exception as e:
            print(f"❌ 同步客户失败: {e}")

    async def update_account_status(self, account_id, status):
        """更新账号状态"""
        if not self.connected:
            return

        message = {
            "type": "account_status_change",
            "activation_code": self.activation_code,
            "data": {
                "line_account_id": account_id,
                "online_status": status,
                "timestamp": datetime.now().isoformat()
            }
        }

        try:
            await self.websocket.send(json.dumps(message))
            print(f"🔄 账号状态: {account_id} -> {status}")
        except Exception as e:
            print(f"❌ 更新账号状态失败: {e}")

    def login_admin(self, username, password):
        """管理员登录获取token"""
        try:
            login_data = {
                "username": username,
                "password": password
            }
            response = requests.post(f"{API_BASE_URL}/auth/login", json=login_data)
            response.raise_for_status()
            data = response.json()

            # 检查响应码，成功是1000
            if data.get("code") == 1000 and "data" in data:
                self.admin_token = data["data"]["token"]
                print(f"✅ 管理员登录成功: {username}")
                return True
            else:
                print(f"❌ 管理员登录失败: {data.get('message', '未知错误')}")
                return False

        except Exception as e:
            print(f"❌ 管理员登录出错: {e}")
            return False

    def login_subaccount(self, activation_code, password=""):
        """子账号登录获取token"""
        try:
            login_data = {
                "activation_code": activation_code,
                "password": password
            }
            response = requests.post(f"{API_BASE_URL}/auth/login-subaccount", json=login_data)
            response.raise_for_status()
            data = response.json()

            # 检查响应码，成功是1000
            if data.get("code") == 1000 and "data" in data:
                self.admin_token = data["data"]["token"]
                print(f"✅ 子账号登录成功: {activation_code}")
                return True
            else:
                print(f"❌ 子账号登录失败: {data.get('message', '未知错误')}")
                return False

        except Exception as e:
            print(f"❌ 子账号登录出错: {e}")
            return False

    async def disconnect(self):
        """断开连接"""
        if self.websocket:
            await self.websocket.close()
            self.connected = False
            print("👋 已断开连接")

    def get_groups(self):
        """获取分组列表"""
        if not self.admin_token:
            print("❌ 未登录管理员")
            return []

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        try:
            response = requests.get(f"{API_BASE_URL}/groups", headers=headers)
            response.raise_for_status()
            data = response.json()
            if data.get("code") == 1000 and "data" in data:
                return data["data"]["list"]  # 分页响应，提取list
            else:
                print(f"❌ 获取分组列表失败: {data.get('message', '响应格式错误')}")
                return []
        except Exception as e:
            print(f"❌ 获取分组列表失败: {e}")
            return []

    def get_line_accounts(self, group_id=None):
        """获取Line账号列表"""
        if not self.admin_token:
            print("❌ 未登录管理员")
            return []

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        params = {}
        if group_id:
            params["group_id"] = group_id

        try:
            response = requests.get(f"{API_BASE_URL}/line-accounts", headers=headers, params=params)
            response.raise_for_status()
            data = response.json()
            if data.get("code") == 1000 and "data" in data:
                return data["data"]["list"]  # 分页响应，提取list
            else:
                print(f"❌ 获取Line账号列表失败: {data.get('message', '响应格式错误')}")
                return []
        except Exception as e:
            print(f"❌ 获取Line账号列表失败: {e}")
            return []

    def create_group(self, group_data):
        """创建分组"""
        if not self.admin_token:
            print("❌ 未登录管理员")
            return None

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        try:
            response = requests.post(f"{API_BASE_URL}/groups", json=group_data, headers=headers)
            response.raise_for_status()
            data = response.json()
            if data.get("code") == 1000 and "data" in data:
                group_info = data["data"]
                print(f"✅ 成功创建分组: {group_info['activation_code']} ({group_info['remark']})")
                return group_info
            else:
                print(f"❌ 创建分组失败: {data.get('message', '响应格式错误')}")
                return None
        except Exception as e:
            print(f"❌ 创建分组失败: {e}")
            return None

    def create_line_account(self, account_data):
        """创建Line账号"""
        if not self.admin_token:
            print("❌ 未登录管理员")
            return None

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        try:
            response = requests.post(f"{API_BASE_URL}/line-accounts", json=account_data, headers=headers)
            response.raise_for_status()
            data = response.json()
            if data.get("code") == 1000 and "data" in data:
                account_info = data["data"]
                print(f"✅ 成功创建账号: {account_info['display_name']} (ID: {account_info['line_id']})")
                return account_info
            else:
                print(f"❌ 创建账号失败: {data.get('message', '响应格式错误')}")
                return None
        except Exception as e:
            print(f"❌ 创建账号失败: {e}")
            return None

    def delete_line_account(self, account_id):
        """删除Line账号"""
        if not self.admin_token:
            print("❌ 未登录管理员")
            return False

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        try:
            response = requests.delete(f"{API_BASE_URL}/line-accounts/{account_id}", headers=headers)
            response.raise_for_status()
            print(f"🗑️ 成功删除账号: {account_id}")
            return True
        except Exception as e:
            print(f"❌ 删除账号失败: {e}")
            return False

def generate_id(length=8):
    """生成随机ID"""
    return ''.join(random.choices(string.ascii_letters + string.digits, k=length))

async def create_accounts_for_group(client, group_id, group_code, account_count=6):
    """为分组创建指定数量的账号"""
    accounts = []
    print(f"📝 为分组 {group_code} 创建 {account_count} 个账号")

    for i in range(account_count):
        line_id = f"test_acc_{group_code}_{generate_id(6)}"
        display_name = f"测试账号_{group_code}_{i+1}"

        account_data = {
            "group_id": group_id,
            "platform_type": "line",
            "line_id": line_id,
            "display_name": display_name,
            "phone_number": f"139{i:04d}0000",
            "account_remark": f"自动创建的测试账号 {i+1}"
        }

        account = client.create_line_account(account_data)
        if account:
            accounts.append(account)
        else:
            print(f"❌ 创建账号失败: {display_name}")

    return accounts

async def run_group_test(client, group_code, accounts, thread_id):
    """运行分组测试"""
    print(f"🧵 线程 {thread_id}: 开始测试分组 {group_code} 的 {len(accounts)} 个账号")

    try:
        # 随机选择账号进行状态更新
        while True:
            # 随机选择一个账号
            account = random.choice(accounts)
            account_id = account["line_id"]
            display_name = account["display_name"]

            # 随机决定上线还是下线
            current_status = account.get("online_status", "offline")
            if current_status == "offline":
                new_status = "online"
                print(f"🧵 线程 {thread_id}: {display_name} 上线")
            else:
                new_status = "offline"
                print(f"🧵 线程 {thread_id}: {display_name} 下线")

            await client.update_account_status(account_id, new_status)

            # 更新本地状态
            account["online_status"] = new_status

            # 随机等待一段时间 (10-30秒)
            wait_time = random.randint(10, 30)
            await asyncio.sleep(wait_time)
    except asyncio.CancelledError:
        print(f"🧵 线程 {thread_id}: 测试任务被取消")
        raise

async def run_test():
    """运行完整测试"""
    print("🚀 WebSocket多分组并发账号状态测试客户端")
    print("=" * 60)
    print(f"WebSocket服务器: {WS_URL}")
    print(f"REST API服务器: {API_BASE_URL}")
    print(f"激活码: {ACTIVATION_CODE}")
    print(f"管理员账号: {ADMIN_USERNAME}")
    print(f"登录接口: {API_BASE_URL}/auth/login")
    print("=" * 60)

    group_clients = {}  # 初始化group_clients变量

    try:
        # 1. 创建临时客户端获取分组列表
        temp_client = WebSocketTestClient(ACTIVATION_CODE)
        if not temp_client.login_admin(ADMIN_USERNAME, ADMIN_PASSWORD):
            print("❌ 管理员登录失败，无法获取分组列表")
            return

        # 2. 获取现有分组列表
        print("\n📂 第二步: 获取现有分组列表")
        all_groups = temp_client.get_groups()
        if not all_groups:
            print("❌ 没有找到分组")
            return

        # 过滤激活的分组
        active_groups = [g for g in all_groups if g.get("is_active", True)]
        if not active_groups:
            print("❌ 没有激活的分组")
            return

        print(f"✅ 找到 {len(active_groups)} 个激活分组:")
        for group in active_groups:
            print(f"  - {group['activation_code']} ({group.get('remark', '无备注')})")

        groups_data = active_groups

        # 3. 为每个分组创建WebSocket连接并检查账号
        print("\n🔌 第三步: 为每个分组创建WebSocket连接并检查账号")
        group_clients = {}

        for group in groups_data:
            group_id = group["id"]
            group_code = group["activation_code"]
            group_client = WebSocketTestClient(group_code)

            # 为分组客户端登录子账号
            print(f"🔐 为分组 {group_code} 登录子账号...")
            if not group_client.login_subaccount(group_code):
                print(f"❌ 分组 {group_code} 子账号登录失败")
                return

            # 连接到分组的WebSocket
            print(f"🔌 连接到分组 {group_code}...")
            if not await group_client.connect():
                print(f"❌ 连接分组 {group_code} 失败")
                return

            await asyncio.sleep(1)  # 等待连接稳定

            # 获取分组的账号
            accounts = group_client.get_line_accounts(group_id)
            if not accounts:
                print(f"⚠️ 分组 {group_code} 没有账号，跳过此分组")
                continue

            print(f"✅ 分组 {group_code} 找到 {len(accounts)} 个账号")

            group_clients[group_code] = {
                "client": group_client,
                "accounts": accounts,
                "group_id": group_id
            }

        if not group_clients:
            print("❌ 没有找到任何有账号的分组")
            return

        print(f"✅ 分组和账号准备完成:")
        total_accounts = 0
        for group_code, data in group_clients.items():
            account_count = len(data['accounts'])
            total_accounts += account_count
            print(f"  - 分组 {group_code}: {account_count} 个账号")

        print(f"📊 总共 {len(group_clients)} 个分组，{total_accounts} 个账号")

        # 4. 启动并发测试
        num_groups = len(group_clients)
        print(f"\n🔄 第四步: 启动并发测试")
        print(f"将为{num_groups}个分组启动并发测试线程")

        # 创建任务列表
        tasks = []

        # 为每个分组创建至少一个测试任务
        client_index = 1
        for group_code, data in group_clients.items():
            group_client = data["client"]
            accounts = data["accounts"]
            task = asyncio.create_task(run_group_test(group_client, group_code, accounts, client_index))
            tasks.append(task)
            client_index += 1

        # 为前几个分组创建额外的测试任务来增加并发度（最多创建到3个任务）
        if num_groups > 0:
            groups_for_extra_tasks = list(group_clients.keys())[:min(3, num_groups)]
            for group_code in groups_for_extra_tasks:
                if len(tasks) >= 3:  # 最多3个并发任务
                    break
                group_client = group_clients[group_code]["client"]
                accounts = group_clients[group_code]["accounts"]
                task = asyncio.create_task(run_group_test(group_client, group_code, accounts, client_index))
                tasks.append(task)
                client_index += 1

        print(f"✅ 启动了 {len(tasks)} 个并发测试任务")

        # 运行所有任务30分钟
        print("⏳ 测试将运行30分钟，请观察前端状态同步")
        try:
            await asyncio.wait_for(asyncio.gather(*tasks, return_exceptions=True), timeout=1800)  # 30分钟
        except asyncio.TimeoutError:
            print("⏹️ 测试时间到，停止所有测试任务")

        # 5. 测试重复进线
        print("\n📨 第五步: 测试重复进线（可选）")

        # 使用第一个分组的第一个账号进行重复进线测试
        first_group_code = list(group_clients.keys())[0]
        first_accounts = group_clients[first_group_code]["accounts"]
        first_client = group_clients[first_group_code]["client"]

        if first_accounts:
            first_account = first_accounts[0]
            account_id = str(first_account["line_id"])
            display_name = first_account["display_name"]
            print(f"使用账号 {display_name} 进行重复进线测试")

            # 测试重复进线的用户ID
            duplicate_users = ["user_L64GlCfl", "user_aRTLd5vS", "user_mYRQ2YFK"]

            for user_id in duplicate_users:
                incoming_data = {
                    "incoming_line_id": user_id,
                    "display_name": f"重复用户_{user_id[-4:]}",  # 使用后4位作为显示名
                    "phone_number": f"13900{user_id[-4:]}"  # 使用后4位作为电话号码
                }

                print(f"📥 发送重复进线: {incoming_data['display_name']} ({user_id})")
                await first_client.send_incoming(account_id, incoming_data)
                await asyncio.sleep(2)  # 等待2秒再发送下一个

        print("\n✅ 所有测试完成！请查看前端是否实时同步更新了账号状态")

    except KeyboardInterrupt:
        print("\n⏹️ 测试被用户中断")
    except Exception as e:
        print(f"\n❌ 测试出错: {e}")
    finally:
        # 断开所有WebSocket连接
        print("\n🔌 断开所有WebSocket连接...")
        disconnect_tasks = []
        for group_code, data in group_clients.items():
            client = data["client"]
            disconnect_tasks.append(client.disconnect())

        if disconnect_tasks:
            await asyncio.gather(*disconnect_tasks, return_exceptions=True)
        print("👋 所有连接已断开")

if __name__ == "__main__":
    print("WebSocket 多分组并发账号状态测试")
    print("此脚本将查询现有分组，使用分组中已有的账号，并启动相应数量的线程并发测试")
    print("运行前请确保:")
    print("1. 后端服务器正在运行 (localhost:8080)")
    print("2. 前端页面已打开")
    print("3. 分组中已有账号")
    print("-" * 60)
    print("测试内容:")
    print("- 查询所有激活的分组")
    print("- 为每个分组进行子账号登录")
    print("- 使用分组中已有的账号")
    print("- 根据分组数量启动相应线程并发测试")
    print("- 支持多个账号同时在线")
    print("- 测试持续30分钟")
    print("-" * 60)
    print("注意事项:")
    print("- 脚本会自动为每个分组进行子账号登录")
    print("- 只使用现有账号，不会创建新账号")
    print("- 如果分组没有账号，会跳过该分组")
    print("- 账号状态随机变化，间隔10-30秒")
    print("- 可通过Ctrl+C提前停止测试")
    print("-" * 60)

    asyncio.run(run_test())
