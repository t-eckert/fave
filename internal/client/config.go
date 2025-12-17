package client

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config holds the client configuration.
type Config struct {
	Host          string
	Password      string
	Timeout       time.Duration
	DialTimeout   time.Duration
	KeepAlive     time.Duration
	RetryAttempts int
	RetryDelay    time.Duration
	RetryMaxDelay time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Host:          "http://localhost:8080",
		Password:      "",
		Timeout:       30 * time.Second,
		DialTimeout:   10 * time.Second,
		KeepAlive:     30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
		RetryMaxDelay: 10 * time.Second,
	}
}

// LoadConfig loads configuration from multiple sources with precedence:
// CLI flags > environment variables > config file > defaults.
func LoadConfig(args []string) (Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Load from config file if it exists
	if err := loadConfigFile(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to load config file: %w", err)
	}

	// Override with environment variables
	loadFromEnv(&cfg)

	// Override with CLI flags
	if err := loadFromFlags(&cfg, args); err != nil {
		return Config{}, fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if c.DialTimeout <= 0 {
		return fmt.Errorf("dial_timeout must be positive")
	}
	if c.KeepAlive < 0 {
		return fmt.Errorf("keep_alive cannot be negative")
	}
	if c.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts cannot be negative")
	}
	if c.RetryDelay < 0 {
		return fmt.Errorf("retry_delay cannot be negative")
	}
	if c.RetryMaxDelay < 0 {
		return fmt.Errorf("retry_max_delay cannot be negative")
	}
	return nil
}

// loadConfigFile loads configuration from ~/.fave/client-config.json if it exists.
func loadConfigFile(cfg *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get home dir, just skip config file
		return nil
	}

	configPath := filepath.Join(homeDir, ".fave", "client-config.json")

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// No config file, not an error
		return nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var fileConfig struct {
		Host          string `json:"host,omitempty"`
		Password      string `json:"password,omitempty"`
		Timeout       string `json:"timeout,omitempty"`
		DialTimeout   string `json:"dial_timeout,omitempty"`
		KeepAlive     string `json:"keep_alive,omitempty"`
		RetryAttempts int    `json:"retry_attempts,omitempty"`
		RetryDelay    string `json:"retry_delay,omitempty"`
		RetryMaxDelay string `json:"retry_max_delay,omitempty"`
	}

	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply non-empty values
	if fileConfig.Host != "" {
		cfg.Host = fileConfig.Host
	}
	if fileConfig.Password != "" {
		cfg.Password = fileConfig.Password
	}
	if fileConfig.Timeout != "" {
		d, err := time.ParseDuration(fileConfig.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout: %w", err)
		}
		cfg.Timeout = d
	}
	if fileConfig.DialTimeout != "" {
		d, err := time.ParseDuration(fileConfig.DialTimeout)
		if err != nil {
			return fmt.Errorf("invalid dial_timeout: %w", err)
		}
		cfg.DialTimeout = d
	}
	if fileConfig.KeepAlive != "" {
		d, err := time.ParseDuration(fileConfig.KeepAlive)
		if err != nil {
			return fmt.Errorf("invalid keep_alive: %w", err)
		}
		cfg.KeepAlive = d
	}
	if fileConfig.RetryAttempts > 0 {
		cfg.RetryAttempts = fileConfig.RetryAttempts
	}
	if fileConfig.RetryDelay != "" {
		d, err := time.ParseDuration(fileConfig.RetryDelay)
		if err != nil {
			return fmt.Errorf("invalid retry_delay: %w", err)
		}
		cfg.RetryDelay = d
	}
	if fileConfig.RetryMaxDelay != "" {
		d, err := time.ParseDuration(fileConfig.RetryMaxDelay)
		if err != nil {
			return fmt.Errorf("invalid retry_max_delay: %w", err)
		}
		cfg.RetryMaxDelay = d
	}

	return nil
}

// loadFromEnv loads configuration from environment variables.
func loadFromEnv(cfg *Config) {
	if v := os.Getenv("FAVE_HOST"); v != "" {
		cfg.Host = v
	}
	if v := os.Getenv("FAVE_PASSWORD"); v != "" {
		cfg.Password = v
	}
	if v := os.Getenv("FAVE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Timeout = d
		}
	}
	if v := os.Getenv("FAVE_DIAL_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.DialTimeout = d
		}
	}
	if v := os.Getenv("FAVE_KEEP_ALIVE"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.KeepAlive = d
		}
	}
	if v := os.Getenv("FAVE_RETRY_ATTEMPTS"); v != "" {
		var attempts int
		if _, err := fmt.Sscanf(v, "%d", &attempts); err == nil {
			cfg.RetryAttempts = attempts
		}
	}
	if v := os.Getenv("FAVE_RETRY_DELAY"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.RetryDelay = d
		}
	}
	if v := os.Getenv("FAVE_RETRY_MAX_DELAY"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.RetryMaxDelay = d
		}
	}
}

// loadFromFlags loads configuration from CLI flags.
func loadFromFlags(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("client", flag.ContinueOnError)

	// Define flags
	host := fs.String("host", cfg.Host, "Server URL")
	password := fs.String("password", cfg.Password, "Authentication password")
	timeout := fs.Duration("timeout", cfg.Timeout, "Request timeout")
	dialTimeout := fs.Duration("dial-timeout", cfg.DialTimeout, "Connection dial timeout")
	keepAlive := fs.Duration("keep-alive", cfg.KeepAlive, "Keep-alive duration")
	retryAttempts := fs.Int("retry-attempts", cfg.RetryAttempts, "Number of retry attempts")
	retryDelay := fs.Duration("retry-delay", cfg.RetryDelay, "Initial retry delay")
	retryMaxDelay := fs.Duration("retry-max-delay", cfg.RetryMaxDelay, "Maximum retry delay")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Apply flags
	cfg.Host = *host
	cfg.Password = *password
	cfg.Timeout = *timeout
	cfg.DialTimeout = *dialTimeout
	cfg.KeepAlive = *keepAlive
	cfg.RetryAttempts = *retryAttempts
	cfg.RetryDelay = *retryDelay
	cfg.RetryMaxDelay = *retryMaxDelay

	return nil
}
