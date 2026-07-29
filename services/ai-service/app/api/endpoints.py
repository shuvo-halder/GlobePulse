from fastapi import APIRouter, HTTPException, BackgroundTasks
from app.models.schemas import AnalysisRequest, AnalysisResult
from app.worker.tasks import analyze_news_task
from app.core.redis import get_cache
import logging

router = APIRouter()
logger = logging.getLogger(__name__)

@router.post("/analyze", response_model=dict)
async def analyze_news_async(request: AnalysisRequest):
    """
    Submits a news article for async AI analysis.
    """
    try:
        task = analyze_news_task.delay(request.model_dump())
        return {"task_id": task.id, "message": "Analysis started asynchronously"}
    except Exception as e:
        logger.error(f"Failed to submit task: {e}")
        raise HTTPException(status_code=500, detail="Failed to submit analysis task")

@router.get("/result/{news_id}", response_model=AnalysisResult)
async def get_analysis_result(news_id: str):
    """
    Retrieves the analysis result for a given news ID.
    """
    result = get_cache(f"analysis:{news_id}")
    if not result:
        raise HTTPException(status_code=404, detail="Analysis result not found or processing")
    
    return AnalysisResult(**result)
