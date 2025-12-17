package server

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// Config holds all server configuration.
type Config struct {
	// Server settings
	Port string `json:"port"`
	Host string `json:"host"`

	// Storage settings
	StoreFileName string `json:"store_file"`

	// Auth settings
	AuthPassword string `json:"auth_password"`

	// Logging settings
	LogLevel string `json:"log_level"` // debug, info, warn, error
	LogJSON  bool   `json:"log_json"`

	// Snapshot settings
	SnapshotInterval string `json:"snapshot_interval"` // e.g., "1s", "5s", "1m"
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Port:             "8080",
		Host:             "localhost",
		StoreFileName:    "./data/bookmarks.json",
		AuthPassword:     "", // Empty means no auth required
		LogLevel:         "info",
		LogJSON:          false,
		SnapshotInterval: "1s",
	}
}

// LoadConfig loads configuration from multiple sources with precedence:
// CLI flags > Environment variables > Config file > Defaults
func LoadConfig(args []string) (Config, error) {
	cfg := DefaultConfig()

	// Create FlagSet with all flags
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	configFile := fs.String("config", "", "Path to config file (JSON)")
	port := fs.String("port", cfg.Port, "Server port")
	host := fs.String("host", cfg.Host, "Server host")
	storeFile := fs.String("store-file", cfg.StoreFileName, "Path to bookmarks storage file")
	password := fs.String("password", cfg.AuthPassword, "Authentication password (empty = no auth)")
	logLevel := fs.String("log-level", cfg.LogLevel, "Log level (debug, info, warn, error)")
	logJSON := fs.Bool("log-json", cfg.LogJSON, "Output logs as JSON")
	snapshotInterval := fs.String("snapshot-interval", cfg.SnapshotInterval, "Snapshot save interval (e.g., 1s, 5s, 1m)")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}

	// Track which flags were explicitly set
	explicitFlags := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		explicitFlags[f.Name] = true
	})

	// 1. Load from config file if specified
	if *configFile != "" {
		if err := loadConfigFile(&cfg, *configFile); err != nil {
			return cfg, fmt.Errorf("loading config file: %w", err)
		}
	}

	// 2. Apply environment variables (override config file)
	if v := os.Getenv("FAVE_PORT"); v != "" {
		cfg.Port = v
	}
	if v := os.Getenv("FAVE_HOST"); v != "" {
		cfg.Host = v
	}
	if v := os.Getenv("FAVE_STORE_FILE"); v != "" {
		cfg.StoreFileName = v
	}
	if v := os.Getenv("FAVE_AUTH_PASSWORD"); v != "" {
		cfg.AuthPassword = v
	}
	if v := os.Getenv("FAVE_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("FAVE_LOG_JSON"); v == "true" {
		cfg.LogJSON = true
	}
	if v := os.Getenv("FAVE_SNAPSHOT_INTERVAL"); v != "" {
		cfg.SnapshotInterval = v
	}

	// 3. Apply CLI flags (highest precedence) - only if explicitly set
	if explicitFlags["port"] {
		cfg.Port = *port
	}
	if explicitFlags["host"] {
		cfg.Host = *host
	}
	if explicitFlags["store-file"] {
		cfg.StoreFileName = *storeFile
	}
	if explicitFlags["password"] {
		cfg.AuthPassword = *password
	}
	if explicitFlags["log-level"] {
		cfg.LogLevel = *logLevel
	}
	if explicitFlags["log-json"] {
		cfg.LogJSON = *logJSON
	}
	if explicitFlags["snapshot-interval"] {
		cfg.SnapshotInterval = *snapshotInterval
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("port cannot be empty")
	}
	if c.StoreFileName == "" {
		return fmt.Errorf("store file name cannot be empty")
	}

	// Validate log level
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
		// Valid
	default:
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.LogLevel)
	}

	return nil
}

// LogLevelValue returns the slog.Level for the configured log level.
func (c Config) LogLevelValue() slog.Level {
	switch c.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Addr returns the full address for the server to listen on.
func (c Config) Addr() string {
	return c.Host + ":" + c.Port
}

// loadConfigFile loads config from a JSON file.
func loadConfigFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	ext := filepath.Ext(path)
	if ext != ".json" {
		return fmt.Errorf("unsupported config file format: %s (use .json)", ext)
	}

	return json.Unmarshal(data, cfg)
}
