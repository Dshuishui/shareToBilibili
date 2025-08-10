# backend/app/utils.py
from pathlib import Path
from typing import Union
from fastapi import UploadFile
import aiofiles
from typing import Dict

async def save_upload_file(upload_file: UploadFile, destination: Union[str, Path]):
    """
    将 UploadFile 异步写入磁盘（按块读，避免一次性读入内存）
    """
    dest_path = str(destination)
    async with aiofiles.open(dest_path, "wb") as out_file:
        while True:
            chunk = await upload_file.read(1024 * 1024)  # 1MB
            if not chunk:
                break
            await out_file.write(chunk)
    await upload_file.close()

def parse_cookie_string(cookie_str: str) -> Dict[str, str]:
    """
    将形如 "SESSDATA=xxx; bili_jct=yyy; buvid3=zzz" 的 cookie 字符串解析成 dict。
    """
    out = {}
    if not cookie_str:
        return out
    # 如果 cookie_str 是 JSON（已经是对象格式），尽量兼容处理
    if isinstance(cookie_str, dict):
        return cookie_str
    s = cookie_str.strip()
    # 如果传的是 JSON 字符串（含大括号），尝试解析
    if s.startswith("{") and s.endswith("}"):
        try:
            import json
            return json.loads(s)
        except Exception:
            pass
    parts = [p.strip() for p in s.split(";") if p.strip()]
    for p in parts:
        if "=" in p:
            k, v = p.split("=", 1)
            out[k.strip()] = v.strip()
    return out
