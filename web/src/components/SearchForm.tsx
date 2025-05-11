import { useState } from 'react';
import type { ArxivPaper } from '../services/api';
import axios from 'axios';

interface SearchFormProps {
  onSearch: (query: string, papers: ArxivPaper[]) => void;
  isLoading: boolean;
}

const SearchForm = ({ onSearch, isLoading }: SearchFormProps) => {
  const [query, setQuery] = useState('');
  const [jsonInput, setJsonInput] = useState('');
  const [error, setError] = useState('');
  const [isSampleLoading, setSampleLoading] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!query.trim()) {
      setError('Please enter a search query');
      return;
    }

    try {
      // Parse the JSON input
      const data = JSON.parse(jsonInput);
      if (!data.results || !Array.isArray(data.results)) {
        setError('Invalid JSON format. Expected "results" array.');
        return;
      }

      // Call the search function
      onSearch(query, data.results);
      setError('');
    } catch (err) {
      setError('Invalid JSON: ' + (err as Error).message);
    }
  };
  
  const loadSampleData = async () => {
    try {
      setSampleLoading(true);
      setError('');
      const response = await axios.get('/sample-arxiv.json');
      setJsonInput(JSON.stringify(response.data, null, 2));
      setQuery('cross encoders for neural retrieval');
      setSampleLoading(false);
    } catch (err) {
      setError('Failed to load sample data');
      setSampleLoading(false);
    }
  };

  return (
    <div className="card mb-4">
      <div className="card-header bg-primary text-white">
        <h4 className="mb-0">ArXiv Paper Reranker</h4>
      </div>
      <div className="card-body">
        <form onSubmit={handleSubmit}>
          <div className="mb-3">
            <label htmlFor="query" className="form-label">Search Query</label>
            <input
              type="text"
              className="form-control"
              id="query"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Enter your search query"
              disabled={isLoading}
            />
          </div>
          
          <div className="mb-3">
            <label htmlFor="jsonInput" className="form-label">ArXiv JSON Results</label>
            <textarea
              className="form-control"
              id="jsonInput"
              rows={10}
              value={jsonInput}
              onChange={(e) => setJsonInput(e.target.value)}
              placeholder='Paste ArXiv JSON results with format: {"results": [...]}'
              disabled={isLoading}
            />
          </div>

          {error && (
            <div className="alert alert-danger" role="alert">
              {error}
            </div>
          )}

          <div className="d-flex gap-2">
            <button 
              type="submit" 
              className="btn btn-primary" 
              disabled={isLoading || isSampleLoading}
            >
              {isLoading ? (
                <>
                  <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                  Reranking...
                </>
              ) : 'Rerank Papers'}
            </button>
            
            <button 
              type="button" 
              className="btn btn-outline-secondary" 
              onClick={loadSampleData}
              disabled={isLoading || isSampleLoading}
            >
              {isSampleLoading ? (
                <>
                  <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                  Loading...
                </>
              ) : 'Load Sample Data'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default SearchForm;