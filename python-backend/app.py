import asyncio
import json
import os
import tempfile
import shutil
import base64
import time
from pathlib import Path
from flask import Flask, request, jsonify
from flask_cors import CORS
from bilibili_api import Credential
from bilibili_api.video_uploader import VideoUploader, VideoUploaderPage, VideoMeta
from bilibili_api.utils.picture import Picture

app = Flask(__name__)
CORS(app, origins=[
    "http://localhost:5173",  # XBuilder开发服务器
    "http://localhost:3000",  # Node.js服务
    "http://127.0.0.1:5173",  # 备用地址
], supports_credentials=True)

# 全局变量存储用户凭证
user_credential = None

# 确保uploads目录存在
UPLOAD_DIR = Path(__file__).parent / "uploads"
UPLOAD_DIR.mkdir(exist_ok=True)

def create_default_cover():
    """创建默认封面图片 - 参考test_upload.py"""
    # 1x1像素的PNG图片 base64 数据
    minimal_png_data = base64.b64decode(
        'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChAGA4o7+3AAAAABJRU5ErkJggg=='
    )
    
    picture = Picture()
    picture.content = minimal_png_data
    return picture

@app.route('/health', methods=['GET'])
def health_check():
    """健康检查接口"""
    return jsonify({
        "status": "ok",
        "message": "Python后端服务运行正常",
        "service": "bilibili-api-backend",
        "upload_dir": str(UPLOAD_DIR),
        "logged_in": user_credential is not None
    })

@app.route('/login', methods=['POST'])
def login():
    """处理用户登录 - 接收Cookie信息"""
    try:
        data = request.json
        cookies = data.get('cookies', [])
        
        if not cookies:
            return jsonify({
                "success": False,
                "message": "未收到Cookie信息"
            })
        
        # 从Cookie数组中提取需要的值
        cookie_dict = {}
        for cookie in cookies:
            cookie_dict[cookie['name']] = cookie['value']
        
        # 检查必要的Cookie
        required_cookies = ['SESSDATA', 'bili_jct', 'buvid3']
        missing_cookies = [name for name in required_cookies if name not in cookie_dict]
        
        if missing_cookies:
            return jsonify({
                "success": False,
                "message": f"缺少必要的Cookie: {', '.join(missing_cookies)}"
            })
        
        # 创建凭证对象 - 完全按照test_upload.py的方式
        global user_credential
        user_credential = Credential(
            sessdata=cookie_dict['SESSDATA'],
            bili_jct=cookie_dict['bili_jct'],
            buvid3=cookie_dict['buvid3'],
            dedeuserid=cookie_dict.get('DedeUserID')  # 可选，但推荐
        )
        
        print(f"✅ 用户凭证创建成功")
        print(f"   SESSDATA: {cookie_dict['SESSDATA'][:20]}...")
        print(f"   bili_jct: {cookie_dict['bili_jct'][:10]}...")
        print(f"   buvid3: {cookie_dict['buvid3'][:20]}...")
        if 'DedeUserID' in cookie_dict:
            print(f"   DedeUserID: {cookie_dict['DedeUserID']}")
        
        return jsonify({
            "success": True,
            "message": "登录凭证已保存，可以开始上传视频",
            "cookieCount": len(cookies)
        })
        
    except Exception as e:
        print(f"❌ 登录处理失败: {e}")
        return jsonify({
            "success": False,
            "message": f"登录处理失败: {str(e)}"
        })

@app.route('/check-login', methods=['GET'])
def check_login():
    """检查登录状态"""
    global user_credential
    return jsonify({
        "isLoggedIn": user_credential is not None,
        "message": "用户已登录" if user_credential else "用户未登录"
    })

@app.route('/upload-video', methods=['POST'])
def upload_video():
    """上传视频接口 - 修正版本"""
    try:
        global user_credential
        
        if not user_credential:
            return jsonify({
                "success": False,
                "message": "请先登录B站账号"
            })
        
        # 获取表单数据
        video_file = request.files.get('video')
        title = request.form.get('title', '').strip()
        description = request.form.get('description', '').strip()
        tags = request.form.get('tags', '').strip()
        category_id = request.form.get('category', '')
        
        # 验证必要参数
        if not video_file:
            return jsonify({
                "success": False,
                "message": "未选择视频文件"
            })
        
        if not title:
            return jsonify({
                "success": False,
                "message": "视频标题不能为空"
            })
        
        if not category_id:
            return jsonify({
                "success": False,
                "message": "请选择视频分区"
            })
        
        try:
            category_id = int(category_id)
        except ValueError:
            return jsonify({
                "success": False,
                "message": "无效的分区ID"
            })
        
        # 保存视频文件到本地
        try:
            timestamp = int(time.time())
            safe_filename = f"{timestamp}_{video_file.filename}"
            local_video_path = UPLOAD_DIR / safe_filename
            
            print(f"📁 保存视频文件到: {local_video_path}")
            video_file.save(str(local_video_path))
            
            if not local_video_path.exists():
                raise FileNotFoundError(f"文件保存失败: {local_video_path}")
            
            file_size = local_video_path.stat().st_size
            print(f"✅ 文件保存成功，大小: {file_size / 1024 / 1024:.2f} MB")
            
        except Exception as e:
            print(f"❌ 文件保存失败: {e}")
            return jsonify({
                "success": False,
                "message": f"文件保存失败: {str(e)}"
            })
        
        # 准备标签列表
        tag_list = [tag.strip() for tag in tags.split(',') if tag.strip()] if tags else []
        
        # 使用新的事件循环运行异步上传
        try:
            # 创建新的事件循环
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)
            
            result = loop.run_until_complete(async_upload_video(
                str(local_video_path), 
                title, 
                description, 
                tag_list, 
                category_id, 
                user_credential
            ))
            
            loop.close()
        except Exception as e:
            print(f"❌ 事件循环错误: {e}")
            return jsonify({
                "success": False,
                "message": f"上传过程中出现错误: {str(e)}"
            })
        
        # 清理本地文件
        try:
            if local_video_path.exists():
                local_video_path.unlink()
                print(f"🗑️ 临时文件已删除: {local_video_path}")
        except Exception as e:
            print(f"⚠️ 临时文件删除失败: {e}")
        
        return jsonify(result)
        
    except Exception as e:
        print(f"❌ 上传接口错误: {e}")
        import traceback
        traceback.print_exc()
        return jsonify({
            "success": False,
            "message": f"上传失败: {str(e)}"
        })

