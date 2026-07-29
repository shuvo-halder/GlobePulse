from fastapi.testclient import TestClient
from app.main import app

client = TestClient(app)

def test_health_check():
    response = client.get("/health")
    assert response.status_code == 200
    assert response.json() == {"status": "healthy", "service": "AI Analysis Service"}

def test_analyze_endpoint():
    req_data = {
        "id": "test_id_123",
        "title": "Global Markets Rally Amid Positive Economic Data",
        "content": "Stocks surged on Wall Street today as new economic reports indicated inflation is cooling down.",
        "language": "en"
    }
    response = client.post("/api/v1/analysis/analyze", json=req_data)
    assert response.status_code == 200
    assert "task_id" in response.json()
