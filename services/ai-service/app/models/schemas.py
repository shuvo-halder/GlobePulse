from pydantic import BaseModel, Field
from typing import List, Optional
from datetime import datetime

class AnalysisRequest(BaseModel):
    id: str = Field(..., description="Unique ID of the news article")
    title: str = Field(..., description="Title of the news")
    content: str = Field(..., description="Main content of the news")
    source_url: Optional[str] = None
    language: str = "en"

class EntitySchema(BaseModel):
    name: str
    type: str
    confidence: float

class TopicSchema(BaseModel):
    name: str
    score: float

class AnalysisResult(BaseModel):
    news_id: str
    summary: str
    sentiment_score: float
    sentiment_label: str # POSITIVE, NEGATIVE, NEUTRAL
    entities: List[EntitySchema]
    topics: List[TopicSchema]
    countries: List[str]
    event_type: str
    importance_score: float # 0 to 1
    ai_insights: str
    processed_at: datetime
