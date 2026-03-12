package models

import sharedmodels "github.com/billgoswell/commandlinetodo/shared-types/models"

// Re-export from shared-types for backwards compatibility
type User = sharedmodels.User
type TodoList = sharedmodels.TodoList
type Task = sharedmodels.Task
type SyncRequest = sharedmodels.SyncRequest
type SyncResponse = sharedmodels.SyncResponse
type PushRequest = sharedmodels.PushRequest
type PushResponse = sharedmodels.PushResponse
type IDMapping = sharedmodels.IDMapping
type HealthResponse = sharedmodels.HealthResponse
type ErrorResponse = sharedmodels.ErrorResponse
type Change = sharedmodels.Change
type SyncStatus = sharedmodels.SyncStatus
