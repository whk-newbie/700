#!/usr/bin/env python3
"""
WebSocket性能测试脚本
用于测试WebSocket服务器的并发连接能力和压力承载能力

测试内容：
1. 800个并发连接测试
2. 压力测试（高频消息发送）
3. 内存泄漏检测

运行方法:
python py_script/websocket_perf_test.py

需要修改ACTIVATION_CODE为实际的激活码
"""

import asyncio
import websockets
import requests
import json
import time
import random
import string
import threading
import psutil
import os
from datetime import datetime
import statistics
import matplotlib.pyplot as plt
import numpy as np

# ===== 配置区域 =====
WS_URL = "ws://localhost:8080/api/ws/client"  # WebSocket服务器地址
API_BASE_URL = "http://localhost:8080/api/v1"  # REST API基础地址
ACTIVATION_CODE = "HG66OP88"  # ⚠️ 需要修改为实际激活码
ADMIN_USERNAME = "admin"  # 管理员用户名
ADMIN_PASSWORD = "admin123"  # 管理员密码

# ===== 性能测试配置 =====
CONCURRENT_CONNECTIONS = 800  # 总并发连接数
GROUPS_COUNT = 100           # 分组数量（进一步分散负载）
CONNECTIONS_PER_GROUP = CONCURRENT_CONNECTIONS // GROUPS_COUNT  # 每个分组的连接数
PRESSURE_TEST_DURATION = 300  # 压力测试持续时间（秒）
MEMORY_CHECK_INTERVAL = 5     # 内存检查间隔（秒）
HEARTBEAT_INTERVAL = 30       # 心跳间隔（秒）

class WebSocketPerfTestClient:
    """WebSocket性能测试客户端"""

    def __init__(self, client_id, activation_code):
        self.client_id = client_id
        self.activation_code = activation_code
        self.websocket = None
        self.connected = False
        self.connect_time = None
        self.disconnect_time = None
        self.message_count = 0
        self.error_count = 0
        self.last_heartbeat = time.time()
        self.admin_token = None

    async def connect(self):
        """连接到WebSocket服务器"""
        try:
            uri = f"{WS_URL}?activation_code={self.activation_code}"
            start_time = time.time()
            self.websocket = await asyncio.wait_for(
                websockets.connect(uri),
                timeout=10
            )
            connect_duration = time.time() - start_time
            self.connected = True
            self.connect_time = time.time()

            # 启动消息监听任务
            asyncio.create_task(self.listen_messages())

            return True, connect_duration

        except Exception as e:
            return False, 0

    async def listen_messages(self):
        """监听服务器消息"""
        try:
            while self.connected:
                try:
                    message = await asyncio.wait_for(
                        self.websocket.recv(),
                        timeout=60
                    )
                    self.message_count += 1
                    # 简单处理消息，不解析内容以提高性能
                except asyncio.TimeoutError:
                    # 超时，发送心跳
                    await self.send_heartbeat()
                except websockets.exceptions.ConnectionClosed:
                    break

        except Exception as e:
            self.error_count += 1

    async def send_heartbeat(self):
        """发送心跳"""
        if not self.connected or time.time() - self.last_heartbeat < HEARTBEAT_INTERVAL:
            return

        message = {
            "type": "heartbeat",
            "activation_code": self.activation_code,
            "timestamp": int(time.time()),
            "client_id": self.client_id
        }

        try:
            await self.websocket.send(json.dumps(message))
            self.last_heartbeat = time.time()
        except Exception as e:
            self.error_count += 1

    async def send_test_message(self):
        """发送测试消息"""
        if not self.connected:
            return False

        message = {
            "type": "test_message",
            "activation_code": self.activation_code,
            "data": {
                "client_id": self.client_id,
                "timestamp": int(time.time()),
                "payload": "A" * 100  # 100字节的测试负载
            }
        }

        try:
            await self.websocket.send(json.dumps(message))
            return True
        except Exception as e:
            self.error_count += 1
            return False

    async def disconnect(self):
        """断开连接"""
        if self.websocket:
            await self.websocket.close()
            self.disconnect_time = time.time()
        self.connected = False

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

    def create_group(self, group_data):
        """创建分组"""
        if not self.admin_token:
            print("❌ 未登录管理员")
            return None

        # 移除activation_code，因为它是自动生成的
        create_data = {k: v for k, v in group_data.items() if k != 'activation_code'}
        # 确保包含user_id（管理员的ID）
        create_data['user_id'] = 1  # 假设admin用户ID是1

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        try:
            response = requests.post(f"{API_BASE_URL}/groups", json=create_data, headers=headers)
            response.raise_for_status()
            data = response.json()
            if data.get("code") == 1000 and "data" in data:
                group_info = data["data"]
                print(f"✅ 成功创建分组: {group_info['activation_code']} ({group_info['remark']})")
                return group_info
            else:
                print(f"❌ 创建分组失败: {data.get('message', '响应格式错误')}")
                print(f"响应详情: {data}")
                return None
        except Exception as e:
            print(f"❌ 创建分组失败: {e}")
            return None

    def get_groups(self):
        """获取所有分组列表"""
        if not self.admin_token:
            print("❌ 未登录管理员")
            return []

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        all_groups = []
        page = 1
        page_size = 100  # 每页获取100个分组

        try:
            while True:
                params = {"page": page, "page_size": page_size}
                response = requests.get(f"{API_BASE_URL}/groups", headers=headers, params=params)
                response.raise_for_status()
                data = response.json()

                if data.get("code") == 1000 and "data" in data:
                    groups = data["data"]["list"]
                    all_groups.extend(groups)

                    # 检查是否还有更多页
                    total = data["data"].get("total", 0)
                    if len(all_groups) >= total or len(groups) < page_size:
                        break

                    page += 1
                else:
                    print(f"❌ 获取分组列表失败: {data.get('message', '响应格式错误')}")
                    return []

            return all_groups

        except Exception as e:
            print(f"❌ 获取分组列表失败: {e}")
            return []

    def get_stats(self):
        """获取客户端统计信息"""
        return {
            "client_id": self.client_id,
            "connected": self.connected,
            "connect_time": self.connect_time,
            "disconnect_time": self.disconnect_time,
            "message_count": self.message_count,
            "error_count": self.error_count,
            "uptime": time.time() - self.connect_time if self.connect_time else 0
        }

