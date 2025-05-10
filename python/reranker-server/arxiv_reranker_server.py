import json
import logging
from typing import List, Dict, Any, Optional, Union

from fastapi import FastAPI, HTTPException, Request
import traceback
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
import uvicorn

from sentence_transformers import CrossEncoder

import os

# Configure logging
log_level = os.environ.get('LOG_LEVEL', 'DEBUG')
log_file = os.environ.get('LOG_FILE', 'reranker.log')

# Convert string log level to actual level
log_level_map = {
    'DEBUG': logging.DEBUG,
    'INFO': logging.INFO,
    'WARNING': logging.WARNING,
    'ERROR': logging.ERROR,
    'CRITICAL': logging.CRITICAL
}

logging.basicConfig(
    level=log_level_map.get(log_level, logging.DEBUG),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(),
        logging.FileHandler(log_file)
    ]
)
logger = logging.getLogger("arxiv_reranker")

logger.info(f"Starting ArXiv Paper Reranker API using logging level {log_level} and log file {log_file}")

# Initialize the FastAPI app
app = FastAPI(
    title="ArXiv Paper Reranker API",
    description="A REST API for reranking arXiv paper search results based on query relevance using cross-encoders",
    version="1.0.0",
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Add middleware to catch validation errors
@app.middleware("http")
async def log_requests(request: Request, call_next):
    body = await request.body()
    if body:
        logger.debug(f"Request body: {body.decode()}")
    try:
        response = await call_next(request)
        if response.status_code == 422:
            logger.error(f"Validation error for request to {request.url.path}")
            logger.error(f"Request body: {body.decode() if body else 'Empty'}")
        return response
    except Exception as e:
        logger.error(f"Request error: {str(e)}")
        raise

# Get model name from environment variable or use default
model_name = os.environ.get('MODEL_NAME', 'cross-encoder/ms-marco-MiniLM-L-6-v2')
logger.info(f"Loading cross-encoder model: {model_name}")

# Load the cross-encoder model
try:
    cross_encoder = CrossEncoder(model_name)
    logger.info("Model loaded successfully")
except Exception as e:
    logger.error(f"Failed to load model: {e}", exc_info=True)
    # Fall back to default model if specified one fails
    if model_name != 'cross-encoder/ms-marco-MiniLM-L-6-v2':
        logger.info("Falling back to default model: cross-encoder/ms-marco-MiniLM-L-6-v2")
        cross_encoder = CrossEncoder('cross-encoder/ms-marco-MiniLM-L-6-v2')

# Define the request and response models
class ArxivPaper(BaseModel):
    Title: str
    Authors: List[str]
    Abstract: str
    Published: str
    DOI: str = ""
    PDFURL: str = ""
    SourceURL: str = ""
    SourceName: str = "arxiv"
    OAStatus: str = "green"
    License: str = ""
    FileSize: str = ""
    Citations: int = 0
    Type: str = ""
    JournalInfo: str = ""
    Metadata: Dict[str, Any] = Field(default_factory=dict)
    
    class Config:
        # Allow dict access and dict() conversion
        from_attributes = True  # For Pydantic v2 (orm_mode in v1)
        arbitrary_types_allowed = True
        schema_extra = {
            "example": {
                "Title": "Interpreting Multilingual and Document-Length Sensitive Relevance Computations in Neural Retrieval Models",
                "Authors": ["Oliver Savolainen", "Dur e Najaf Amjad", "Roxana Petcu"],
                "Abstract": "This reproducibility study analyzes and extends the paper...",
                "Published": "2025-05-04T15:30:45Z",
                "PDFURL": "http://arxiv.org/pdf/2505.02154v1",
                "SourceURL": "http://arxiv.org/abs/2505.02154v1",
                "Metadata": {"primary_category": "cs.IR"}
            }
        }

class RerankerRequest(BaseModel):
    query: str = Field(..., description="The search query or intent")
    results: List[ArxivPaper] = Field(..., description="The list of papers to rerank")
    top_n: Optional[int] = Field(10, description="Number of top results to return")
    
    class Config:
        from_attributes = True  # For Pydantic v2 (orm_mode in v1)
        arbitrary_types_allowed = True
        schema_extra = {
            "example": {
                "query": "cross encoders for neural retrieval",
                "results": [{"Title": "Example Paper", "Authors": ["Author 1"], "Abstract": "Abstract text", "Published": "2025-05-01T00:00:00Z"}],
                "top_n": 5
            }
        }

class ScoredPaper(ArxivPaper):
    score: float = Field(..., description="Relevance score from cross-encoder")

class RerankerResponse(BaseModel):
    query: str
    reranked_results: List[ScoredPaper]
    
    class Config:
        from_attributes = True  # For Pydantic v2 (orm_mode in v1)
        arbitrary_types_allowed = True
        schema_extra = {
            "example": {
                "query": "cross encoders for neural retrieval",
                "reranked_results": [
                    {
                        "Title": "Example Paper",
                        "Authors": ["Author 1"],
                        "Abstract": "Abstract text",
                        "Published": "2025-05-01T00:00:00Z",
                        "score": 9.45
                    }
                ]
            }
        }

@app.get("/")
async def root() -> Dict[str, str]:
    return {"message": "ArXiv Paper Reranker API. Use /docs for API documentation."}

@app.post("/rerank", response_model=RerankerResponse, 
          summary="Rerank ArXiv Papers",
          description="Reranks ArXiv paper search results based on their relevance to the query using a cross-encoder model")
async def rerank_papers(request: RerankerRequest) -> RerankerResponse:
    try:
        logger.debug(f"Reranking request received with query: '{request.query}' for {len(request.results)} papers")
        logger.debug(f"Request details: {request}")
        
        # Construct cross-encoder inputs (query, paper)
        cross_inputs = []
        papers = []
        
        for paper in request.results:
            # Create a context string from paper data
            # Title and abstract are most important for relevance
            paper_context = f"{paper.Title}. {paper.Abstract}"
            cross_inputs.append([request.query, paper_context])
            papers.append(paper)
        
        logger.debug(f"Prepared {len(cross_inputs)} paper-query pairs for cross-encoder scoring")
        
        # Get relevance scores from cross-encoder
        scores = cross_encoder.predict(cross_inputs)
        logger.debug(f"Cross-encoder scoring complete. Score range: {min(scores):.2f} to {max(scores):.2f}")
        
        # Pair papers with scores
        scored_papers = []
        for paper, score in zip(papers, scores):
            # Create a ScoredPaper combining the original paper with its score
            # Support both Pydantic v1 and v2
            if hasattr(paper, 'model_dump'):
                scored_paper_dict = paper.model_dump()
            else:
                scored_paper_dict = paper.dict()
                
            scored_paper_dict["score"] = float(score)
            scored_papers.append(ScoredPaper(**scored_paper_dict))
        
        # Sort by score (descending)
        scored_papers.sort(key=lambda x: x.score, reverse=True)
        
        # Limit to top_n results
        top_results = scored_papers[:request.top_n]
        logger.debug(f"Returning top {len(top_results)} results. Top score: {top_results[0].score:.2f}")
        
        # Construct the response object
        response = RerankerResponse(
            query=request.query,
            reranked_results=top_results
        )
        
        return response
        
    except Exception as e:
        logger.error(f"Reranking failed: {str(e)}", exc_info=True)
        stack_trace = traceback.format_exc()
        logger.error(f"Stack trace: {stack_trace}")
        raise HTTPException(status_code=500, detail=f"Reranking failed: {str(e)}")

@app.post("/rerank_json", 
          summary="Rerank ArXiv JSON Format",
          description="Reranks papers from the standard JSON format used in the ArXiv search results",
          response_model=RerankerResponse)
async def rerank_json(query: str, arxiv_json: Dict[str, Any], top_n: int = 10) -> RerankerResponse:
    try:
        logger.info(f"Reranking JSON data for query: '{query}', requesting top {top_n} results")
        
        # Extract papers from the standard JSON format
        if "results" not in arxiv_json:
            logger.error("Invalid JSON format: 'results' key missing")
            raise HTTPException(status_code=400, detail="Invalid JSON format. Expected 'results' key.")
        
        # Count papers in the data    
        paper_count = len(arxiv_json["results"])
        logger.debug(f"Found {paper_count} papers in JSON data")
            
        papers = [ArxivPaper(**paper) for paper in arxiv_json["results"]]
        
        # Create a reranker request
        reranker_request = RerankerRequest(
            query=query,
            results=papers,
            top_n=top_n
        )
        
        # Call the rerank endpoint
        logger.debug("Forwarding request to rerank_papers function")
        response = await rerank_papers(reranker_request)
        
        # Convert response to dict for better compatibility
        return response
        
    except Exception as e:
        logger.error(f"JSON reranking failed: {str(e)}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"JSON reranking failed: {str(e)}")

@app.get("/models", 
         summary="List Available Models",
         description="Returns information about the currently loaded cross-encoder model")
async def get_models() -> Dict[str, Any]:
    # Get model descriptions
    model_descriptions = {
        "cross-encoder/ms-marco-MiniLM-L-6-v2": "A small and fast cross-encoder model (6 layers) fine-tuned on MS MARCO passage ranking dataset",
        "cross-encoder/ms-marco-MiniLM-L-12-v2": "A medium-sized cross-encoder model (12 layers) fine-tuned on MS MARCO passage ranking dataset",
        "cross-encoder/ms-marco-electra-base": "An ELECTRA-based cross-encoder with strong performance on ranking tasks",
        "cross-encoder/nli-roberta-base": "A RoBERTa-based cross-encoder trained on natural language inference tasks"
    }
    
    # Get description for current model
    current_description = model_descriptions.get(
        model_name, 
        "A cross-encoder model for reranking papers based on query relevance"
    )
    
    logger.debug(f"Returning models information with current model: {model_name}")
    
    return {
        "current_model": model_name,
        "description": current_description,
        "alternatives": list(model_descriptions.keys())
    }

def load_sample_data(file_path: str) -> Dict[str, Any]:
    """Load sample data from a JSON file for testing"""
    with open(file_path, 'r') as f:
        return json.load(f)

if __name__ == "__main__":
    logger.info("Starting ArXiv Paper Reranker API...")
    logger.info("Loading cross-encoder model: cross-encoder/ms-marco-MiniLM-L-6-v2")
    logger.info("Visit http://localhost:8000/docs for API documentation")
    
    # Check if model is loaded correctly
    try:
        # Test the model with a simple prediction
        test_result = cross_encoder.predict([["test query", "test document"]])
        logger.info(f"Model loaded successfully. Test prediction: {test_result[0]:.4f}")
    except Exception as e:
        logger.error(f"Error loading model: {str(e)}", exc_info=True)
    
    # Start the server
    uvicorn.run(app, host="0.0.0.0", port=8000)
