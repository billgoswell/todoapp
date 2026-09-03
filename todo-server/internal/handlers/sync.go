package handlers

import (
	"log"
	"time"

	"github.com/billgoswell/commandlinetodo-server/internal/db"
	"github.com/billgoswell/commandlinetodo-server/internal/middleware"
	"github.com/billgoswell/commandlinetodo-server/internal/models"
	"github.com/billgoswell/commandlinetodo-server/internal/request"
	"github.com/billgoswell/commandlinetodo-server/internal/response"
	"github.com/gin-gonic/gin"
)

// Pull retrieves all tasks and lists modified since the given timestamp
func Pull(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var req models.SyncRequest
		if err := request.BindJSON(c, &req); err != nil {
			return
		}

		// Get tasks and lists modified since the timestamp
		tasks, err := database.GetTasksSince(userID, req.Since)
		if err != nil {
			log.Printf("ERROR: Failed to fetch tasks: %v", err)
			response.InternalError(c, "failed to fetch tasks", err.Error())
			return
		}

		lists, err := database.GetListsSince(userID, req.Since)
		if err != nil {
			log.Printf("ERROR: Failed to fetch lists: %v", err)
			response.InternalError(c, "failed to fetch lists", err.Error())
			return
		}

		// Convert map responses to proper model structs
		taskModels := convertTaskMaps(tasks)
		listModels := convertListMaps(lists)

		response.OK(c, models.SyncResponse{
			Tasks: taskModels,
			Lists: listModels,
		})
	}
}

// Push accepts client changes and merges them into the server database
func Push(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var req models.PushRequest
		if err := request.BindJSON(c, &req); err != nil {
			return
		}

		var listIDMappings []models.IDMapping
		var taskIDMappings []models.IDMapping

		// Build mapping from list clientID to server ID for task foreign key resolution
		listClientIDToServerID := make(map[string]int)

		// Upsert all lists first (before tasks due to foreign key constraint)
		// Upsert all lists in a single batch operation for better performance
		if len(req.Lists) > 0 {
			listMaps := make([]map[string]any, len(req.Lists))
			for i, list := range req.Lists {
				listMap := convertListToMap(list)
				listMap["user_id"] = userID
				listMaps[i] = listMap
			}

			// Upsert all lists in a single batch operation and capture ID mappings
			mappings, err := database.UpsertListsBatch(userID, listMaps)
			if err != nil {
				log.Printf("ERROR: Failed to sync lists: %v", err)
				response.InternalError(c, "failed to sync lists", err.Error())
				return
			}
			// Convert db.IDMapping to models.IDMapping with entity type
			for _, mapping := range mappings {
				listIDMappings = append(listIDMappings, models.IDMapping{
					ClientID:   mapping.ClientID,
					ServerID:   mapping.ServerID,
					EntityType: "list",
				})
				// Build clientID -> serverID mapping for task FK resolution
				listClientIDToServerID[mapping.ClientID] = mapping.ServerID
			}
		}

		// Convert task models to maps and prepare for batch upsert
		if len(req.Tasks) > 0 {
			taskMaps := make([]map[string]any, 0, len(req.Tasks))
			for _, task := range req.Tasks {
				taskMap := convertTaskToMap(task)
				taskMap["user_id"] = userID

				// NEW: Resolve TodoListClientID to server ID for database FK
				if task.TodoListClientID != "" {
					if serverID, ok := listClientIDToServerID[task.TodoListClientID]; ok {
						// We just synced this list, use the server ID
						taskMap["todo_list_id"] = serverID
					} else {
						// List wasn't in this push, try to look it up in the database
						serverID, err := database.GetListIDByClientID(userID, task.TodoListClientID)
						if err != nil {
							log.Printf("WARN: Cannot resolve list for task %s: %v", task.ClientID, err)
							continue // Skip this task to avoid FK violation
						}
						taskMap["todo_list_id"] = serverID
					}
				}
				// else: use existing TodoListID (backward compatibility)

				taskMaps = append(taskMaps, taskMap)
			}

			// Upsert all tasks in a single batch operation and capture ID mappings
			mappings, err := database.UpsertTasksBatch(userID, taskMaps)
			if err != nil {
				log.Printf("ERROR: Failed to sync tasks: %v", err)
				response.InternalError(c, "failed to sync tasks", err.Error())
				return
			}
			// Convert db.IDMapping to models.IDMapping with entity type
			for _, mapping := range mappings {
				taskIDMappings = append(taskIDMappings, models.IDMapping{
					ClientID:   mapping.ClientID,
					ServerID:   mapping.ServerID,
					EntityType: "task",
				})
			}
		}

		// Return push response with ID mappings
		response.OK(c, models.PushResponse{
			Status:           "ok",
			ListIDMappings:   listIDMappings,
			TaskIDMappings:   taskIDMappings,
		})
	}
}

