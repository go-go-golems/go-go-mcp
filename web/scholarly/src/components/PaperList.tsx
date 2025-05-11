import React from 'react';
import { ScholarlyPaper } from '../types/scholarly';

interface PaperListProps {
  papers: ScholarlyPaper[];
  query: string;
  sources: string[];
}

const PaperList: React.FC<PaperListProps> = ({ papers, query, sources }) => {
  console.log('PaperList render with papers:', papers);
  
  if (papers.length === 0) {
    console.log('No papers to display');
    return null;
  }

  return (
    <div className="card mb-4 shadow-sm">
      <div className="card-header bg-success text-white d-flex justify-content-between">
        <h4 className="mb-0">Results</h4>
        <span className="badge bg-light text-dark">{papers.length} papers found</span>
      </div>
      <ul className="list-group list-group-flush">
        {papers.map((paper, index) => (
          <li key={paper.DOI || paper.SourceURL || index} className="list-group-item p-3">
            <div className="d-flex justify-content-between align-items-start">
              <h5 className="mb-1">{paper.Title}</h5>
              {paper.reranked && (
                <span className="badge bg-primary rounded-pill ms-2">
                  Score: {paper.reranker_score?.toFixed(2)}
                </span>
              )}
            </div>
            
            <p className="mb-1 text-muted">
              {paper.Authors.join(', ')}
            </p>
            
            <div className="mb-2 small">
              <span className="badge bg-secondary me-2">{paper.SourceName}</span>
              {paper.Published && <span className="badge bg-light text-dark me-2">{paper.Published.substring(0, 4)}</span>}
              {paper.OAStatus && <span className="badge bg-success me-2">Open Access</span>}
              {paper.Citations > 0 && (
                <span className="badge bg-info text-dark me-2">Citations: {paper.Citations}</span>
              )}
            </div>
            
            <p className="mb-3">
              {paper.Abstract.length > 300 
                ? `${paper.Abstract.substring(0, 300)}...` 
                : paper.Abstract}
            </p>
            
            <div className="d-flex gap-2">
              {paper.PDFURL && (
                <a 
                  href={paper.PDFURL} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="btn btn-sm btn-outline-primary"
                >
                  PDF
                </a>
              )}
              
              {paper.DOI && (
                <a 
                  href={`https://doi.org/${paper.DOI}`} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="btn btn-sm btn-outline-secondary"
                >
                  DOI
                </a>
              )}
              
              {paper.SourceURL && (
                <a 
                  href={paper.SourceURL} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="btn btn-sm btn-outline-secondary"
                >
                  Source
                </a>
              )}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default PaperList;