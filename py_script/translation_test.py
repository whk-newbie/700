"""
翻译API测试脚本

测试新的中日文翻译接口功能
"""

import requests
import json
import urllib3

# 禁用SSL警告（因为使用自签名证书）
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# API配置 - 通过输入获取
def get_config():
    """获取API配置"""
    print("="*60)
    print("请输入API配置信息")
    print("="*60)
    
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
    
    username = input("请输入用户名: ").strip()
    if not username:
        username = "admin"
        print("使用默认用户名: admin")
    
    password = input("请输入密码: ").strip()
    if not password:
        print("❌ 密码不能为空")
        return None, None, None
    
    return base_url, username, password


def login(base_url, username, password):
    """登录获取token"""
    url = f"{base_url}/auth/login"
    data = {
        "username": username,
        "password": password
    }
    
    print("登录中...")
    try:
        response = requests.post(url, json=data, verify=False)  # 跳过SSL证书验证
        response.raise_for_status()
        result = response.json()
        
        # 检查响应码，成功是1000
        if result.get("code") == 1000 and "data" in result:
            token = result["data"]["token"]
            print(f"✅ 登录成功! Token: {token[:20]}...")
            return token
        else:
            print(f"❌ 登录失败: {result.get('message', '未知错误')}")
            print(f"响应内容: {json.dumps(result, indent=2, ensure_ascii=False)}")
            return None
    except Exception as e:
        print(f"❌ 登录出错: {e}")
        return None


def test_translation(base_url, token, text):
    """测试翻译接口"""
    url = f"{base_url}/llm/translate"
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    data = {
        "text": text
    }
    
    print(f"\n{'='*60}")
    print(f"测试翻译: {text}")
    print(f"{'='*60}")
    
    try:
        response = requests.post(url, json=data, headers=headers, verify=False)  # 跳过SSL证书验证
        response.raise_for_status()
        result = response.json()
        
        # 检查响应码，成功是1000
        if result.get("code") == 1000 and "data" in result:
            data = result["data"]
            print(f"✅ 翻译成功!")
            print(f"  原文: {data['original_text']}")
            print(f"  译文: {data['translated_text']}")
            print(f"  源语言: {data['source_language']}")
            print(f"  目标语言: {data['target_language']}")
            if data.get('tokens_used'):
                print(f"  Token使用: {data['tokens_used']} (prompt: {data.get('prompt_tokens')}, completion: {data.get('completion_tokens')})")
            return True
        else:
            print(f"❌ 翻译失败: {result.get('message', '未知错误')}")
            print(f"响应内容: {json.dumps(result, indent=2, ensure_ascii=False)}")
            return False
    except Exception as e:
        print(f"❌ 请求出错: {e}")
        return False


def test_conversation_reuse(base_url, token):
    """测试对话复用功能"""
    print(f"\n{'='*60}")
    print("测试对话复用功能 - 连续翻译多条文本")
    print(f"{'='*60}")
    
    test_texts = [
        "你好，很高兴认识你",
        "今天天气真好",
        "我喜欢学习日语",
        "こんにちは",
        "ありがとうございます",
        "今日はいい天気ですね"
    ]
    
    for i, text in enumerate(test_texts, 1):
        print(f"\n第 {i} 次翻译:")
        test_translation(base_url, token, text)


def main():
    """主函数"""
    print("="*60)
    print("翻译API测试")
    print("="*60)
    
    # 获取配置
    base_url, username, password = get_config()
    if not base_url or not password:
        print("❌ 配置获取失败，测试中止")
        return
    
    # 登录
    token = login(base_url, username, password)
    if not token:
        print("无法获取token，测试中止")
        return
    
    # 测试中文翻译成日文
    print("\n\n【测试1: 中文翻译成日文】")
    test_translation(base_url, token, "你好，世界！")
    
    # 测试日文翻译成中文
    print("\n\n【测试2: 日文翻译成中文】")
    test_translation(base_url, token, "こんにちは、世界！")
    
    # 测试较长的文本
    print("\n\n【测试3: 较长文本翻译】")
    test_translation(base_url, token, "我们公司是一家专注于人工智能和大数据技术的创新型企业，致力于为客户提供最优质的技术解决方案。")
    
    # 测试对话复用
    print("\n\n【测试4: 对话复用功能】")
    test_conversation_reuse(base_url, token)
    
    print("\n" + "="*60)
    print("所有测试完成!")
    print("="*60)


if __name__ == "__main__":
    main()

