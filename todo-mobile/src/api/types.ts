/**
 * TypeScript types for API requests and responses
 *
 * Matches the server API contract exactly
 */

export interface Task {
  id?: number;
  client_id: string;
  todo_list_id: number;
  todo: string;
  priority: number;
  done: boolean;
  date_added: number;
  date_completed?: number | null;
  due_date?: number | null;
  deleted: boolean;
  deleted_at?: number | null;
  created_at?: number;
  updated_at: number;
  version: number;
}

export interface TodoList {
  id?: number;
  client_id: string;
  name: string;
  display_order: number;
  archived: boolean;
  created_at?: number;
  updated_at: number;
  version: number;
}

export interface PullRequest {
  since: number;
}

export interface PullResponse {
  tasks: Task[];
  lists: TodoList[];
}

export interface PushRequest {
  tasks: Task[];
  lists: TodoList[];
}

export interface PushResponse {
  status: string;
}

export interface HealthResponse {
  status: string;
}

export interface SyncResult {
  success: boolean;
  error?: Error;
}

export interface ApiError {
  status: number;
  message: string;
  details?: string;
}
