package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/billgoswell/commandlinetodo-server/internal/middleware"
	"github.com/gin-gonic/gin"
)

// ============================================================================
// AUTH MIDDLEWARE TESTS
// ============================================================================

// TestAuth_ValidAPIKey tests that valid API keys are accepted
func TestAuth_ValidAPIKey(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	apiKey := "test-api-key-valid"
	_ = ltdb.CreateTestUser(t, apiKey) // User created for auth validation

	router := gin.Default()
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(ltdb.GetDB()))
	{
		protected.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
	}

	req := httptest.NewRequest("GET", "/api/v1/protected", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "success" {
		t.Errorf("Expected 'success', got '%v'", response["message"])
	}
}

// TestAuth_MissingAuthHeader tests that missing auth header is rejected
func TestAuth_MissingAuthHeader(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	router := gin.Default()
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(ltdb.GetDB()))
	{
		protected.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
	}

	req := httptest.NewRequest("GET", "/api/v1/protected", nil)
	// No Authorization header

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if _, ok := response["error"]; !ok {
		t.Error("Expected error field in response")
	}
}

// TestAuth_InvalidAuthHeader tests that malformed headers are rejected
func TestAuth_InvalidAuthHeader(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	router := gin.Default()
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(ltdb.GetDB()))
	{
		protected.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
	}

	testCases := []string{
		"InvalidToken",           // Missing "Bearer"
		"Bearer",                 // Missing token
		"Bearer ",                // Empty token
		"Basic dGVzdDp0ZXN0",    // Wrong scheme
		"Bearer token with spaces", // Token with spaces
	}

	for _, authHeader := range testCases {
		req := httptest.NewRequest("GET", "/api/v1/protected", nil)
		req.Header.Set("Authorization", authHeader)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Header '%s': Expected 401, got %d", authHeader, w.Code)
		}
	}
}

// TestAuth_InvalidAPIKey tests that invalid API keys are rejected
func TestAuth_InvalidAPIKey(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	router := gin.Default()
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(ltdb.GetDB()))
	{
		protected.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
	}

	req := httptest.NewRequest("GET", "/api/v1/protected", nil)
	req.Header.Set("Authorization", "Bearer non-existent-key-12345")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid key, got %d", w.Code)
	}
}

// TestAuth_ContextPopulation tests that user context is set correctly
func TestAuth_ContextPopulation(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	apiKey := "test-api-key-context"
	expectedUserID := ltdb.CreateTestUser(t, apiKey)

	router := gin.Default()
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(ltdb.GetDB()))
	{
		protected.GET("/protected", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not in context"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"user_id": userID,
				"message": "context populated",
			})
		})
	}

	req := httptest.NewRequest("GET", "/api/v1/protected", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["user_id"] != float64(expectedUserID) {
		t.Errorf("Expected user_id %d, got %v", expectedUserID, response["user_id"])
	}

	if response["message"] != "context populated" {
		t.Error("Context was not properly populated")
	}
}

// TestAuth_ProtectsEndpoint tests that auth middleware protects endpoints
func TestAuth_ProtectsEndpoint(t *testing.T) {
	config := GetTestConfig()
	if !config.UseLocalServer {
		t.Skip("Skipping local test - using live server")
	}

	ltdb := SetupLocalTestDB(t)
	defer ltdb.Cleanup()

	router := gin.Default()
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(ltdb.GetDB()))
	{
		protected.GET("/tasks", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"tasks": []interface{}{}})
		})
	}

	// Unprotected request should be rejected
	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d - endpoint is not protected", w.Code)
	}
}
