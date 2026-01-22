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

// GetLists retrieves all lists for the authenticated user
func GetLists(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		// Get all lists since timestamp 0 (all time)
		lists, err := database.GetListsSince(userID, 0)
		if err != nil {
			log.Printf("ERROR: Failed to fetch lists: %v", err)
			response.InternalError(c, "failed to fetch lists", err.Error())
			return
		}

		listModels := convertListMaps(lists)
		response.OK(c, listModels)
	}
}

// CreateList creates a new list
func CreateList(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var list models.TodoList
		if err := request.BindJSON(c, &list); err != nil {
			return
		}

		// Validate the list
		if err := validation.ValidateList(list); err != nil {
			if validationErr, ok := err.(*validation.ValidationError); ok {
				response.ValidationError(c, validationErr.Message())
				return
			}
			response.BadRequest(c, err.Error())
			return
		}

		listMap := convertListToMap(list)
		listMap["user_id"] = userID
		id, err := database.UpsertList(userID, listMap)
		if err != nil {
			log.Printf("ERROR: UpsertList failed: %v", err)
			response.InternalError(c, "failed to create list", err.Error())
			return
		}

		list.ID = id
		response.Created(c, list)
	}
}

// UpdateList updates an existing list
func UpdateList(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		listID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "invalid list id")
			return
		}

		var list models.TodoList
		if err := request.BindJSON(c, &list); err != nil {
			return
		}

		// Validate the list
		if err := validation.ValidateList(list); err != nil {
			if validationErr, ok := err.(*validation.ValidationError); ok {
				response.ValidationError(c, validationErr.Message())
				return
			}
			response.BadRequest(c, err.Error())
			return
		}

		list.ID = listID
		listMap := convertListToMap(list)
		listMap["user_id"] = userID
		_, err = database.UpsertList(userID, listMap)
		if err != nil {
			log.Printf("ERROR: UpsertList failed on update: %v", err)
			response.InternalError(c, "failed to update list", err.Error())
			return
		}

		response.OK(c, list)
	}
}

// DeleteList deletes a list
func DeleteList(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		listID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "invalid list id")
			return
		}

		// For soft delete, we need to fetch the list and mark it as archived
		lists, err := database.GetListsSince(userID, 0)
		if err != nil {
			log.Printf("ERROR: Failed to fetch list: %v", err)
			response.InternalError(c, "failed to fetch list", err.Error())
			return
		}

		var targetList map[string]any
		for _, l := range lists {
			if l["id"].(int) == listID {
				targetList = l
				break
			}
		}

		if targetList == nil {
			response.NotFound(c, "list not found")
			return
		}

		// Mark as archived instead of deleting
		targetList["archived"] = true
		_, err = database.UpsertList(userID, targetList)
		if err != nil {
			log.Printf("ERROR: Failed to delete list: %v", err)
			response.InternalError(c, "failed to delete list", err.Error())
			return
		}

		response.NoContent(c)
	}
}
