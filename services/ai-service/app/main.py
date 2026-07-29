import logging
from fastapi import FastAPI
from contextlib import asynccontextmanager
from app.core.config import settings
from app.api.router import api_router
from app.core.database import engine, Base
from app.services.ai_pipeline import get_ai_pipeline

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Initializing database schema...")
    Base.metadata.create_all(bind=engine)
    
    logger.info("Pre-loading AI models into memory...")
    get_ai_pipeline()
    
    yield
    
    logger.info("Shutting down AI Analysis Service...")

app = FastAPI(
    title=settings.APP_NAME,
    description="Transforms raw news text into structured intelligence using Hugging Face Transformers.",
    version="1.0.0",
    lifespan=lifespan
)

app.include_router(api_router, prefix="/api/v1")

@app.get("/health", tags=["System"])
def health_check():
    return {"status": "healthy", "service": settings.APP_NAME}
