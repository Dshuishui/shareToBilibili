import asyncio
import json
import os
import tempfile
import shutil
from pathlib import Path
from flask import Flask, request, jsonify
from flask_cors import CORS
from bilibili_api import Credential
from bilibili_api.video_uploader import VideoUploader, VideoUploaderPage, VideoMeta

app = Flask(__name__)
CORS(app)

# 全局变量存储用户凭证
user_credential = None

# 确保uploads目录存在
UPLOAD_DIR = Path(__file__).parent / "uploads"
UPLOAD_DIR.mkdir(exist_ok=True)

@app.route('/health', methods=['GET'])
def health_check():
    """健康检查接口"""
    return jsonify({
        "status": "ok",
        "message": "Python后端服务运行正常",
        "service": "bilibili-api-backend",
        "upload_dir": str(UPLOAD_DIR)
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
        sessdata = None
        bili_jct = None
        buvid3 = None
        dedeuserid = None
        
        for cookie in cookies:
            if cookie['name'] == 'SESSDATA':
                sessdata = cookie['value']
            elif cookie['name'] == 'bili_jct':
                bili_jct = cookie['value']
            elif cookie['name'] == 'buvid3':
                buvid3 = cookie['value']
            elif cookie['name'] == 'DedeUserID':
                dedeuserid = cookie['value']
        
        if not all([sessdata, bili_jct, buvid3]):
            return jsonify({
                "success": False,
                "message": f"缺少必要的Cookie信息。需要：SESSDATA, bili_jct, buvid3"
            })
        
        # 创建凭证对象
        global user_credential
        user_credential = Credential(
            sessdata=sessdata,
            bili_jct=bili_jct,
            buvid3=buvid3,
            dedeuserid=dedeuserid  # 可选，但推荐
        )
        
        print(f"✅ 用户凭证创建成功")
        print(f"   SESSDATA: {sessdata[:20]}...")
        print(f"   bili_jct: {bili_jct[:10]}...")
        print(f"   buvid3: {buvid3[:20]}...")
        if dedeuserid:
            print(f"   DedeUserID: {dedeuserid}")
        
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
    """上传视频接口"""
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
        
        # 保存视频文件到本地
        try:
            # 创建安全的文件名
            safe_filename = f"{int(asyncio.get_event_loop().time())}_{video_file.filename}"
            local_video_path = UPLOAD_DIR / safe_filename
            
            print(f"📁 保存视频文件到: {local_video_path}")
            video_file.save(str(local_video_path))
            
            # 验证文件是否成功保存
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
        
        # 异步上传视频
        def run_upload():
            return asyncio.run(async_upload_video(
                str(local_video_path), 
                title, 
                description, 
                tag_list, 
                int(category_id), 
                user_credential
            ))
        
        result = run_upload()
        
        # 清理本地文件
        try:
            local_video_path.unlink()
            print(f"🗑️ 临时文件已删除: {local_video_path}")
        except Exception as e:
            print(f"⚠️ 临时文件删除失败: {e}")
        
        return jsonify(result)
        
    except Exception as e:
        print(f"❌ 上传接口错误: {e}")
        return jsonify({
            "success": False,
            "message": f"上传失败: {str(e)}"
        })

async def async_upload_video(video_path, title, description, tags, category_id, credential):
    """异步上传视频的具体实现"""
    try:
        print(f"🎬 开始上传视频")
        print(f"   文件路径: {video_path}")
        print(f"   标题: {title}")
        print(f"   分区ID: {category_id}")
        print(f"   标签: {tags}")
        
        # 验证文件存在
        video_file_path = Path(video_path)
        if not video_file_path.exists():
            raise FileNotFoundError(f"视频文件不存在: {video_path}")
        
        # 创建视频元数据
        meta = VideoMeta(
            tid=category_id,          # 分区ID
            title=title,              # 视频标题
            desc=description,         # 视频描述
            tags=tags,                # 标签列表
            copyright=1,              # 1为原创，2为转载
            no_reprint=1,            # 1为禁止转载
            open_elec=1              # 1为开启充电
        )
        
        print("✅ VideoMeta创建成功")
        
        # 创建视频页面对象
        page = VideoUploaderPage(
            path=video_path,         # 视频文件路径
            title=title,             # 分P标题
            description=description   # 分P描述
        )
        
        print("✅ VideoUploaderPage创建成功")
        
        # 创建上传器
        uploader = VideoUploader(
            pages=[page],
            meta=meta,
            credential=credential
        )
        
        print("✅ VideoUploader创建成功，开始上传...")
        
        # 添加上传事件监听
        @uploader.on("__ALL__")
        async def upload_event_handler(data):
            print(f"📡 上传事件: {data}")
        
        # 开始上传
        result = await uploader.start()
        
        print(f"🎉 上传成功！")
        if isinstance(result, dict) and 'bvid' in result:
            print(f"   BV号: {result['bvid']}")
        
        return {
            "success": True,
            "message": f"视频上传成功！标题：{title}",
            "result": result
        }
        
    except Exception as e:
        import traceback
        error_detail = traceback.format_exc()
        print(f"❌ 上传过程详细错误:")
        print(error_detail)
        
        return {
            "success": False,
            "message": f"上传过程出错: {str(e)}"
        }

if __name__ == '__main__':
    print("🚀 启动Python后端服务...")
    print(f"📁 上传目录: {UPLOAD_DIR}")
    print("📡 服务地址: http://localhost:5001")
    print("🔗 健康检查: http://localhost:5001/health")
    app.run(debug=True, port=5001, host='0.0.0.0')