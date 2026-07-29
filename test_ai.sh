#!/bin/bash
cd services/ai-service
python3 -c "import app.main" || echo "Failed to import ai-service"
