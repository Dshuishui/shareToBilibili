# app/main.py
import os
import json
import uuid
from fastapi import FastAPI, File, Form, UploadFile
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles

from bilibili_api import login, Credential
from bilibili_api.video_uploader import VideoUploader, VideoUploaderPage, VideoMeta

app = FastAPI()

# 静态资源目录（前端 HTML/JS/CSS）
app.mount("/static", StaticFiles(directory="backend/static"), name="static")

# 存任务信息
tasks = {}

# Cookie 存储路径
COOKIE_FILE = "cookies.json"

# 读取已保存 cookie
def load_cookie():
    if os.path.exists(COOKIE_FILE):
        with open(COOKIE_FILE, "r", encoding="utf-8") as f:
            return json.load(f)
    return {}

# 保存 cookie
def save_cookie(cookie_dict):
    with open(COOKIE_FILE, "w", encoding="utf-8") as f:
        json.dump(cookie_dict, f, ensure_ascii=False, indent=2)

# ------------------
# 首页 HTML
# ------------------
@app.get("/")
async def index():
    return FileResponse("backend/static/index.html")

# ------------------
# 扫码登录
# ------------------
@app.get("/qrcode_login")
async def qrcode_login():
    # 获取扫码 URL 和登录 Key
    qrcode_info = await login.login_with_qrcode()
    return qrcode_info

@app.post("/save_cookie")
async def save_cookie_endpoint(cookie: dict):
    save_cookie(cookie)
    return {"status": "ok"}

# ------------------
# 上传视频 + 保存任务信息
# ------------------
@app.post("/upload")
async def upload(
    file: UploadFile = File(...),
    title: str = Form(...),
    description: str = Form(""),
    tags: str = Form(""),
    partition: str = Form("171")  # 默认动画区
):
    # 保存视频文件
    file_id = str(uuid.uuid4())
    save_path = os.path.join("uploads", f"{file_id}_{file.filename}")
    os.makedirs("uploads", exist_ok=True)
    with open(save_path, "wb") as f:
        f.write(await file.read())

    # 保存任务数据
    tasks[file_id] = {
        "filename": save_path,
        "title": title,
        "description": description,
        "tags": tags,
        "partition": partition
    }

    return {"task_id": file_id}

# ------------------
# 发布视频
# ------------------
@app.post("/publish/{task_id}")
async def publish(task_id: str):
    task = tasks.get(task_id)
    if not task:
        return {"error": "任务不存在"}

    # 读取 cookie
    saved_cookie = load_cookie()
    sessdata = saved_cookie.get("SESSDATA")
    bili_jct = saved_cookie.get("bili_jct")
    buvid3 = saved_cookie.get("buvid3")

    if not all([sessdata, bili_jct]):
        return {"error": "请先扫码登录"}

    credential = Credential(
        sessdata=sessdata,
        bili_jct=bili_jct,
        buvid3=buvid3
    )

    # 创建 meta
    meta = VideoMeta(
        tid=int(task["partition"]),
        title=task["title"],
        desc=task["description"],
        cover="",  # 可选封面路径
        tags=task["tags"].split(",") if task["tags"] else [],
        no_reprint=True,
        open_elec=False
    )

    page = VideoUploaderPage(
        path=task["filename"],
        title=task["title"],
        description=task["description"]
    )

    uploader = VideoUploader(
        pages=[page],
        meta=meta,
        credential=credential
    )

    await uploader.start()

    return {"status": "ok", "message": "投稿任务已提交"}
