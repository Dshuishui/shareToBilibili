# backend/app/models.py
from pydantic import BaseModel
from typing import Optional

class UploadMetadata(BaseModel):
    title: str
    description: Optional[str] = None
    tags: Optional[str] = None
    partition: Optional[str] = None
    filename: str
