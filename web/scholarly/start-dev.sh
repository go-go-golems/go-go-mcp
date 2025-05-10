#!/bin/bash

# Start the Scholarly API server in the background
cd ../../
go run ./cmd/apps/scholarly/main.go serve --port 8080 --cors-origins="http://localhost:5173" &
API_PID=$!

# Wait a moment for the server to start
sleep 2

# Start the React frontend
cd web/scholarly
bun run dev

# Handle cleanup when script is terminated
trap "kill $API_PID; exit" INT TERM EXIT