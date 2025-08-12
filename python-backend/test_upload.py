# test_upload.py - 修正版本
import asyncio
import base64
from pathlib import Path
from bilibili_api import Credential
from bilibili_api.video_uploader import VideoUploader, VideoUploaderPage, VideoMeta
from bilibili_api.utils.picture import Picture  # 添加这个导入

async def test_upload():
    """测试视频上传功能"""
    
    # 1. 设置你的Cookie信息（请替换为真实值）
    credential = Credential(
        sessdata="XXX",  # 替换为真实值
        bili_jct="XX",   # 替换为真实值
        buvid3="XXX",       # 替换为真实值
        dedeuserid="XXX"  # 可选，替换为真实值
    )
    
    # 2. 设置本地视频文件路径（请确保文件存在）
    video_path = "./video.mp4"  # 替换为真实的视频文件路径
    
    # 验证文件存在
    if not Path(video_path).exists():
        print(f"❌ 视频文件不存在: {video_path}")
        print("请修改video_path为实际的视频文件路径")
        return
    
    try:
        print("🎬 开始测试视频上传...")
        print(f"📁 视频文件: {video_path}")
        
        # 3. 创建一个简单的封面图片（1x1像素的PNG）
        # minimal_png_data = base64.b64decode(
        #     'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChAGA4o7+3AAAAABJRU5ErkJggg=='
        # )
        
        # # 创建Picture对象
        # cover_pic = Picture()
        # cover_pic.content = minimal_png_data
        
        print("✅ 封面图片创建成功")
        
        # 4. 创建视频元数据
        meta = VideoMeta(
            tid=1,                    # 分区ID：1=动画
            title="测试视频上传",      # 视频标题
            desc="这是一个测试视频",   # 视频描述
            cover="./小小怪下士.jpg",          # 使用Picture对象
            tags=["测试", "demo"],     # 标签列表
            original=True,            # True=原创
            no_reprint=True,         # True=禁止转载
            open_elec=True           # True=开启充电
        )
        
        print("✅ VideoMeta创建成功")
        
        # 5. 创建视频页面
        page = VideoUploaderPage(
            path=video_path,
            title="测试视频",
            description="测试描述"
        )
        
        print("✅ VideoUploaderPage创建成功")
        
        # 6. 创建上传器
        uploader = VideoUploader(
            pages=[page],
            meta=meta,
            credential=credential
        )
        
        print("✅ VideoUploader创建成功")
        
        # 7. 添加事件监听
        @uploader.on("__ALL__")
        async def upload_event_handler(data):
            print(f"📡 上传事件: {data}")
        
        # 8. 开始上传
        print("🚀 开始上传...")
        result = await uploader.start()
        
        print("🎉 上传完成！")
        print(f"结果: {result}")
        
        return result
        
    except Exception as e:
        import traceback
        print("❌ 上传失败:")
        print(traceback.format_exc())
        return None

def main():
    """主函数"""
    print("=== B站视频上传测试 ===")
    
    # 运行异步测试
    try:
        result = asyncio.run(test_upload())
        if result:
            print("\n🎊 测试成功！")
        else:
            print("\n💥 测试失败！")
    except Exception as e:
        print(f"\n💥 程序错误: {e}")

if __name__ == "__main__":
    main()