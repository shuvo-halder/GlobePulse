from celery import Celery
from app.core.config import settings

celery_app = Celery(
    "ai_service",
    broker=settings.RABBITMQ_URL,
    backend=f"redis://:{settings.REDIS_PASS}@{settings.REDIS_ADDR}:{settings.REDIS_PORT}/5" if settings.REDIS_PASS else f"redis://{settings.REDIS_ADDR}:{settings.REDIS_PORT}/5"
)

celery_app.conf.update(
    task_serializer="json",
    accept_content=["json"],
    result_serializer="json",
    timezone="UTC",
    enable_utc=True,
    task_routes={
        'app.worker.tasks.analyze_news_task': {'queue': 'ai_analysis_queue'},
    }
)
