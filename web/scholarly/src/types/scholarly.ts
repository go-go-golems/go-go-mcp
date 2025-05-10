// Types for the Scholarly API

export interface ScholarlyPaper {
  id?: string;
  Title: string;
  Authors: string[];
  Abstract: string;
  Published: string;
  DOI: string;
  PDFURL: string;
  SourceURL: string;
  SourceName: string;
  OAStatus: string;
  License: string;
  FileSize: string;
  Citations: number;
  Type: string;
  JournalInfo: string;
  Metadata?: Record<string, any>;
  reranked?: boolean;
  reranker_score?: number;
  original_index?: number;
}

export interface SearchParams {
  query?: string;
  sources?: string;
  limit?: number;
  author?: string;
  title?: string;
  category?: string;
  'work-type'?: string;
  'from-year'?: number;
  'to-year'?: number;
  sort?: 'relevance' | 'newest' | 'oldest';
  'open-access'?: string;
  mailto?: string;
  'disable-rerank'?: boolean;
}

export interface SearchResponse {
  results: ScholarlyPaper[];
  query: string;
  count: number;
  sources: string[];
}

export interface SourcesResponse {
  sources: string[];
}

export interface HealthResponse {
  status: string;
  version: string;
  timestamp: string;
}