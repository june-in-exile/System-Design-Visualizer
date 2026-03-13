#!/bin/bash

DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Starting backend..."
cd "$DIR/api" && go run _cmd/main.go &
BACKEND_PID=$!

sleep 2

echo "Starting frontend..."
cd "$DIR/frontend" && npm run dev &
FRONTEND_PID=$!

echo "Backend PID: $BACKEND_PID"
echo "Frontend PID: $FRONTEND_PID"
echo ""
echo "Development servers running:"
echo "  Backend:  http://localhost:8080"
echo "  Frontend: http://localhost:5173"
echo ""
echo "Press Ctrl+C to stop all servers"

trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null" EXIT

wait
