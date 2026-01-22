package handlers

import (
	"net/http"

	"github.com/billgoswell/commandlinetodo-server/internal/models"
	"github.com/gin-gonic/gin"
)

// Health returns the server health status
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, models.HealthResponse{
		Status: "ok",
	})
}
