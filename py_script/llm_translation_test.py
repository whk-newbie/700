#!/usr/bin/env python3
"""
大模型翻译测试脚本
用于测试LLM功能的翻译能力，通过API调用OpenAI接口

运行方法:
python py_script/llm_translation_test.py

需要修改配置:
- API_BASE_URL: 后端API地址
- ADMIN_USERNAME: 管理员用户名
- ADMIN_PASSWORD: 管理员密码
- TEST_TEXT: 要翻译的文本
"""

import requests
import json
import time
from datetime import datetime

# ===== 配置区域 =====
# 这些变量将在运行时通过输入获取
API_BASE_URL = None
ADMIN_USERNAME = None
ADMIN_PASSWORD = None
TEST_TEXT = "Hello, this is a test message for translation. I hope you can understand it and translate it correctly."  # 要翻译的文本

def get_config():
    """获取API配置"""
    print("="*80)
    print("请输入API配置信息")
    print("="*80)
    
    domain = input("请输入域名: ").strip()
    if not domain:
        print("❌ 域名不能为空")
        return None, None, None
    
    # 处理域名，自动拼接为完整的API地址
    # 如果用户输入的是完整URL，则使用；否则拼接
    if domain.startswith("http://") or domain.startswith("https://"):
        # 如果已经包含 /api/v1，则直接使用
        if "/api/v1" in domain:
            base_url = domain
        else:
            # 移除末尾的斜杠，然后拼接 /api/v1
            base_url = domain.rstrip("/") + "/api/v1"
    else:
        # 只有域名，添加 https:// 和 /api/v1
        base_url = f"https://{domain.rstrip('/')}/api/v1"
    
    username = input("请输入管理员用户名: ").strip()
    if not username:
        username = "admin"
        print("使用默认用户名: admin")
    
    password = input("请输入管理员密码: ").strip()
    if not password:
        print("❌ 密码不能为空")
        return None, None, None
    
    return base_url, username, password

