#!/bin/bash
echo "Starting redis"
docker compose up -d redis

echo "Starting postgres"
docker compose up -d postgres

echo "Starting frontend on port 3001"
cd frontend
npm run dev &
FRONTEND_PID=$!

echo "Starting backend on port 3000"
cd ../backend
go run main.go &
BACKEND_PID=$!

# Wait for both processes
wait $FRONTEND_PID $BACKEND_PID
