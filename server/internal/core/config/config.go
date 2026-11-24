package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	OAuth  OAuthConfig
	Logger LoggerConfig
}

type ServerConfig struct {
	Port            string
	ShutdownTimeout time.Duration
	AllowedOrigins  []string
}

type DBConfig struct {
	DSN             string
	OpTimeout       time.Duration
	ConnMaxLifetime time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
}

type OAuthConfig struct {
	FrontendURL string
	Google      GoogleOAuthConfig
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
}

type LoggerConfig struct {
	Level      string
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

func Load() (*Config, error) {
	dsn := strings.TrimSpace(os.Getenv("DB_DSN"))
	if dsn == "" {
		return nil, fmt.Errorf("DB_DSN environment variable is required")
	}

	frontendURL := strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_URL")), "/")

	return &Config{
		Server: ServerConfig{
			Port:            envOrDefault("SERVER_PORT", "5301"),
			ShutdownTimeout: 10 * time.Second,
			AllowedOrigins:  envSlice("ALLOWED_ORIGINS", frontendURL),
		},
		DB: DBConfig{
			DSN:             dsn,
			OpTimeout:       5 * time.Second,
			ConnMaxLifetime: 3 * time.Minute,
			MaxOpenConns:    10,
			MaxIdleConns:    10,
		},
		OAuth: OAuthConfig{
			FrontendURL: frontendURL,
			Google: GoogleOAuthConfig{
				ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
			},
		},
		Logger: LoggerConfig{
			Level:      envOrDefault("LOG_LEVEL", "info"),
			FilePath:   envOrDefault("LOG_FILE", "logs/server.log"),
			MaxSizeMB:  100,
			MaxBackups: 7,
			MaxAgeDays: 7,
			Compress:   true,
		},
	}, nil
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envSlice(key string, fallback ...string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return compactStrings(fallback)
	}
	return compactStrings(strings.Split(raw, ","))
}

func compactStrings(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if v := strings.TrimSpace(item); v != "" {
			result = append(result, v)
		}
	}
	return result
}
