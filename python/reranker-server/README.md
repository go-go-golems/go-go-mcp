# ArXiv Paper Reranker Server

This is a FastAPI server that provides API endpoints for reranking arXiv paper search results based on query relevance using cross-encoder models.

## Setup

### Prerequisites

- Python 3.8+
- PyTorch
- sentence-transformers
- FastAPI
- uvicorn

### Installation

1. Install dependencies
```bash
pip install sentence-transformers fastapi uvicorn pydantic torch
```

2. Run the server
```bash
python arxiv_reranker_server.py
```

The server will start on http://localhost:8000

## API Endpoints

- `GET /`: Root endpoint with basic information
- `POST /rerank`: Rerank ArXiv papers based on query relevance
- `POST /rerank_json`: Rerank papers from standard ArXiv JSON format
- `GET /models`: Get information about the currently loaded model

## Frontend

A React frontend for this API is available in the `web/` directory. To run both the frontend and backend together, use the `start-dev.sh` script in the web directory:

```bash
cd web
./start-dev.sh
```