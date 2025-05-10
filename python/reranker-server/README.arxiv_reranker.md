# ArXiv Paper Reranker API

This API service provides paper reranking capabilities for ArXiv search results using cross-encoder models. It reranks papers based on their relevance to a user query.

## Features

- Rerank ArXiv papers using cross-encoder models 
- Process standard ArXiv search result JSON format
- Configurable number of top results to return
- RESTful API with Swagger documentation

## Requirements

- Python 3.8+
- Dependencies listed in `requirements.txt`

## Installation

```bash
# Create and activate a virtual environment (recommended)
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt
```

## Running the Server

```bash
python arxiv_reranker_server.py
```

The server will start on `http://localhost:8000`. Visit `http://localhost:8000/docs` for interactive API documentation.

## API Endpoints

### Rerank Papers

**POST** `/rerank`

Reranks a list of ArXiv papers based on their relevance to a query.

**Request Body:**
```json
{
  "query": "cross encoders for neural retrieval",
  "results": [
    {
      "Title": "Paper Title",
      "Authors": ["Author 1", "Author 2"],
      "Abstract": "Paper abstract text...",
      "Published": "2025-05-01T00:00:00Z",
      "PDFURL": "http://arxiv.org/pdf/2505.00000v1",
      "SourceURL": "http://arxiv.org/abs/2505.00000v1"
    }
  ],
  "top_n": 5
}
```

**Response:**
```json
{
  "query": "cross encoders for neural retrieval",
  "reranked_results": [
    {
      "Title": "Paper Title",
      "Authors": ["Author 1", "Author 2"],
      "Abstract": "Paper abstract text...",
      "Published": "2025-05-01T00:00:00Z",
      "PDFURL": "http://arxiv.org/pdf/2505.00000v1",
      "SourceURL": "http://arxiv.org/abs/2505.00000v1",
      "score": 9.45
    }
  ]
}
```

### Rerank JSON

**POST** `/rerank_json`

Reranks papers from the standard ArXiv search results JSON format.

**Parameters:**
- `query` (string): The search query
- `arxiv_json` (object): ArXiv search results JSON  
- `top_n` (integer, optional): Number of top results to return (default: 10)

### Available Models

**GET** `/models`

Returns information about the currently loaded cross-encoder model and available alternatives.

### Example Usage with curl

```bash
curl -X POST http://localhost:8000/rerank \
  -H "Content-Type: application/json" \
  -d '{
    "query": "cross encoders for neural retrieval",
    "results": [
      {
        "Title": "Improving Neural Information Retrieval with Cross-Encoders",
        "Authors": ["Jane Smith", "John Doe"],
        "Abstract": "We present a novel approach to neural information retrieval using cross-encoder architectures...",
        "Published": "2025-05-01T00:00:00Z",
        "PDFURL": "http://arxiv.org/pdf/2505.00000v1",
        "SourceURL": "http://arxiv.org/abs/2505.00000v1"
      }
    ],
    "top_n": 5
  }'
```

## How It Works

The service uses a cross-encoder model from the SentenceTransformers library to compute relevance scores between the user query and each paper. The papers are then sorted by these scores in descending order.

For each paper, the model primarily uses the title and abstract to determine relevance, as these fields contain the most important semantic information about the paper's content.

## Performance Considerations

- The first request may be slower as the model needs to be loaded into memory
- Cross-encoder scoring is computationally intensive compared to bi-encoders
- For large result sets (>100 papers), consider using a bi-encoder first to pre-filter results

## Advanced Usage

To integrate this reranker with an existing arXiv search system:

1. Perform the initial search using your existing system
2. Pass the search results and user query to the `/rerank_json` endpoint
3. Display the reranked results to the user

This approach improves search relevance while maintaining the benefits of your existing search infrastructure.