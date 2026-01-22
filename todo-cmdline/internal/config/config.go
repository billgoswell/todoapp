package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/billgoswell/commandlinetodo/internal/logger"
	"github.com/google/uuid"
)

type Config struct {
	DBPath   string     `json:"db_path"`
	LogLevel string     `json:"log_level"`
	Sync     SyncConfig `json:"sync"`
}

type SyncConfig struct {
	Enabled             bool   `json:"enabled"`
	ServerURL           string `json:"server_url"`
	APIKey              string `json:"api_key"`
	DeviceID            string `json:"device_id"`
	SyncIntervalSeconds int    `json:"sync_interval_seconds"`
	AutoSyncOnChange    bool   `json:"auto_sync_on_change"`
	RetryAttempts       int    `json:"retry_attempts"`
	TimeoutSeconds      int    `json:"timeout_seconds"`
}

// Config file location
const ConfigDir = ".config/todo"
const configFileName = "config"

// defaultDBPath is the default location for the database file
const defaultDBPath = "./todo.db"

// Default sync configuration values
const (
	defaultSyncInterval    = 60
	defaultRetryAttempts   = 3
	defaultTimeoutSeconds  = 10
)

// LoadConfig loads the application configuration from file or returns defaults
func LoadConfig() Config {
	// Get config file path
	configPath, err := getConfigPath()
	if err != nil {
		logger.LogError("determine config file path", "error", err)
		cfg := getDefaultConfig()
		ensureDBDirectory(cfg.DBPath)
		return cfg
	}

	// Ensure config directory exists
	if err := ensureConfigDirectory(configPath); err != nil {
		logger.LogError("create config directory", "error", err)
		cfg := getDefaultConfig()
		ensureDBDirectory(cfg.DBPath)
		return cfg
	}

	// Load config from file
	cfg, err := loadConfigFromFile(configPath)
	if err != nil {
		logger.LogError("load config file", "error", err)
		cfg = getDefaultConfig()
	}

	// Validate and apply defaults
	validateAndApplyDefaults(&cfg)

	// Ensure DB directory exists
	if err := ensureDBDirectory(cfg.DBPath); err != nil {
		logger.LogError("create database directory", "error", err)
	}

	return cfg
}

// GetConfigDir returns the config directory path
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ConfigDir), nil
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ConfigDir, configFileName), nil
}

func ensureConfigDirectory(configPath string) error {
	dir := filepath.Dir(configPath)

	if _, err := os.Stat(dir); err == nil {
		return nil
	}

	return os.MkdirAll(dir, 0755)
}

func getDefaultConfig() Config {
	return Config{
		DBPath:   defaultDBPath,
		LogLevel: "info",
		Sync: SyncConfig{
			Enabled:             false,
			ServerURL:           "",
			APIKey:              "",
			DeviceID:            generateClientID(),
			SyncIntervalSeconds: defaultSyncInterval,
			AutoSyncOnChange:    true,
			RetryAttempts:       defaultRetryAttempts,
			TimeoutSeconds:      defaultTimeoutSeconds,
		},
	}
}

func createDefaultConfigFile(configPath string) error {
	cfg := getDefaultConfig()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

func loadConfigFromFile(configPath string) (Config, error) {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config file if it doesn't exist
		if err := createDefaultConfigFile(configPath); err != nil {
			logger.LogError("create default config file", "error", err)
		}
		return getDefaultConfig(), nil
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	// Unmarshal JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validateAndApplyDefaults(cfg *Config) {
	// Apply default DB path if empty
	if cfg.DBPath == "" {
		cfg.DBPath = defaultDBPath
	}

	// Validate and apply default log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[cfg.LogLevel] {
		cfg.LogLevel = "info"
	}

	// Apply default sync interval if invalid
	if cfg.Sync.SyncIntervalSeconds <= 0 {
		cfg.Sync.SyncIntervalSeconds = defaultSyncInterval
	}

	// Apply default retry attempts if invalid
	if cfg.Sync.RetryAttempts <= 0 {
		cfg.Sync.RetryAttempts = defaultRetryAttempts
	}

	// Apply default timeout if invalid
	if cfg.Sync.TimeoutSeconds <= 0 {
		cfg.Sync.TimeoutSeconds = defaultTimeoutSeconds
	}

	// Auto-generate device ID if empty
	if cfg.Sync.DeviceID == "" {
		cfg.Sync.DeviceID = generateClientID()
	}
}

func ensureDBDirectory(dbPath string) error {
	dir := filepath.Dir(dbPath)

	if dir == "." || dir == "" {
		return nil
	}

	if _, err := os.Stat(dir); err == nil {
		return nil
	}

	return os.MkdirAll(dir, 0755)
}

// generateClientID creates a UUID for client identification
func generateClientID() string {
	return uuid.New().String()
}
