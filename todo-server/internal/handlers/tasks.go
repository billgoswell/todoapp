package handlers

import (
	"log"
	"strconv"

	"github.com/billgoswell/commandlinetodo-server/internal/db"
	"github.com/billgoswell/commandlinetodo-server/internal/middleware"
	"github.com/billgoswell/commandlinetodo-server/internal/models"
	"github.com/billgoswell/commandlinetodo-server/internal/request"
	"github.com/billgoswell/commandlinetodo-server/internal/response"
	"github.com/billgoswell/commandlinetodo-server/internal/validation"
	"github.com/gin-gonic/gin"
)

// GetTasks retrieves all tasks for the authenticated user
func GetTasks(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		// Get all tasks since timestamp 0 (all time)
		tasks, err := database.GetTasksSince(userID, 0)
		if err != nil {
			log.Printf("ERROR: Failed to fetch tasks: %v", err)
			response.InternalError(c, "failed to fetch tasks", err.Error())
			return
		}

		taskModels := convertTaskMaps(tasks)
		response.OK(c, taskModels)
	}
}

// CreateTask creates a new task
func CreateTask(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var task models.Task
		if err := request.BindJSON(c, &task); err != nil {
			return
		}

		// Validate the task
		if err := validation.ValidateTask(task); err != nil {
			if validationErr, ok := err.(*validation.ValidationError); ok {
				response.ValidationError(c, validationErr.Message())
				return
			}
			response.BadRequest(c, err.Error())
			return
		}

		taskMap := convertTaskToMap(task)
		taskMap["user_id"] = userID
		id, err := database.UpsertTask(userID, taskMap)
		if err != nil {
			log.Printf("ERROR: UpsertTask failed: %v", err)
			response.InternalError(c, "failed to create task", err.Error())
			return
		}

		task.ID = id
		response.Created(c, task)
	}
}

// UpdateTask updates an existing task
func UpdateTask(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		taskID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "invalid task id")
			return
		}

		var task models.Task
		if err := request.BindJSON(c, &task); err != nil {
			return
		}

		// Validate the task
		if err := validation.ValidateTask(task); err != nil {
			if validationErr, ok := err.(*validation.ValidationError); ok {
				response.ValidationError(c, validationErr.Message())
				return
			}
			response.BadRequest(c, err.Error())
			return
		}

		task.ID = taskID
		taskMap := convertTaskToMap(task)
		taskMap["user_id"] = userID
		_, err = database.UpsertTask(userID, taskMap)
		if err != nil {
			log.Printf("ERROR: UpsertTask failed on update: %v", err)
			response.InternalError(c, "failed to update task", err.Error())
			return
		}

		response.OK(c, task)
	}
}

// DeleteTask deletes a task
func DeleteTask(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		taskID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "invalid task id")
			return
		}

		// For soft delete, we need to fetch the task and mark it as deleted
		tasks, err := database.GetTasksSince(userID, 0)
		if err != nil {
			log.Printf("ERROR: Failed to fetch task: %v", err)
			response.InternalError(c, "failed to fetch task", err.Error())
			return
		}

		var targetTask map[string]any
		for _, t := range tasks {
			if t["id"].(int) == taskID {
				targetTask = t
				break
			}
		}

		if targetTask == nil {
			response.NotFound(c, "task not found")
			return
		}

		// Mark as deleted
		targetTask["deleted"] = true
		_, err = database.UpsertTask(userID, targetTask)
		if err != nil {
			log.Printf("ERROR: Failed to delete task: %v", err)
			response.InternalError(c, "failed to delete task", err.Error())
			return
		}

		response.NoContent(c)
	}
}
