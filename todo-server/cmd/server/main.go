package main

import (
	"fmt"
	"log"
	"os"

	"github.com/billgoswell/commandlinetodo-server/internal/db"
	"github.com/billgoswell/commandlinetodo-server/internal/handlers"
	"github.com/billgoswell/commandlinetodo-server/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration from environment variables
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "commandlinetodo"
	}

	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		serverHost = "127.0.0.1"
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	// Initialize database
	database, err := db.NewDB(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create router
	router := gin.Default()

	// Health endpoint (no auth required)
	router.GET("/api/v1/health", handlers.Health)

	// API routes (auth required)
	authorized := router.Group("/api/v1")
	authorized.Use(middleware.AuthMiddleware(database))
	{
		// Sync endpoints
		authorized.POST("/sync/pull", handlers.Pull(database))
		authorized.POST("/sync/push", handlers.Push(database))

		// Task endpoints
		authorized.GET("/tasks", handlers.GetTasks(database))
		authorized.POST("/tasks", handlers.CreateTask(database))
		authorized.PUT("/tasks/:id", handlers.UpdateTask(database))
		authorized.DELETE("/tasks/:id", handlers.DeleteTask(database))

		// List endpoints
		authorized.GET("/lists", handlers.GetLists(database))
		authorized.POST("/lists", handlers.CreateList(database))
		authorized.PUT("/lists/:id", handlers.UpdateList(database))
		authorized.DELETE("/lists/:id", handlers.DeleteList(database))
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", serverHost, serverPort)
	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
