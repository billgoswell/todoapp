/**
 * HTTP API Client
 *
 * Handles all communication with the CommandLineTodo server.
 * Uses Bearer token authentication via API key.
 */

import axios, { AxiosInstance, AxiosError, AxiosResponse } from 'axios';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { Task, TodoList, PullRequest, PullResponse, PushRequest, PushResponse, HealthResponse, ApiError } from './types';
import { TIMEOUTS } from '../utils/constants';

export class ApiClient {
  private client: AxiosInstance;
  private serverUrl: string = '';
  private apiKey: string = '';

  constructor() {
    this.client = axios.create({
      timeout: TIMEOUTS.NETWORK,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add interceptor for error handling
    this.client.interceptors.response.use(
      response => response,
      error => this.handleError(error)
    );
  }

  /**
   * Initialize the API client with stored credentials
   */
  async initialize(): Promise<void> {
    try {
      const storedUrl = await AsyncStorage.getItem('server_url');
      const storedKey = await AsyncStorage.getItem('api_key');

      if (storedUrl) {
        this.serverUrl = storedUrl;
      }
      if (storedKey) {
        this.apiKey = storedKey;
      }

      this.updateBaseURL();
    } catch (error) {
      console.error('Failed to initialize API client:', error);
    }
  }

  /**
   * Set API credentials and save to storage
   */
  async setCredentials(serverUrl: string, apiKey: string): Promise<void> {
    this.serverUrl = serverUrl;
    this.apiKey = apiKey;

    // Normalize URL
    if (!this.serverUrl.startsWith('http')) {
      this.serverUrl = `http://${this.serverUrl}`;
    }

    // Save to storage
    await AsyncStorage.setItem('server_url', this.serverUrl);
    await AsyncStorage.setItem('api_key', this.apiKey);

    this.updateBaseURL();
  }

  /**
   * Clear stored credentials
   */
  async clearCredentials(): Promise<void> {
    this.serverUrl = '';
    this.apiKey = '';
    await AsyncStorage.removeItem('server_url');
    await AsyncStorage.removeItem('api_key');
  }

  /**
   * Get current API key
   */
  getApiKey(): string {
    return this.apiKey;
  }

  /**
   * Get current server URL
   */
  getServerUrl(): string {
    return this.serverUrl;
  }

  /**
   * Update axios base URL and auth headers
   */
  private updateBaseURL(): void {
    this.client.defaults.baseURL = this.serverUrl;

    if (this.apiKey) {
      this.client.defaults.headers.common['Authorization'] = `Bearer ${this.apiKey}`;
    } else {
      delete this.client.defaults.headers.common['Authorization'];
    }
  }

  /**
   * Test connection to server
   */
  async testConnection(): Promise<boolean> {
    try {
      const response = await this.client.get<HealthResponse>('/health');
      return response.status === 200 && response.data.status === 'ok';
    } catch (error) {
      console.error('Connection test failed:', error);
      return false;
    }
  }

  /**
   * Pull changes from server
   * Gets all tasks and lists modified since the given timestamp
   */
  async pull(since: number): Promise<PullResponse> {
    try {
      const request: PullRequest = { since };
      const response = await this.client.post<PullResponse>('/sync/pull', request);
      return response.data;
    } catch (error) {
      console.error('Pull failed:', error);
      throw this.convertError(error);
    }
  }

  /**
   * Push changes to server
   * Sends local tasks and lists to server for synchronization
   */
  async push(data: PushRequest): Promise<void> {
    try {
      await this.client.post<PushResponse>('/sync/push', data);
    } catch (error) {
      console.error('Push failed:', error);
      throw this.convertError(error);
    }
  }

  /**
   * Get all tasks (direct fetch, not sync-based)
   */
  async getTasks(): Promise<Task[]> {
    try {
      const response = await this.client.get<Task[]>('/tasks');
      return response.data;
    } catch (error) {
      console.error('Get tasks failed:', error);
      throw this.convertError(error);
    }
  }

  /**
   * Create a new task
   */
  async createTask(task: Task): Promise<Task> {
    try {
      const response = await this.client.post<Task>('/tasks', task);
      return response.data;
    } catch (error) {
      console.error('Create task failed:', error);
      throw this.convertError(error);
    }
  }

  /**
   * Update an existing task
   */
  async updateTask(id: number, task: Task): Promise<Task> {
    try {
      const response = await this.client.put<Task>(`/tasks/${id}`, task);
      return response.data;
    } catch (error) {
      console.error('Update task failed:', error);
      throw this.convertError(error);
    }
  }

  /**
   * Delete a task (soft delete on server)
   */
  async deleteTask(id: number): Promise<void> {
    try {
      await this.client.delete(`/tasks/${id}`);
    } catch (error) {
      console.error('Delete task failed:', error);
      throw this.convertError(error);
    }
  }

  /**
   * Get all lists (direct fetch, not sync-based)
   */
  async getLists(): Promise<TodoList[]> {
    try {
      const response = await this.client.get<TodoList[]>('/lists');
      return response.data;
    } catch (error) {
      console.error('Get lists failed:', error);
      throw this.convertError(error);
    }
  }

  /**
   * Create a new list
   */
  async createList(list: TodoList): Promise<TodoList> {
    try {
      const response = await this.client.post<TodoList>('/lists', list);
      return response.data;
    } catch (error) {
      console.error('Create list failed:', error);
      throw this.convertError(error);
    }
  }

  /**
   * Update an existing list
   */
  async updateList(id: number, list: TodoList): Promise<TodoList> {
    try {
      const response = await this.client.put<TodoList>(`/lists/${id}`, list);
      return response.data;
    } catch (error) {
      console.error('Update list failed:', error);
      throw this.convertError(error);
    }
  }

  /**
   * Delete a list (soft delete on server)
   */
  async deleteList(id: number): Promise<void> {
    try {
      await this.client.delete(`/lists/${id}`);
    } catch (error) {
      console.error('Delete list failed:', error);
      throw this.convertError(error);
    }
  }

  /**
   * Handle HTTP errors and convert to ApiError
   */
  private handleError(error: AxiosError): Promise<AxiosError> {
    const apiError = this.convertError(error);
    console.error('API Error:', {
      status: apiError.status,
      message: apiError.message,
      details: apiError.details,
    });
    return Promise.reject(error);
  }

  /**
   * Convert axios error to ApiError
   */
  private convertError(error: any): ApiError {
    if (axios.isAxiosError(error)) {
      const status = error.response?.status || 0;
      const message = error.response?.data?.error || error.message || 'Unknown error';
      const details = error.response?.data?.details;

      return {
        status,
        message,
        details,
      };
    }

    return {
      status: 0,
      message: error?.message || 'Unknown error',
    };
  }

  /**
   * Check if an error is a 401 Unauthorized (invalid API key)
   */
  isUnauthorizedError(error: any): boolean {
    if (axios.isAxiosError(error)) {
      return error.response?.status === 401;
    }
    return false;
  }

  /**
   * Check if error is a network error (no connection)
   */
  isNetworkError(error: any): boolean {
    if (axios.isAxiosError(error)) {
      return !error.response || error.code === 'ECONNABORTED' || error.code === 'ENOTFOUND';
    }
    return false;
  }
}

// Export singleton instance
export const apiClient = new ApiClient();
