import logging
from datetime import datetime
from app.core.celery_app import celery_app
from app.services.ai_pipeline import get_ai_pipeline
from app.models.schemas import AnalysisRequest, AnalysisResult
from app.core.redis import set_cache

logger = logging.getLogger(__name__)

@celery_app.task(name="app.worker.tasks.analyze_news_task", bind=True, max_retries=3)
def analyze_news_task(self, request_dict: dict):
    logger.info(f"Starting analysis for news {request_dict['id']}")
    try:
        request = AnalysisRequest(**request_dict)
        pipeline = get_ai_pipeline()
        
        text_to_analyze = f"{request.title}. {request.content}"
        
        summary = pipeline.summarize(text_to_analyze)
        sentiment_label, sentiment_score = pipeline.analyze_sentiment(text_to_analyze)
        entities, countries = pipeline.extract_entities_and_countries(text_to_analyze)
        topics, event_type = pipeline.detect_topics_and_events(text_to_analyze)
        insights = pipeline.generate_insights(text_to_analyze)
        importance = pipeline.score_importance(text_to_analyze)
        
        # We could also store embeddings here for similarity search
        # embedding = pipeline.get_embedding(text_to_analyze)
        
        result = AnalysisResult(
            news_id=request.id,
            summary=summary,
            sentiment_score=sentiment_score,
            sentiment_label=sentiment_label,
            entities=entities,
            topics=topics,
            countries=countries,
            event_type=event_type,
            importance_score=importance,
            ai_insights=insights,
            processed_at=datetime.utcnow()
        )
        
        result_dict = result.model_dump(mode='json')
        
        # Store result in Redis for quick access
        set_cache(f"analysis:{request.id}", result_dict, expire_seconds=86400) # 24h
        
        logger.info(f"Successfully analyzed news {request.id}")
        return result_dict
    except Exception as exc:
        logger.error(f"Error analyzing news {request_dict['id']}: {exc}")
        self.retry(exc=exc, countdown=60)
