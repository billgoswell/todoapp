package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standardized API response
type Response struct {
	Status string     `json:"status"`
	Data   any        `json:"data,omitempty"`
	Error  *ErrorInfo `json:"error,omitempty"`
}

// ErrorInfo represents error details in the response
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Success returns a successful response with data
func Success(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, Response{
		Status: "success",
		Data:   data,
	})
}

// Error returns an error response
func Error(c *gin.Context, statusCode int, code, message string, details ...string) {
	var detailStr string
	if len(details) > 0 {
		detailStr = details[0]
	}
	c.JSON(statusCode, Response{
		Status: "error",
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: detailStr,
		},
	})
}

// OK returns a 200 OK response with data
func OK(c *gin.Context, data interface{}) {
	Success(c, http.StatusOK, data)
}

// Created returns a 201 Created response with data
func Created(c *gin.Context, data interface{}) {
	Success(c, http.StatusCreated, data)
}

// NoContent returns a 204 No Content response
func NoContent(c *gin.Context) {
	c.String(http.StatusNoContent, "")
}

// BadRequest returns a 400 Bad Request error response
func BadRequest(c *gin.Context, message string, details ...string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message, details...)
}

// Unauthorized returns a 401 Unauthorized error response
func Unauthorized(c *gin.Context, message string, details ...string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message, details...)
}

// Forbidden returns a 403 Forbidden error response
func Forbidden(c *gin.Context, message string, details ...string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message, details...)
}

// NotFound returns a 404 Not Found error response
func NotFound(c *gin.Context, message string, details ...string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message, details...)
}

// Conflict returns a 409 Conflict error response
func Conflict(c *gin.Context, message string, details ...string) {
	Error(c, http.StatusConflict, "CONFLICT", message, details...)
}

// InternalError returns a 500 Internal Server Error response
func InternalError(c *gin.Context, message string, details ...string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message, details...)
}

// ValidationError returns a 400 Bad Request response with validation error code
func ValidationError(c *gin.Context, message string, details ...string) {
	Error(c, http.StatusBadRequest, "VALIDATION_ERROR", message, details...)
}
