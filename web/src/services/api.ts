import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

export interface ArxivPaper {
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
  Metadata: Record<string, any>;
}

export interface ScoredPaper extends ArxivPaper {
  score: number;
}

export interface RerankerRequest {
  query: string;
  results: ArxivPaper[];
  top_n?: number;
}

export interface RerankerResponse {
  query: string;
  reranked_results: ScoredPaper[];
}

export interface ModelsResponse {
  current_model: string;
  description: string;
  alternatives: string[];
}

// Define the API service
export const arxivApi = createApi({
  reducerPath: 'arxivApi',
  baseQuery: fetchBaseQuery({ baseUrl: 'http://localhost:8000' }),
  endpoints: (builder) => ({
    rerank: builder.mutation<RerankerResponse, RerankerRequest>({
      query: (request) => ({
        url: '/rerank',
        method: 'POST',
        body: request,
      }),
    }),
    getModels: builder.query<ModelsResponse, void>({
      query: () => '/models',
    }),
  }),
});

export const { useRerankMutation, useGetModelsQuery } = arxivApi;