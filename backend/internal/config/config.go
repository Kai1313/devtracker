package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type AppConfig struct {
	Name            string
	Env             string
	Port            string
	BasePath        string
	BodyLimit       int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	URL             string
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	TimeZone        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	SlowThreshold   time.Duration
	RunMigrations   bool
}

type JWTConfig struct {
	Secret         string
	Issuer         string
	AccessTokenTTL time.Duration
}

func Load() Config {
	return Config{
		App: AppConfig{
			Name:            getEnv("APP_NAME", "Load Developer Sheets API"),
			Env:             getEnv("APP_ENV", "development"),
			Port:            getEnv("APP_PORT", "8080"),
			BasePath:        normalizeBasePath(getEnv("APP_BASE_PATH", "/api")),
			BodyLimit:       getEnvAsInt("APP_BODY_LIMIT", 4*1024*1024),
			ReadTimeout:     getEnvAsDuration("APP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvAsDuration("APP_WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout: getEnvAsDuration("APP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseConfig{
			URL:             os.Getenv("DATABASE_URL"),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "loaddev"),
			Password:        getEnv("DB_PASSWORD", "loaddevpass"),
			Name:            getEnv("DB_NAME", "loaddevdb"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			TimeZone:        getEnv("DB_TIMEZONE", "Asia/Jakarta"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", time.Hour),
			SlowThreshold:   getEnvAsDuration("DB_SLOW_THRESHOLD", 200*time.Millisecond),
			RunMigrations:   getEnvAsBool("DB_RUN_MIGRATIONS", true),
		},
		JWT: JWTConfig{
			Secret:         getEnv("JWT_SECRET", "devtracker-local-secret-change-me"),
			Issuer:         getEnv("JWT_ISSUER", "load-developer-sheets"),
			AccessTokenTTL: getEnvAsDuration("JWT_ACCESS_TOKEN_TTL", 24*time.Hour),
		},
	}
}

func (c DatabaseConfig) DSN() string {
	if c.URL != "" {
		return c.URL
	}

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		c.Host,
		c.User,
		c.Password,
		c.Name,
		c.Port,
		c.SSLMode,
		c.TimeZone,
	)
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvAsBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err == nil {
		return parsed
	}

	seconds, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return time.Duration(seconds) * time.Second
}

func normalizeBasePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		return ""
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return strings.TrimRight(path, "/")
}
