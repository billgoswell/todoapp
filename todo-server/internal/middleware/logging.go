package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogging logs incoming HTTP requests with method, path, and response time
func RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Log request details
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()
		log.Printf("[%s] %s %s - %d (%v)", method, path, c.Request.RemoteAddr, statusCode, duration)
	}
}

// ErrorLogging logs errors that occur during request processing
func ErrorLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Log errors if any occurred
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Printf("ERROR: %v (type: %v)", err.Error(), err.Type)
			}
		}
	}
}

// RecoveryWithLogging recovers from panics and logs them
func RecoveryWithLogging() gin.HandlerFunc {
	return gin.RecoveryWithWriter(nil) // Use default logger
}
