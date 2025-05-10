# Scholarly - Academic Paper Search Tool

## Overview

Scholarly is a comprehensive academic paper search tool that allows searching across multiple sources including ArXiv, Crossref, and OpenAlex. It provides both a command-line interface and a web API with a modern React frontend.

## Components

### CLI Commands

Scholarly provides several CLI commands to search different academic sources:

```bash
# Search across all sources
scholarly search --query "quantum computing" --limit 10

# Search with specific filters
scholarly search --author "Hinton" --from-year 2020 --source arxiv

# Run the API server
scholarly serve --port 8080
```

### Web API

The Scholarly API provides RESTful endpoints for searching academic papers programmatically. See [scholarly-api.md](scholarly-api.md) for detailed API documentation.

```bash
# Start the API server
scholarly serve --port 8080 --cors-origins="http://localhost:5173"
```

The API server also serves the React frontend from the `web/scholarly/dist` directory.

### React Frontend

The Scholarly web UI is a modern React application that provides a user-friendly interface for searching academic papers. It's located in the `web/scholarly` directory.

```bash
# Install dependencies
cd web/scholarly
bun install

# Start development server
bun run dev

# Build for production
bun run build
```

## Development Workflow

For development, you can use the provided start script that runs both the API server and the React frontend:

```bash
cd web/scholarly
./start-dev.sh
```

This will start the API server on port 8080 and the React development server on port 5173.

## Production Deployment

For production deployment:

1. Build the React frontend:
   ```bash
   cd web/scholarly
   bun run build
   ```

2. Start the API server, which will serve the built frontend files:
   ```bash
   scholarly serve --port 8080
   ```

## Features

- Search across multiple academic sources (ArXiv, Crossref, OpenAlex)
- Filter by author, year, category, etc.
- Sort by relevance, newest, or oldest
- Reranking of search results for improved relevance
- Open access filtering
- Web API for programmatic access
- Modern React UI for user-friendly searching