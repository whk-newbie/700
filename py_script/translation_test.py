"""
翻译API测试脚本

测试新的中日文翻译接口功能
"""

import requests
import json

# API配置
BASE_URL = "http://localhost:8080/api"
USERNAME = "admin"
PASSWORD = "admin123"


def login():
    """登录获取token"""
    url = f"{BASE_URL}/auth/login"
    data = {
        "username": USERNAME,
        "password": PASSWORD
    }
    
    print(f"登录中... {url}")
    response = requests.post(url, json=data)
    
    if response.status_code == 200:
        result = response.json()
        token = result["data"]["access_token"]
        print(f"登录成功! Token: {token[:20]}...")
        return token
    else:
        print(f"登录失败: {response.status_code}")
        print(response.text)
        return None


def test_translation(token, text):
    """测试翻译接口"""
    url = f"{BASE_URL}/llm/translate"
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
    
    response = requests.post(url, json=data, headers=headers)
    
    if response.status_code == 200:
        result = response.json()
        if result.get("success"):
            data = result["data"]
            print(f"✓ 翻译成功!")
            print(f"  原文: {data['original_text']}")
            print(f"  译文: {data['translated_text']}")
            print(f"  源语言: {data['source_language']}")
            print(f"  目标语言: {data['target_language']}")
            if data.get('tokens_used'):
                print(f"  Token使用: {data['tokens_used']} (prompt: {data.get('prompt_tokens')}, completion: {data.get('completion_tokens')})")
            return True
        else:
            print(f"✗ 翻译失败: {result.get('message')}")
            return False
    else:
        print(f"✗ 请求失败: {response.status_code}")
        print(response.text)
        return False


def test_conversation_reuse(token):
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
        test_translation(token, text)


def main():
    """主函数"""
    print("="*60)
    print("翻译API测试")
    print("="*60)
    
    # 登录
    token = login()
    if not token:
        print("无法获取token，测试中止")
        return
    
    # 测试中文翻译成日文
    print("\n\n【测试1: 中文翻译成日文】")
    test_translation(token, "你好，世界！")
    
    # 测试日文翻译成中文
    print("\n\n【测试2: 日文翻译成中文】")
    test_translation(token, "こんにちは、世界！")
    
    # 测试较长的文本
    print("\n\n【测试3: 较长文本翻译】")
    test_translation(token, "我们公司是一家专注于人工智能和大数据技术的创新型企业，致力于为客户提供最优质的技术解决方案。")
    
    # 测试对话复用
    print("\n\n【测试4: 对话复用功能】")
    test_conversation_reuse(token)
    
    print("\n" + "="*60)
    print("所有测试完成!")
    print("="*60)


if __name__ == "__main__":
    main()

