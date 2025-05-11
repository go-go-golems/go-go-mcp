import { useState, useEffect } from 'react';
import { Provider } from 'react-redux';
import { store } from './store/store';
import { useLazySearchQuery, useGetSourcesQuery } from './store/scholarlyApi';
import type { SearchParams, ScholarlyPaper } from './types/scholarly';
import SearchForm from './components/SearchForm';
import PaperList from './components/PaperList';
import Header from './components/Header';

import 'bootstrap/dist/css/bootstrap.min.css';
import './App.css';

const AppContent = () => {
  const [searchResults, setSearchResults] = useState<ScholarlyPaper[]>([]);
  const [currentQuery, setCurrentQuery] = useState<string>('');
  const [currentSources, setCurrentSources] = useState<string[]>(['all']);

  // Fetch available sources
  const { 
    data: sourcesData,
    isLoading: isLoadingSources, 
    error: sourcesError 
  } = useGetSourcesQuery();

  // Setup search query hook
  const [
    triggerSearch, 
    { isLoading: isSearching, error: searchError, data: searchData }
  ] = useLazySearchQuery();

  const handleSearch = (params: SearchParams) => {
    console.log('Search requested with params:', params);
    triggerSearch(params, true)
      .unwrap()
      .then(data => {
        console.log('Search successful:', data);
        if (data?.results) {
          setSearchResults(data.results);
          setCurrentQuery(data.query || '');
          setCurrentSources(data.sources || ['all']);
        } else {
          console.warn('No results in response:', data);
          setSearchResults([]);
        }
      })
      .catch(err => {
        console.error('Search error:', err);
        setSearchResults([]);
      });
  };
  
  // Add effect to log searchResults changes
  useEffect(() => {
    console.log('Current search results:', searchResults);
  }, [searchResults]);

  // Available sources with fallback
  const availableSources = sourcesData || ['all', 'arxiv', 'crossref', 'openalex'];

  return (
    <div className="scholarly-app">
      <Header />
      
      <div className="container py-4">
        <div className="p-5 mb-4 bg-light rounded-3">
          <div className="container-fluid py-5">
            <h1 className="display-5 fw-bold">Scholarly Search</h1>
            <p className="col-md-8 fs-4">
              Search for academic papers across ArXiv, Crossref, and OpenAlex to find the latest research in your field.
            </p>
          </div>
        </div>

        <div className="row">
          <div className="col-lg-4">
            <SearchForm 
              onSearch={handleSearch} 
              isLoading={isSearching || isLoadingSources} 
              availableSources={availableSources}
            />
            
            {sourcesError && (
              <div className="alert alert-danger mt-3">
                Error loading sources: {(sourcesError as Error).message}
              </div>
            )}
            
            {searchError && (
              <div className="alert alert-danger mt-3">
                Search error: {(searchError as Error).message}
              </div>
            )}
          </div>
          
          <div className="col-lg-8">
            {isSearching ? (
              <div className="d-flex justify-content-center my-5">
                <div className="spinner-border text-primary" role="status">
                  <span className="visually-hidden">Loading...</span>
                </div>
              </div>
            ) : (
              searchResults.length > 0 && (
                <PaperList 
                  papers={searchResults} 
                  query={currentQuery} 
                  sources={currentSources}
                />
              )
            )}
            
            {searchResults.length === 0 && !isSearching && searchData && (
              <div className="alert alert-info">
                No results found for your search criteria. Try adjusting your search terms or filters.
              </div>
            )}
          </div>
        </div>
      </div>
      
      <footer className="footer mt-auto py-3 bg-light">
        <div className="container text-center">
          <span className="text-muted">Scholarly Search © {new Date().getFullYear()}</span>
        </div>
      </footer>
    </div>
  );
}

function App() {
  return (
    <Provider store={store}>
      <AppContent />
    </Provider>
  );
}

export default App;