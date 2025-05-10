#!/bin/bash

# Start the ArXiv Reranker server in the background
cd ../python/reranker-server
python arxiv_reranker_server.py &
PY_PID=$!

# Wait a moment for the server to start
sleep 2

# Start the React frontend
cd ../../web
bun run dev

# Handle cleanup when script is terminated
trap "kill $PY_PID; exit" INT TERM EXIT