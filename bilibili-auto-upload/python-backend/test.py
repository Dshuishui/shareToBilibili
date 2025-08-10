# test.py - 验证环境
print("🔍 检查Python环境...")

try:
    import bilibili_api
    print("✅ bilibili-api 导入成功!")
    print(f"   版本: {bilibili_api.__version__}")
except ImportError as e:
    print("❌ bilibili-api 导入失败:", e)

try:
    from flask import Flask
    print("✅ Flask 导入成功!")
except ImportError as e:
    print("❌ Flask 导入失败:", e)

try:
    import aiohttp
    print("✅ aiohttp 导入成功!")
except ImportError as e:
    print("❌ aiohttp 导入失败:", e)

print("\n🎉 环境检查完成!")
print("📁 当前工作目录:", __import__('os').getcwd())