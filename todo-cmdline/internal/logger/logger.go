package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	maxLogSize = 10 * 1024 * 1024 // 10MB
	logDir     = "logs"
	logFile    = "todo.log"
)

var logger *slog.Logger

// LogLevel represents configurable log levels
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// InitLogger sets up the logging system
// configDir should be the config directory path (e.g., ~/.config/todo)
func InitLogger(configDir string, level LogLevel) error {
	logDirPath := filepath.Join(configDir, logDir)
	logFilePath := filepath.Join(logDirPath, logFile)

	// Create log directory
	if err := os.MkdirAll(logDirPath, 0755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	// Check if rotation is needed
	if err := rotateLogIfNeeded(logFilePath); err != nil {
		return fmt.Errorf("rotate log: %w", err)
	}

	// Open log file
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	// Parse log level
	var slogLevel slog.Level
	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	// Create handler with JSON formatting
	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slogLevel,
	})

	logger = slog.New(handler)
	logger.Info("Logger initialized", "level", string(level))

	return nil
}

// rotateLogIfNeeded checks file size and rotates if needed
func rotateLogIfNeeded(logFilePath string) error {
	info, err := os.Stat(logFilePath)
	if os.IsNotExist(err) {
		return nil // No rotation needed for new file
	}
	if err != nil {
		return err
	}

	// Check if file exceeds max size
	if info.Size() >= maxLogSize {
		oldPath := logFilePath + ".old"

		// Remove old backup if exists
		os.Remove(oldPath)

		// Rename current log to .old
		if err := os.Rename(logFilePath, oldPath); err != nil {
			return err
		}
	}

	return nil
}

// LogDebug logs a debug-level message
func LogDebug(msg string, args ...any) {
	if logger != nil {
		logger.Debug(msg, args...)
	}
}

// LogInfo logs an info-level message
func LogInfo(msg string, args ...any) {
	if logger != nil {
		logger.Info(msg, args...)
	}
}

// LogWarn logs a warning-level message
func LogWarn(msg string, args ...any) {
	if logger != nil {
		logger.Warn(msg, args...)
	}
}

// LogError logs an error-level message
func LogError(msg string, args ...any) {
	if logger != nil {
		logger.Error(msg, args...)
	}
}

// LogFatal logs an error and exits the program
func LogFatal(msg string, args ...any) {
	if logger != nil {
		logger.Error(msg, args...)
	}
	fmt.Fprintf(os.Stderr, "FATAL: %s\n", msg)
	os.Exit(1)
}
