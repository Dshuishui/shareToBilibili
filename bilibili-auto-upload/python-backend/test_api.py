# 在python-backend目录创建一个测试文件 test_api.py
import asyncio
from bilibili_api import video_uploader

# 测试bilibili-api的正确用法
async def test_api():
    try:
        # 测试VideoUploaderPage的参数
        print("测试VideoUploaderPage参数...")
        
        # 查看帮助信息
        help(video_uploader.VideoUploaderPage)
        
    except Exception as e:
        print(f"测试错误: {e}")

if __name__ == "__main__":
    asyncio.run(test_api())