from fastapi import APIRouter
from app.api.endpoints import router as analysis_router

api_router = APIRouter()
api_router.include_router(analysis_router, prefix="/analysis", tags=["AI Analysis"])
