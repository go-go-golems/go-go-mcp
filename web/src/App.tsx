import React, { useState } from 'react';
import { Provider } from 'react-redux';
import { store } from './store/store';
import SearchForm from './components/SearchForm';
import ResultsList from './components/ResultsList';
import ModelInfo from './components/ModelInfo';
import { useRerankMutation } from './services/api';
import type { ArxivPaper, ScoredPaper } from './services/api';

import 'bootstrap/dist/css/bootstrap.min.css';
import './App.css';

const AppContent: React.FC = () => {
  const [rerank, { isLoading }] = useRerankMutation();
  const [results, setResults] = useState<ScoredPaper[]>([]);
  const [searchQuery, setSearchQuery] = useState('');

  const handleSearch = async (query: string, papers: ArxivPaper[]) => {
    try {
      const response = await rerank({
        query,
        results: papers,
        top_n: 10
      }).unwrap();
      
      setResults(response.reranked_results);
      setSearchQuery(query);
    } catch (error) {
      console.error('Failed to rerank papers:', error);
      alert('Failed to rerank papers. Please try again.');
    }
  };

  return (
    <div className="container py-4">
      <header className="pb-3 mb-4 border-bottom">
        <h1 className="display-5 fw-bold">ArXiv Paper Reranker</h1>
        <p className="lead">Rerank arXiv paper search results based on query relevance using cross-encoders</p>
      </header>

      <div className="row">
        <div className="col-lg-8">
          <SearchForm onSearch={handleSearch} isLoading={isLoading} />
          <ResultsList results={results} query={searchQuery} />
        </div>
        <div className="col-lg-4">
          <ModelInfo />
        </div>
      </div>
    </div>
  );
};

const App: React.FC = () => {
  return (
    <Provider store={store}>
      <AppContent />
    </Provider>
  );
};

export default App;