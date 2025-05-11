# Scholarly Search UI

A React frontend for the Scholarly API, providing a user-friendly interface to search for academic papers across multiple sources including ArXiv, Crossref, and OpenAlex.

## Features

- Search for papers using various criteria (query text, author, title, etc.)
- Filter results by source, year range, and open access status
- Sort results by relevance, newest, or oldest
- View paper details including abstracts, authors, and publication info
- Access PDFs and source links directly from the interface

## Getting Started

### Prerequisites

- Bun.js (latest version)
- The Scholarly API server running (typically on http://localhost:8080)

### Installation

1. Install dependencies
```bash
bun install
```

2. Run the development server
```bash
bun run dev
```

3. Build for production
```bash
bun run build
```

## Deployment

The build process creates a `dist` directory with static files that can be served by any static file server. The Scholarly API's serve command is configured to serve these files.

## API Connection

This frontend connects to the Scholarly API running at the URL specified in `src/services/api.ts` (default: http://localhost:8080/api). Update this URL if your API is running on a different host or port.

## Technologies Used

- React
- TypeScript
- TanStack Query (React Query)
- Formik & Yup
- Bootstrap
- Axios