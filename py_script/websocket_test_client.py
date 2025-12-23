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

async def run_test():
    """运行完整测试"""
    print("🚀 WebSocket分组账号状态测试客户端")
    print("=" * 60)
    print(f"WebSocket服务器: {WS_URL}")
    print(f"REST API服务器: {API_BASE_URL}")
    print(f"激活码: {ACTIVATION_CODE}")
    print(f"管理员账号: {ADMIN_USERNAME}")
    print(f"登录接口: {API_BASE_URL}/auth/login")
    print("=" * 60)

    client = WebSocketTestClient(ACTIVATION_CODE)

    # 管理员登录
    print("\n🔐 第一步: 管理员登录")
    if not client.login_admin(ADMIN_USERNAME, ADMIN_PASSWORD):
        print("❌ 管理员登录失败，无法继续测试")
        return

    # 连接WebSocket
    if not await client.connect():
        return

    try:
        await asyncio.sleep(1)  # 等待连接稳定

        # 2. 获取分组列表
        print("\n📂 第二步: 获取分组列表")
        groups = client.get_groups()
        if not groups:
            print("❌ 没有找到分组，请检查管理员token和权限")
            return

        # 选择第一个激活的分组
        active_groups = [g for g in groups if g.get("is_active", True)]
        if not active_groups:
            print("❌ 没有激活的分组")
            return

        selected_group = active_groups[0]
        group_id = selected_group["id"]
        group_code = selected_group["activation_code"]
        group_remark = selected_group.get("remark", "无备注")
        print(f"✅ 选择分组: {group_code} ({group_remark}) (ID: {group_id})")

        # 3. 获取分组的账号
        print("\n👥 第三步: 获取分组账号")
        accounts = client.get_line_accounts(group_id)
        if not accounts:
            print(f"❌ 分组 {group_code} 中没有账号")
            return

        print(f"✅ 找到 {len(accounts)} 个账号:")
        for acc in accounts:
            print(f"  - {acc['display_name']} (ID: {acc['id']}, Line ID: {acc['line_id']})")

        # 4. 账号状态更新测试 - 每个账号间隔30秒上线
        print("\n🔄 第四步: 测试账号状态更新")
        print("每个账号将上线30秒后下线，观察前端同步状态")

        for i, account in enumerate(accounts):
            account_id = account["id"]
            line_id = account["line_id"]
            display_name = account["display_name"]

            print(f"📤 账号 {display_name} 上线")
            await client.update_account_status(line_id, "online")

            print(f"⏳ 账号 {display_name} 保持在线30秒...")
            await asyncio.sleep(30)

            print(f"📥 账号 {display_name} 下线")
            await client.update_account_status(line_id, "offline")

            # 最后一个账号不需要等待
            if i < len(accounts) - 1:
                await asyncio.sleep(2)

        # 5. 测试重复进线
        print("\n📨 第五步: 测试重复进线")

        # 使用第一个账号进行重复进线测试
        if accounts:
            first_account = accounts[0]
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
                await client.send_incoming(account_id, incoming_data)
                await asyncio.sleep(2)  # 等待2秒再发送下一个

        # 6. 删除最后一个账号测试
        if len(accounts) > 0:
            print("\n🗑️ 第六步: 删除账号测试")
            last_account = accounts[-1]
            account_id_to_delete = last_account["id"]
            display_name_to_delete = last_account["display_name"]

            print(f"⚠️ 将删除账号: {display_name_to_delete} (ID: {account_id_to_delete})")
            await asyncio.sleep(2)  # 给用户时间观察

            success = client.delete_line_account(account_id_to_delete)
            if success:
                print("✅ 账号删除成功，请查看前端是否同步更新")
            else:
                print("❌ 账号删除失败")

        # 7. 心跳测试
        print("\n💓 第七步: 心跳测试")
        for i in range(3):
            await client.send_heartbeat()
            await asyncio.sleep(2)

        print("\n✅ 所有测试完成！请查看前端是否实时同步更新了账号状态和删除操作")

    except KeyboardInterrupt:
        print("\n⏹️ 测试被用户中断")
    except Exception as e:
        print(f"\n❌ 测试出错: {e}")
    finally:
        await client.disconnect()

if __name__ == "__main__":
    print("WebSocket 分组账号状态测试")
    print("此脚本将基于当前分组测试账号在线状态更新和删除操作")
    print("运行前请确保:")
    print("1. 后端服务器正在运行 (localhost:8080)")
    print("2. 激活码正确")
    print("3. 前端页面已打开")
    print("-" * 60)
    print("注意事项:")
    print("- 脚本会自动使用管理员账号登录")
    print("- 此脚本会获取真实的分组和账号数据进行测试")
    print("- 账号状态更新每个账号间隔60秒")
    print("- 最后会删除一个账号，请谨慎使用")
    print("-" * 60)

    asyncio.run(run_test())
