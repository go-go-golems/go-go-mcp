# ArXiv Paper Reranker UI

A React frontend for interacting with the ArXiv Paper Reranker API. This application allows you to rerank ArXiv papers based on their relevance to a search query using cross-encoder models.

## Features

- Submit ArXiv search results for reranking
- View reranked papers with relevance scores
- See information about the active reranking model
- Access paper PDFs and source URLs

## Getting Started

### Prerequisites

- Bun.js (latest version)
- The ArXiv Reranker Server running on http://localhost:8000

### Installation

1. Install dependencies
```
bun install
```

2. Run the development server
```
bun run dev
```

3. Build for production
```
bun run build
```

## Usage

1. Enter your search query
2. Paste ArXiv JSON results in the textarea (must include a "results" array with paper objects)
3. Click "Rerank Papers"
4. View the reranked results sorted by relevance score

## Technologies Used

- React
- Redux Toolkit (RTK Query)
- Bootstrap
- Axios
- TypeScript

## API Connection

This frontend connects to the ArXiv Paper Reranker API running on http://localhost:8000. Make sure the API is running before using this UI.