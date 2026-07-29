#!/bin/bash

# news-service
mkdir -p services/news-service/cmd/api
cat << 'MOD' > services/news-service/go.mod
module github.com/global-news/news-service

go 1.21
MOD
cat << 'MAIN' > services/news-service/cmd/api/main.go
package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status": "ok"}`)
	})
	fmt.Printf("Starting news-service on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
MAIN
cat << 'DOCKER' > services/news-service/Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./services/news-service/cmd/api/main.go || (cd services/news-service && CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/api/main.go)

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
DOCKER

# country-service
mkdir -p services/country-service/cmd/api services/country-service/migrations
cat << 'MOD' > services/country-service/go.mod
module github.com/global-news/country-service

go 1.21
MOD
cat << 'MAIN' > services/country-service/cmd/api/main.go
package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status": "ok"}`)
	})
	fmt.Printf("Starting country-service on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
MAIN
cat << 'DOCKER' > services/country-service/Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./services/country-service/cmd/api/main.go || (cd services/country-service && CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/api/main.go)

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/services/country-service/migrations ./migrations
EXPOSE 8082
CMD ["./main"]
DOCKER

# analytics-service
mkdir -p services/analytics-service/cmd/api services/analytics-service/migrations
cat << 'MOD' > services/analytics-service/go.mod
module github.com/global-news/analytics-service

go 1.21
MOD
cat << 'MAIN' > services/analytics-service/cmd/api/main.go
package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status": "ok"}`)
	})
	fmt.Printf("Starting analytics-service on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
MAIN
cat << 'DOCKER' > services/analytics-service/Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./services/analytics-service/cmd/api/main.go || (cd services/analytics-service && CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/api/main.go)

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/services/analytics-service/migrations ./migrations
EXPOSE 8084
CMD ["./main"]
DOCKER

# ai-service
mkdir -p services/ai-service/app/core
cat << 'REQ' > services/ai-service/requirements.txt
fastapi==0.104.1
uvicorn==0.24.0.post1
celery==5.3.6
pydantic==2.5.2
REQ
cat << 'MAIN' > services/ai-service/app/main.py
from fastapi import FastAPI
app = FastAPI()

@app.get("/health")
def health_check():
    return {"status": "ok"}
MAIN
cat << 'CELERY' > services/ai-service/app/core/celery_app.py
from celery import Celery
import os

celery_app = Celery(
    "worker",
    broker=os.environ.get("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
    backend=os.environ.get("REDIS_URL", "redis://redis:6379/0")
)
CELERY
cat << 'DOCKER' > services/ai-service/Dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY services/ai-service/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY services/ai-service/ .
ENV PYTHONPATH=/app
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8083"]
DOCKER

# Fix docker-compose if needed
