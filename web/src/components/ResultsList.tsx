import type { ScoredPaper } from '../services/api';

interface ResultsListProps {
  results: ScoredPaper[];
  query: string;
}

const ResultsList = ({ results, query }: ResultsListProps) => {
  if (results.length === 0) {
    return null;
  }

  return (
    <div className="card mb-4">
      <div className="card-header bg-success text-white">
        <h4 className="mb-0">Reranked Results for "{query}"</h4>
      </div>
      <div className="card-body p-0">
        <div className="list-group list-group-flush">
          {results.map((paper, index) => (
            <div key={index} className="list-group-item p-3">
              <div className="d-flex justify-content-between align-items-start">
                <h5 className="mb-1">{paper.Title}</h5>
                <span className="badge bg-primary rounded-pill ms-2">
                  Score: {paper.score.toFixed(2)}
                </span>
              </div>
              
              <p className="mb-1 text-muted">
                {paper.Authors.join(', ')}
              </p>
              
              <p className="mb-2">
                {paper.Abstract.length > 300 
                  ? `${paper.Abstract.substring(0, 300)}...` 
                  : paper.Abstract}
              </p>
              
              <div className="d-flex justify-content-between align-items-center">
                <small className="text-muted">
                  Published: {new Date(paper.Published).toLocaleDateString()}
                </small>
                
                {paper.PDFURL && (
                  <a 
                    href={paper.PDFURL} 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="btn btn-sm btn-outline-primary"
                  >
                    View PDF
                  </a>
                )}
                
                {paper.SourceURL && (
                  <a 
                    href={paper.SourceURL} 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="btn btn-sm btn-outline-secondary ms-2"
                  >
                    Source
                  </a>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default ResultsList;