class LLMTranslationTest:
    """大模型翻译测试客户端"""

    def __init__(self):
        self.admin_token = None

    def test_connection(self, api_base_url):
        """测试API服务器连接"""
        try:
            print("🔍 正在测试API服务器连接...")
            response = requests.get(f"{api_base_url.replace('/api/v1', '/health')}", verify=False, timeout=10)
            print(f"✅ 服务器响应: {response.status_code}")
            return True
        except requests.exceptions.RequestException as e:
            print(f"⚠️ 无法访问健康检查端点: {e}")
            # 尝试直接测试登录端点
            try:
                response = requests.options(f"{api_base_url}/auth/login", verify=False, timeout=10)
                print(f"✅ 登录端点可达: {response.status_code}")
                return True
            except:
                print("❌ 无法连接到API服务器")
                return False
        except Exception as e:
            print(f"❌ 连接测试出错: {e}")
            return False

    def login_admin(self, api_base_url, username, password):
        """管理员登录获取token"""
        try:
            login_data = {
                "username": username,
                "password": password
            }
            # 忽略SSL证书验证（处理自签名证书）
            response = requests.post(f"{api_base_url}/auth/login", json=login_data, verify=False)
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

    def check_openai_config(self, api_base_url):
        """检查OpenAI API配置状态"""
        if not self.admin_token:
            print("❌ 未登录，无法检查配置")
            return False

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        try:
            response = requests.get(f"{api_base_url}/admin/llm/openai-key", headers=headers, verify=False)
            response.raise_for_status()
            data = response.json()

            if data.get("code") == 1000 and "data" in data:
                config = data["data"]
                if config.get("has_key"):
                    print("✅ OpenAI API Key已配置")
                    print(f"   更新时间: {config.get('updated_at', '未知')}")
                    return True
                else:
                    print("❌ OpenAI API Key未配置")
                    return False
            else:
                print(f"❌ 获取配置失败: {data.get('message', '未知错误')}")
                return False

        except Exception as e:
            print(f"❌ 检查配置时出错: {e}")
            return False

    def call_llm_translation(self, api_base_url, text, target_language, model="gpt-3.5-turbo"):
        """调用LLM翻译API"""
        if not self.admin_token:
            print("❌ 未登录，无法调用API")
            return None

        headers = {
            "Authorization": f"Bearer {self.admin_token}",
            "Content-Type": "application/json"
        }

        # 构建翻译提示词
        system_prompt = f"You are a professional translator. Translate the following text to {target_language}. Only return the translated text without any explanation or additional content."

        # 构建请求数据
        request_data = {
            "model": model,
            "messages": [
                {
                    "role": "system",
                    "content": system_prompt
                },
                {
                    "role": "user",
                    "content": text
                }
            ],
            "temperature": 0.3,  # 较低的temperature以获得更准确的翻译
            "max_tokens": 2000
        }

        try:
            print(f"🔄 正在翻译为{target_language}...")
            start_time = time.time()

            response = requests.post(
                f"{api_base_url}/llm/proxy/openai",
                json=request_data,
                headers=headers,
                timeout=60,
                verify=False  # 忽略SSL证书验证
            )

            response.raise_for_status()
            data = response.json()

            end_time = time.time()
            duration = end_time - start_time

            # 检查响应是否成功
            if "choices" in data and len(data["choices"]) > 0:
                translated_text = data["choices"][0]["message"]["content"].strip()
                tokens_used = data.get("usage", {}).get("total_tokens", "未知")

                print(f"📊 Token使用: {tokens_used}")
                print(f"📝 翻译结果: {translated_text}")
                print("-" * 80)

                return {
                    "original": text,
                    "translated": translated_text,
                    "target_language": target_language,
                    "model": model,
                    "tokens_used": tokens_used,
                    "duration": duration
                }
            else:
                print(f"❌ API响应格式错误: {data}")
                return None

        except requests.exceptions.Timeout:
            print(f"❌ 请求超时 (60秒)")
            return None
        except requests.exceptions.HTTPError as e:
            print(f"❌ HTTP错误: {e.response.status_code} {e.response.reason}")
            try:
                error_data = e.response.json()
                print(f"   错误详情: {error_data}")
            except:
                print(f"   响应内容: {e.response.text}")
            return None
        except requests.exceptions.RequestException as e:
            print(f"❌ 请求失败: {e}")
            return None
        except Exception as e:
            print(f"❌ 调用API时出错: {e}")
            return None

    def test_translations(self, api_base_url, test_text):
        """测试多种语言翻译"""
        if not self.admin_token:
            print("❌ 未登录，无法进行测试")
            return

        print("🚀 开始大模型翻译测试")
        print("=" * 80)
        print(f"原文: {test_text}")
        print("=" * 80)

        # 要测试的语言列表
        languages = [
            ("英文", "English"),
            ("日文", "Japanese")
        ]

        results = []

        for lang_name, lang_code in languages:
            print(f"\n🌐 正在翻译为{lang_name} ({lang_code})")
            result = self.call_llm_translation(api_base_url, test_text, lang_code)
            if result:
                results.append(result)
            else:
                print(f"❌ {lang_name}翻译失败")

            # 短暂延迟，避免请求过于频繁
            time.sleep(1)

        # 输出总结
        print("\n" + "=" * 80)
        print("📋 测试总结")
        print("=" * 80)

        if results:
            print(f"✅ 成功翻译 {len(results)} 种语言:")
            for result in results:
                lang = result['target_language']
                tokens = result['tokens_used']
                duration = result['duration']
                print(".2f")
        else:
            print("❌ 所有翻译测试都失败了")

        return results

def main():
    """主函数"""
    print("大模型翻译功能测试脚本")
    print("=" * 80)
    
    # 获取配置
    api_base_url, admin_username, admin_password = get_config()
    if not api_base_url or not admin_password:
        print("❌ 配置获取失败，退出测试")
        return
    
    print("=" * 80)
    print(f"管理员账号: {admin_username}")
    print(f"测试文本: {TEST_TEXT}")
    print("=" * 80)

    # 创建测试客户端
    tester = LLMTranslationTest()

    # 测试连接
    print("🔍 正在测试API服务器连接...")
    if not tester.test_connection(api_base_url):
        print("❌ 无法连接到API服务器，请检查网络和服务器状态")
        return

    # 登录
    print("\n🔐 正在登录管理员账号...")
    if not tester.login_admin(api_base_url, admin_username, admin_password):
        print("❌ 登录失败，退出测试")
        return

    # 检查OpenAI配置
    print("\n🔧 检查OpenAI API配置...")
    if not tester.check_openai_config(api_base_url):
        print("❌ OpenAI API未正确配置，请先在管理后台配置API Key")
        return

    # 运行翻译测试
    print("\n🚀 开始翻译测试...")
    results = tester.test_translations(api_base_url, TEST_TEXT)

    # 测试完成
    if results:
        print("✅ 翻译测试完成！")
        print("🎉 大模型功能正常，可以正常调用OpenAI API进行翻译")
    else:
        print("❌ 翻译测试失败，请检查配置和网络连接")
if __name__ == "__main__":
    main()