class MemoryMonitor:
    """内存监控器"""

    def __init__(self, pid):
        self.pid = pid
        self.memory_usage = []
        self.timestamps = []
        self.start_time = time.time()

    def record_memory(self):
        """记录内存使用情况"""
        try:
            process = psutil.Process(self.pid)
            memory_info = process.memory_info()
            memory_mb = memory_info.rss / 1024 / 1024  # 转换为MB

            self.memory_usage.append(memory_mb)
            self.timestamps.append(time.time() - self.start_time)
        except Exception as e:
            print(f"内存监控错误: {e}")

    def get_memory_stats(self):
        """获取内存统计信息"""
        if not self.memory_usage:
            return {}

        return {
            "initial_memory": self.memory_usage[0] if self.memory_usage else 0,
            "final_memory": self.memory_usage[-1] if self.memory_usage else 0,
            "peak_memory": max(self.memory_usage) if self.memory_usage else 0,
            "average_memory": statistics.mean(self.memory_usage) if self.memory_usage else 0,
            "memory_growth": (self.memory_usage[-1] - self.memory_usage[0]) if len(self.memory_usage) > 1 else 0
        }

    def plot_memory_usage(self, filename="memory_usage.png"):
        """绘制内存使用图表"""
        if not self.memory_usage:
            return

        plt.figure(figsize=(12, 6))
        plt.plot(self.timestamps, self.memory_usage, 'b-', linewidth=2, label='Memory Usage (MB)')
        plt.xlabel('Time (seconds)')
        plt.ylabel('Memory Usage (MB)')
        plt.title('WebSocket Server Memory Usage During Performance Test')
        plt.grid(True, alpha=0.3)
        plt.legend()

        # 添加统计信息
        stats = self.get_memory_stats()
        info_text = ".1f"".1f"".1f"".1f"".1f"".1f"f"""
        Memory Stats:
        Initial: {stats['initial_memory']:.1f} MB
        Final: {stats['final_memory']:.1f} MB
        Peak: {stats['peak_memory']:.1f} MB
        Growth: {stats['memory_growth']:+.1f} MB
        """

        plt.figtext(0.02, 0.98, info_text, fontsize=10,
                   verticalalignment='top', fontfamily='monospace',
                   bbox=dict(boxstyle='round', facecolor='wheat', alpha=0.8))

        plt.tight_layout()
        plt.savefig(filename, dpi=150, bbox_inches='tight')
        plt.close()
        print(f"内存使用图表已保存到: {filename}")