// Helper functions to convert between maps and models

// optionalUnix converts an optional timestamp from the repository layer
// (*time.Time or *int64) to a *int64 unix timestamp, nil if unset
func optionalUnix(val any) *int64 {
	switch v := val.(type) {
	case *time.Time:
		if v == nil {
			return nil
		}
		unix := v.Unix()
		return &unix
	case *int64:
		return v
	case int64:
		if v == 0 {
			return nil
		}
		return &v
	default:
		return nil
	}
}

func convertTaskMaps(tasks []map[string]any) []models.Task {
	var result []models.Task
	for _, t := range tasks {
		task := models.Task{
			ID:         t["id"].(int),
			ClientID:   t["client_id"].(string),
			TodoListID: t["todo_list_id"].(int),
			Todo:       t["todo"].(string),
			Priority:   t["priority"].(int),
			Done:       t["done"].(bool),
			Deleted:    t["deleted"].(bool),
			CreatedAt:  t["created_at"].(int64),
			UpdatedAt:  t["updated_at"].(int64),
			Version:    t["version"].(int),
		}

		// Handle optional timestamps
		task.DateCompleted = optionalUnix(t["date_completed"])
		task.DueDate = optionalUnix(t["due_date"])
		task.DeletedAt = optionalUnix(t["deleted_at"])
		if da := t["date_added"]; da != nil {
			if daVal, ok := da.(int64); ok {
				task.DateAdded = daVal
			}
		}

		result = append(result, task)
	}
	return result
}

func convertListMaps(lists []map[string]any) []models.TodoList {
	var result []models.TodoList
	for _, l := range lists {
		result = append(result, models.TodoList{
			ID:           l["id"].(int),
			ClientID:     l["client_id"].(string),
			Name:         l["name"].(string),
			DisplayOrder: l["display_order"].(int),
			Archived:     l["archived"].(bool),
			CreatedAt:    l["created_at"].(int64),
			UpdatedAt:    l["updated_at"].(int64),
			Version:      l["version"].(int),
		})
	}
	return result
}

func convertTaskToMap(task models.Task) map[string]any {
	return map[string]any{
		"client_id":            task.ClientID,
		"todo_list_id":         task.TodoListID,
		"todo_list_client_id":  task.TodoListClientID,
		"todo":                 task.Todo,
		"priority":             task.Priority,
		"done":                 task.Done,
		"date_added":           task.DateAdded,
		"date_completed":       task.DateCompleted,
		"due_date":             task.DueDate,
		"deleted":              task.Deleted,
		"deleted_at":           task.DeletedAt,
		"updated_at":           task.UpdatedAt,
		"version":              task.Version,
	}
}

func convertListToMap(list models.TodoList) map[string]any {
	return map[string]any{
		"client_id":    list.ClientID,
		"name":         list.Name,
		"display_order": list.DisplayOrder,
		"archived":     list.Archived,
		"updated_at":   list.UpdatedAt,
		"version":      list.Version,
	}
}
