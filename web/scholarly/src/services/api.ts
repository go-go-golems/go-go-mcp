import axios from 'axios';
import { SearchParams, SearchResponse, SourcesResponse, HealthResponse } from '../types/scholarly';

// Base API URL
const API_BASE_URL = 'http://localhost:8080/api';

// API client
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// API functions
export const searchPapers = async (params: SearchParams): Promise<SearchResponse> => {
  console.log('Searching with params:', params);
  try {
    const response = await apiClient.get('/search', { params });
    console.log('Search API response:', response.data);
    return response.data;
  } catch (error) {
    console.error('Search API error:', error);
    throw error;
  }
};

export const getSources = async (): Promise<SourcesResponse> => {
  const response = await apiClient.get('/sources');
  return response.data;
};

export const getHealth = async (): Promise<HealthResponse> => {
  const response = await apiClient.get('/health');
  return response.data;
};