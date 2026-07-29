from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    APP_NAME: str = "AI Analysis Service"
    PORT: int = 8083
    APP_ENV: str = "development"
    
    DB_HOST: str = "postgres"
    DB_PORT: str = "5432"
    DB_USER: str = "devuser"
    DB_PASS: str = "devpassword"
    DB_NAME: str = "globepulse"
    
    REDIS_ADDR: str = "redis"
    REDIS_PORT: str = "6379"
    REDIS_PASS: str = ""
    
    RABBITMQ_URL: str = "amqp://guest:guest@rabbitmq:5672/"
    
    class Config:
        env_file = ".env"
        case_sensitive = True

settings = Settings()
