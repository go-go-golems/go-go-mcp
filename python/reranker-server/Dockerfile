FROM python:3.9-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy source code
COPY arxiv_reranker_server.py .

# Expose port for API
EXPOSE 8000

# Command to run the server
CMD ["python", "arxiv_reranker_server.py"]