async def async_upload_video(video_path, title, description, tags, category_id, credential):
    """异步上传视频 - 完全参考test_upload.py的实现"""
    try:
        print(f"🎬 开始上传视频")
        print(f"   文件路径: {video_path}")
        print(f"   标题: {title}")
        print(f"   描述: {description[:50]}..." if description and len(description) > 50 else description or "无描述")
        print(f"   分区ID: {category_id}")
        print(f"   标签: {tags}")
        
        # 验证文件存在
        video_file_path = Path(video_path)
        if not video_file_path.exists():
            raise FileNotFoundError(f"视频文件不存在: {video_path}")
        
        # 创建默认封面 - 关键修复
        # cover = create_default_cover()
        cover = "./1.jpg"  # 使用本地图片路径代替
        print("✅ 封面图片创建成功")
        
        # 创建视频元数据 - 完全按照test_upload.py的参数
        meta = VideoMeta(
            tid=category_id,          # 分区ID
            title=title,              # 视频标题
            desc=description or "这是一个测试视频",  # 视频描述
            cover=cover,              # 封面图片 - 必须提供
            tags=tags,                # 标签列表
            original=True,            # True=原创 (对应test_upload.py的original=True)
            no_reprint=True,         # True=禁止转载 (对应test_upload.py的no_reprint=True)
            open_elec=True           # True=开启充电 (对应test_upload.py的open_elec=True)
        )
        
        print("✅ VideoMeta创建成功")
        
        # 创建视频页面对象 - 完全按照test_upload.py
        page = VideoUploaderPage(
            path=video_path,         # 视频文件路径
            title=title,             # 分P标题
            description=description or "测试描述"  # 分P描述
        )
        
        print("✅ VideoUploaderPage创建成功")
        
        # 创建上传器 - 完全按照test_upload.py
        uploader = VideoUploader(
            pages=[page],
            meta=meta,
            credential=credential
        )
        
        print("✅ VideoUploader创建成功，开始上传...")
        
        # 添加上传进度监听 - 参考test_upload.py
        @uploader.on("__ALL__")
        async def upload_event_handler(data):
            event_name = data.get('name', 'UNKNOWN')
            print(f"📡 上传事件: {event_name}")
            
            # 特殊处理一些关键事件
            if event_name == 'COMPLETE':
                event_data = data.get('data', [{}])
                if event_data and len(event_data) > 0:
                    result_data = event_data[0]
                    if 'bvid' in result_data:
                        print(f"🎉 投稿成功！BV号: {result_data['bvid']}")
        
        # 开始上传 - 关键步骤
        result = await uploader.start()
        
        print(f"🎉 上传完成！")
        print(f"结果: {result}")
        
        # 构造返回结果
        success_message = f"视频上传成功！标题：{title}"
        bvid = None
        aid = None
        video_url = None
        
        if isinstance(result, dict):
            if 'bvid' in result:
                bvid = result['bvid']
                success_message += f"\nBV号：{bvid}"
                video_url = f"https://www.bilibili.com/video/{bvid}"
            if 'aid' in result:
                aid = result['aid']
                success_message += f"\nAV号：av{aid}"
        
        return {
            "success": True,
            "message": success_message,
            "result": result,
            "bvid": bvid,
            "aid": aid,
            "video_url": video_url
        }
        
    except Exception as e:
        import traceback
        error_detail = traceback.format_exc()
        print(f"❌ 上传过程详细错误:")
        print(error_detail)
        
        return {
            "success": False,
            "message": f"上传过程出错: {str(e)}",
            "error_detail": str(e)
        }

if __name__ == '__main__':
    print("🚀 启动Python后端服务...")
    print(f"📁 上传目录: {UPLOAD_DIR}")
    print("📡 服务地址: http://localhost:5001")
    print("🔗 健康检查: http://localhost:5001/health")
    print("=" * 50)
    print("📋 测试建议:")
    print("1. 先访问健康检查确认服务启动")
    print("2. 使用前端进行B站登录获取Cookie")
    print("3. 上传测试视频进行验证")
    print("=" * 50)
    app.run(debug=True, port=5001, host='0.0.0.0')