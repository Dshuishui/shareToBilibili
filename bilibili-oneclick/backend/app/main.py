import os
import uuid
import json
from pathlib import Path
from fastapi import FastAPI, UploadFile, File, Form, HTTPException
from fastapi.responses import JSONResponse, FileResponse, HTMLResponse
from fastapi.staticfiles import StaticFiles
from fastapi.middleware.cors import CORSMiddleware
from .utils import save_upload_file

BASE_DIR = Path(__file__).resolve().parent.parent
UPLOAD_DIR = Path(os.getenv("UPLOAD_DIR", BASE_DIR / "uploads"))
COOKIE_FILE = BASE_DIR / "cookies.json"
UPLOAD_DIR.mkdir(parents=True, exist_ok=True)

app = FastAPI(title="Bili One-Click — Stage2 (cookie + confirm)")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

static_dir = BASE_DIR / "static"
if static_dir.exists():
    app.mount("/static", StaticFiles(directory=static_dir), name="static")


@app.get("/", response_class=HTMLResponse)
async def serve_index():
    index_file = static_dir / "index.html"
    if index_file.exists():
        return FileResponse(index_file)
    return HTMLResponse(content="<h1>index.html not found</h1>", status_code=404)


@app.post("/upload")
async def upload(file: UploadFile = File(...),
                 title: str = Form(...),
                 description: str = Form(""),
                 tags: str = Form(""),
                 partition: str = Form("")):
    task_id = uuid.uuid4().hex
    task_dir = UPLOAD_DIR / task_id
    task_dir.mkdir(parents=True, exist_ok=True)

    dest = task_dir / file.filename
    await save_upload_file(file, dest)

    metadata = {
        "task_id": task_id,
        "filename": file.filename,
        "title": title,
        "description": description,
        "tags": tags,
        "partition": partition
    }
    (task_dir / "metadata.json").write_text(
        json.dumps(metadata, ensure_ascii=False, indent=2),
        encoding="utf-8"
    )

    return {"success": True, "task_id": task_id, "metadata": metadata}


@app.get("/task/{task_id}")
async def get_task(task_id: str):
    task_dir = UPLOAD_DIR / task_id
    if not task_dir.exists():
        raise HTTPException(status_code=404, detail="task 不存在")
    meta_path = task_dir / "metadata.json"
    if not meta_path.exists():
        raise HTTPException(status_code=404, detail="metadata 不存在")
    meta = json.loads(meta_path.read_text(encoding="utf-8"))
    return {"task_id": task_id, "metadata": meta}


@app.post("/save_cookie")
async def save_cookie(cookie: str = Form(...)):
    COOKIE_FILE.write_text(json.dumps({"cookie": cookie}), encoding="utf-8")
    return {"success": True, "message": "cookie 已保存"}


@app.get("/get_cookie")
async def get_cookie():
    if COOKIE_FILE.exists():
        return json.loads(COOKIE_FILE.read_text(encoding="utf-8"))
    return {"cookie": ""}
