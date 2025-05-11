import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { SearchParams, ScholarlyPaper } from '../types/scholarly';

// Define response types
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

// Define the API
export const scholarlyApi = createApi({
  reducerPath: 'scholarlyApi',
  baseQuery: fetchBaseQuery({ baseUrl: 'http://localhost:8080/api' }),
  endpoints: (builder) => ({
    search: builder.query<SearchResponse, SearchParams>({
      query: (params) => ({
        url: '/search',
        params,
      }),
      transformResponse: (response: SearchResponse) => {
        console.log('RTK Query received response:', response);
        return response;
      },
    }),
    getSources: builder.query<string[], void>({
      query: () => '/sources',
      transformResponse: (response: SourcesResponse) => response.sources,
    }),
    getHealth: builder.query<HealthResponse, void>({
      query: () => '/health',
    }),
  }),
});

// Export the auto-generated hooks
export const { 
  useSearchQuery, 
  useLazySearchQuery,
  useGetSourcesQuery, 
  useGetHealthQuery 
} = scholarlyApi;