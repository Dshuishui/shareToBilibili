# backend/app/utils.py
from pathlib import Path
from typing import Union
from fastapi import UploadFile
import aiofiles

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