class PerformanceTest:
    """性能测试主类"""

    def __init__(self):
        self.clients = []
        self.memory_monitor = None
        self.test_results = {
            "connection_test": {},
            "pressure_test": {},
            "memory_test": {}
        }

    def generate_client_id(self, index):
        """生成客户端ID"""
        return f"perf_test_client_{index:04d}"

    async def create_test_groups(self):
        """创建或获取测试分组"""
        print("检查测试分组...")

        # 登录管理员
        temp_client = WebSocketPerfTestClient("admin_client", ACTIVATION_CODE)
        if not temp_client.login_admin(ADMIN_USERNAME, ADMIN_PASSWORD):
            print("❌ 管理员登录失败")
            return None

        # 获取现有分组
        all_groups = temp_client.get_groups()
        if not all_groups:
            print("❌ 获取分组列表失败")
            return None

        # 筛选性能测试分组（通过remark识别）
        perf_groups = [g for g in all_groups if g.get("remark", "").startswith("性能测试分组")]
        print(f"找到 {len(perf_groups)} 个现有性能测试分组")

        # 调试信息：显示所有分组的remark
        print("所有分组的remark:")
        for g in all_groups:
            remark = g.get("remark", "")
            if remark:
                print(f"  - {g.get('activation_code', '')}: {remark}")
            else:
                print(f"  - {g.get('activation_code', '')}: (无备注)")

        # 如果现有分组足够，直接使用
        if len(perf_groups) >= GROUPS_COUNT:
            test_groups = perf_groups[:GROUPS_COUNT]
            print(f"✅ 重用 {len(test_groups)} 个现有性能测试分组")
            return test_groups

        # 需要创建更多分组
        existing_codes = {g["activation_code"] for g in perf_groups}
        groups_to_create = GROUPS_COUNT - len(perf_groups)

        print(f"需要创建 {groups_to_create} 个新分组...")

        for i in range(200):  # 尝试足够多的编号
            if len(perf_groups) >= GROUPS_COUNT:
                break

            group_data = {
                "remark": f"性能测试分组{i+1}",
                "is_active": True
            }

            group = temp_client.create_group(group_data)
            if group:
                perf_groups.append(group)
                print(f"✅ 创建分组: {group['activation_code']} (性能测试分组{len(perf_groups)})")

        if len(perf_groups) < GROUPS_COUNT:
            print(f"⚠️  仅获得 {len(perf_groups)} 个分组，期望 {GROUPS_COUNT} 个")
            return perf_groups if perf_groups else None

        print(f"✅ 准备了 {len(perf_groups)} 个性能测试分组")
        return perf_groups

    async def test_800_connections(self, test_groups):
        """测试800个并发连接"""
        print(f"\n{'='*60}")
        print(f"开始{CONCURRENT_CONNECTIONS}个并发连接测试")
        print(f"分组数量: {len(test_groups)}, 每组连接数: {CONNECTIONS_PER_GROUP}")
        print(f"{'='*60}")

        start_time = time.time()
        connect_results = []

        # 创建客户端，为每个分组分配连接
        self.clients = []
        client_index = 0

        for group in test_groups:
            group_code = group["activation_code"]
            for i in range(CONNECTIONS_PER_GROUP):
                client = WebSocketPerfTestClient(
                    self.generate_client_id(client_index),
                    group_code
                )
                self.clients.append(client)
                client_index += 1

        print(f"创建了 {len(self.clients)} 个测试客户端")

        # 并发连接
        print("开始并发连接...")
        connection_tasks = []

        for client in self.clients:
            task = asyncio.create_task(client.connect())
            connection_tasks.append(task)

        # 分批执行连接，避免一次性创建太多协程
        batch_size = 100
        successful_connections = 0
        failed_connections = 0
        connect_durations = []

        for i in range(0, len(connection_tasks), batch_size):
            batch = connection_tasks[i:i+batch_size]
            print(f"执行第 {i//batch_size + 1} 批连接 ({len(batch)} 个)...")

            batch_results = await asyncio.gather(*batch, return_exceptions=True)

            for result in batch_results:
                if isinstance(result, Exception):
                    failed_connections += 1
                else:
                    success, duration = result
                    if success:
                        successful_connections += 1
                        connect_durations.append(duration)
                    else:
                        failed_connections += 1

            # 批次间稍作延迟
            await asyncio.sleep(0.1)

        connection_time = time.time() - start_time

        # 统计连接结果
        self.test_results["connection_test"] = {
            "total_clients": len(self.clients),
            "successful_connections": successful_connections,
            "failed_connections": failed_connections,
            "success_rate": successful_connections / len(self.clients) * 100,
            "total_connection_time": connection_time,
            "average_connect_time": statistics.mean(connect_durations) if connect_durations else 0,
            "max_connect_time": max(connect_durations) if connect_durations else 0,
            "min_connect_time": min(connect_durations) if connect_durations else 0
        }

        print("\n连接测试结果:")
        print(f"  总客户端数: {len(self.clients)}")
        print(f"  成功连接: {successful_connections}")
        print(f"  失败连接: {failed_connections}")
        print(".2f")
        print(".3f")
        print(".3f")
        print(".3f")
        print(".3f")
        # 等待连接稳定
        print("等待连接稳定...")
        await asyncio.sleep(10)

        # 检查连接状态
        active_connections = sum(1 for client in self.clients if client.connected)
        print(f"连接稳定后活跃连接数: {active_connections}")

        return successful_connections == len(self.clients)

    async def pressure_test(self):
        """压力测试"""
        print(f"\n{'='*60}")
        print("开始压力测试")
        print(f"{'='*60}")

        if not self.clients:
            print("没有可用的客户端连接")
            return False

        # 只使用成功连接的客户端
        active_clients = [client for client in self.clients if client.connected]
        print(f"使用 {len(active_clients)} 个活跃客户端进行压力测试")

        if len(active_clients) == 0:
            print("没有活跃的客户端连接")
            return False

        start_time = time.time()
        test_duration = PRESSURE_TEST_DURATION
        message_interval = 0.01  # 每10ms发送一条消息

        print(f"压力测试持续时间: {test_duration} 秒")
        print(f"消息发送间隔: {message_interval} 秒")
        print(".0f")
        # 压力测试任务
        total_messages_sent = 0
        total_messages_failed = 0

        async def send_pressure_messages(client, client_index):
            """为单个客户端发送压力消息"""
            nonlocal total_messages_sent, total_messages_failed
            messages_sent = 0
            messages_failed = 0

            while time.time() - start_time < test_duration:
                if await client.send_test_message():
                    messages_sent += 1
                else:
                    messages_failed += 1

                await asyncio.sleep(message_interval)

            client_stats = client.get_stats()
            print(f"客户端 {client_index}: 发送 {messages_sent} 成功, {messages_failed} 失败, 收到 {client_stats['message_count']} 消息")

            total_messages_sent += messages_sent
            total_messages_failed += messages_failed

        # 创建压力测试任务
        pressure_tasks = []
        for i, client in enumerate(active_clients):
            task = asyncio.create_task(send_pressure_messages(client, i))
            pressure_tasks.append(task)

        # 等待压力测试完成
        await asyncio.gather(*pressure_tasks, return_exceptions=True)

        actual_duration = time.time() - start_time

        # 计算压力测试结果
        total_messages = total_messages_sent + total_messages_failed
        messages_per_second = total_messages / actual_duration if actual_duration > 0 else 0
        success_rate = total_messages_sent / total_messages * 100 if total_messages > 0 else 0

        self.test_results["pressure_test"] = {
            "test_duration": actual_duration,
            "total_messages_sent": total_messages_sent,
            "total_messages_failed": total_messages_failed,
            "messages_per_second": messages_per_second,
            "success_rate": success_rate,
            "active_clients": len(active_clients)
        }

        print("\n压力测试结果:")
        print(".1f")
        print(f"  总消息数: {total_messages}")
        print(f"  成功发送: {total_messages_sent}")
        print(f"  发送失败: {total_messages_failed}")
        print(".2f")
        print(".2f")
        return True

    async def memory_leak_test(self):
        """内存泄漏检测测试"""
        print(f"\n{'='*60}")
        print("开始内存泄漏检测")
        print(f"{'='*60}")

        # 查找服务器进程 (假设是server.exe或server)
        server_pid = None
        for proc in psutil.process_iter(['pid', 'name', 'cmdline']):
            try:
                if 'server' in proc.info['name'].lower() or \
                   (proc.info['cmdline'] and any('server' in str(cmd).lower() for cmd in proc.info['cmdline'])):
                    server_pid = proc.info['pid']
                    break
            except (psutil.NoSuchProcess, psutil.AccessDenied):
                continue

        if not server_pid:
            print("未找到服务器进程，跳过内存泄漏检测")
            self.test_results["memory_test"] = {"error": "服务器进程未找到"}
            return False

        print(f"找到服务器进程 PID: {server_pid}")

        # 初始化内存监控
        self.memory_monitor = MemoryMonitor(server_pid)

        # 记录基准内存
        print("记录基准内存使用...")
        for _ in range(5):
            self.memory_monitor.record_memory()
            await asyncio.sleep(1)

        # 在测试期间持续监控内存
        print(f"开始内存监控，监控间隔: {MEMORY_CHECK_INTERVAL} 秒")

        monitor_start = time.time()
        while time.time() - monitor_start < PRESSURE_TEST_DURATION + 60:  # 多监控60秒
            self.memory_monitor.record_memory()
            await asyncio.sleep(MEMORY_CHECK_INTERVAL)

        # 生成内存报告
        memory_stats = self.memory_monitor.get_memory_stats()
        self.test_results["memory_test"] = memory_stats

        print("\n内存泄漏检测结果:")
        if memory_stats:
            print(".1f")
            print(".1f")
            print(".1f")
            print(".1f")
            print(".1f")
            # 判断是否存在内存泄漏
            growth_rate = abs(memory_stats['memory_growth']) / memory_stats['initial_memory'] * 100
            if growth_rate > 20:  # 增长超过20%认为有内存泄漏风险
                print(".1f")
                print("⚠️  检测到可能的内存泄漏！")
            else:
                print(".1f")
                print("✅ 内存使用正常")

            # 生成内存使用图表
            self.memory_monitor.plot_memory_usage("websocket_perf_memory.png")

        return True

    async def run_full_test(self):
        """运行完整性能测试"""
        print("🚀 WebSocket性能测试开始")
        print("=" * 80)
        print(f"测试配置:")
        print(f"  总并发连接数: {CONCURRENT_CONNECTIONS}")
        print(f"  测试分组数量: {GROUPS_COUNT}")
        print(f"  每组连接数: {CONNECTIONS_PER_GROUP}")
        print(f"  压力测试时长: {PRESSURE_TEST_DURATION} 秒")
        print(f"  内存检查间隔: {MEMORY_CHECK_INTERVAL} 秒")
        print(f"  WebSocket服务器: {WS_URL}")
        print("=" * 80)

        try:
            # 1. 创建测试分组
            print("第一步: 创建测试分组")
            test_groups = await self.create_test_groups()
            if not test_groups:
                print("❌ 创建测试分组失败，终止测试")
                return

            # 2. 800个并发连接测试
            connection_success = await self.test_800_connections(test_groups)
            if not connection_success:
                print("❌ 连接测试失败，终止测试")
                return

            # 2. 内存泄漏检测（并发进行）
            memory_task = asyncio.create_task(self.memory_leak_test())

            # 3. 压力测试
            pressure_success = await self.pressure_test()

            # 等待内存检测完成
            await memory_task

            # 生成完整测试报告
            self.generate_report()

        except Exception as e:
            print(f"❌ 测试过程中出错: {e}")
        finally:
            # 断开所有连接
            await self.cleanup()

    async def cleanup(self):
        """清理资源"""
        print("\n🔄 清理测试资源...")

        if self.clients:
            disconnect_tasks = []
            for client in self.clients:
                if client.connected:
                    disconnect_tasks.append(client.disconnect())

            if disconnect_tasks:
                await asyncio.gather(*disconnect_tasks, return_exceptions=True)

        print("✅ 测试资源清理完成")

    def generate_report(self):
        """生成测试报告"""
        print(f"\n{'='*80}")
        print("📊 WebSocket性能测试报告")
        print(f"{'='*80}")

        # 连接测试报告
        conn = self.test_results.get("connection_test", {})
        if conn:
            print("\n🔌 连接测试结果:")
            print(f"  总客户端数: {conn.get('total_clients', 0)}")
            print(f"  成功连接: {conn.get('successful_connections', 0)}")
            print(f"  失败连接: {conn.get('failed_connections', 0)}")
            print(".2f")
            print(".3f")
            print(".3f")
        # 压力测试报告
        pressure = self.test_results.get("pressure_test", {})
        if pressure:
            print("\n⚡ 压力测试结果:")
            print(".1f")
            print(f"  总消息数: {pressure.get('total_messages_sent', 0) + pressure.get('total_messages_failed', 0)}")
            print(f"  成功发送: {pressure.get('total_messages_sent', 0)}")
            print(f"  发送失败: {pressure.get('total_messages_failed', 0)}")
            print(".2f")
            print(".2f")
        # 内存测试报告
        memory = self.test_results.get("memory_test", {})
        if memory and "error" not in memory:
            print("\n🧠 内存测试结果:")
            print(".1f")
            print(".1f")
            print(".1f")
            print(".1f")
            print(".1f")
            growth_rate = abs(memory['memory_growth']) / memory['initial_memory'] * 100
            print(".1f")
            if growth_rate > 20:
                print("  ⚠️  状态: 检测到可能的内存泄漏！")
            else:
                print("  ✅ 状态: 内存使用正常")

        print("\n📈 性能评估:")
        # 综合评估
        success_rate = conn.get('success_rate', 0)
        pressure_success_rate = pressure.get('success_rate', 0)
        memory_growth = memory.get('memory_growth', 0) if "error" not in memory else 0

        if success_rate >= 95 and pressure_success_rate >= 95 and abs(memory_growth) < 50:
            print("  🟢 总体评价: 优秀 - 系统表现稳定，性能良好")
        elif success_rate >= 90 and pressure_success_rate >= 90:
            print("  🟡 总体评价: 良好 - 系统基本稳定，建议优化内存使用")
        else:
            print("  🔴 总体评价: 需要改进 - 存在性能或稳定性问题")

        # 保存详细报告到文件
        self.save_detailed_report()

        print("\n✅ 性能测试完成！")        
        print("详细报告已保存到: websocket_perf_report.txt")
        if memory and "error" not in memory:
            print("内存使用图表已保存到: websocket_perf_memory.png")

    def save_detailed_report(self):
        """保存详细报告到文件"""
        try:
            with open("websocket_perf_report.txt", "w", encoding="utf-8") as f:
                f.write("WebSocket性能测试详细报告\n")
                f.write("=" * 50 + "\n")
                f.write(f"测试时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
                f.write(f"服务器地址: {WS_URL}\n")
                f.write(f"激活码: {ACTIVATION_CODE}\n\n")

                # 写入测试配置
                f.write("测试配置:\n")
                f.write(f"  并发连接数: {CONCURRENT_CONNECTIONS}\n")
                f.write(f"  压力测试时长: {PRESSURE_TEST_DURATION} 秒\n")
                f.write(f"  内存检查间隔: {MEMORY_CHECK_INTERVAL} 秒\n\n")

                # 写入各测试结果
                for test_name, results in self.test_results.items():
                    f.write(f"{test_name.upper()} 测试结果:\n")
                    for key, value in results.items():
                        if isinstance(value, float):
                            f.write(".3f")                        
                        else:
                            f.write(f"  {key}: {value}\n")
                    f.write("\n")

                f.write("测试完成\n")

            print("详细报告已保存到 websocket_perf_report.txt")

        except Exception as e:
            print(f"保存报告失败: {e}")

async def main():
    """主函数"""
    print("WebSocket性能测试工具")
    print("此工具将执行以下测试:")
    print("1. 800个并发连接测试")
    print("2. 压力测试（高频消息发送）")
    print("3. 内存泄漏检测")
    print("-" * 50)
    print("注意事项:")
    print("- 确保WebSocket服务器正在运行")
    print("- 测试将持续约10分钟")
    print("- 测试期间会产生大量日志")
    print("- 建议在测试环境执行")
    print("-" * 50)

    # 确认开始测试
    try:
        response = input("是否开始性能测试？(y/N): ").strip().lower()
        if response != 'y':
            print("测试已取消")
            return
    except KeyboardInterrupt:
        print("\n测试已取消")
        return

    # 创建并运行性能测试
    tester = PerformanceTest()
    await tester.run_full_test()

if __name__ == "__main__":
    asyncio.run(main())
