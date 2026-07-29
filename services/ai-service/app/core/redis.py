import redis
from app.core.config import settings
import json

redis_client = redis.Redis(
    host=settings.REDIS_ADDR,
    port=int(settings.REDIS_PORT),
    password=settings.REDIS_PASS if settings.REDIS_PASS else None,
    db=4, # DB 4 for AI Service
    decode_responses=True
)

def set_cache(key: str, value: dict, expire_seconds: int = 3600):
    redis_client.setex(key, expire_seconds, json.dumps(value))

def get_cache(key: str) -> dict | None:
    data = redis_client.get(key)
    if data:
        return json.loads(data)
    return None
