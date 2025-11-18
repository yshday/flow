package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/yourusername/issue-tracker/internal/api"
	"github.com/yourusername/issue-tracker/pkg/cache"
	"github.com/yourusername/issue-tracker/pkg/database"
)

// @title Flow Issue Tracker API
// @version 1.0
// @description 프로젝트 기반 이슈 관리 시스템 API
// @description Jira/Linear와 유사한 칸반 보드 기능과 세밀한 권한 관리를 제공합니다.

// @contact.name API Support
// @contact.email support@issuetracker.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration from environment
	config := loadConfig()

	// Connect to database
	db, err := database.NewPostgresDB(database.Config{
		URL:            config.DatabaseURL,
		MaxConnections: config.DBMaxConnections,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Connect to Redis cache
	redisCache, err := cache.NewRedisCache(config.RedisAddr, config.RedisPassword, config.RedisDB)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisCache.Close()
	slog.Info("Connected to Redis cache", "addr", config.RedisAddr)

	// Create router
	router := api.NewRouter(api.Config{
		DB:                   db,
		Cache:                redisCache,
		JWTSecret:            config.JWTSecret,
		JWTRefreshSecret:     config.JWTRefreshSecret,
		JWTAccessTTL:         config.JWTAccessTTL,
		JWTRefreshTTL:        config.JWTRefreshTTL,
		StoragePath:          config.StoragePath,
		StorageMaxFileSize:   config.StorageMaxFileSize,
		RateLimitEnabled:     config.RateLimitEnabled,
		RateLimitPerMinute:   config.RateLimitPerMinute,
		RateLimitWindow:      config.RateLimitWindow,
		SMTPHost:             config.SMTPHost,
		SMTPPort:             config.SMTPPort,
		SMTPUsername:         config.SMTPUsername,
		SMTPPassword:         config.SMTPPassword,
		SMTPFrom:             config.SMTPFrom,
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Starting server", "port", config.Port, "env", config.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	slog.Info("Server exited")
}

// Config holds application configuration
type Config struct {
	Port               string
	Environment        string
	DatabaseURL        string
	DBMaxConnections   int
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	JWTSecret          string
	JWTRefreshSecret   string
	JWTAccessTTL       time.Duration
	JWTRefreshTTL      time.Duration
	RateLimitEnabled   bool
	RateLimitPerMinute int
	RateLimitWindow    time.Duration
	SMTPHost           string
	SMTPPort           string
	SMTPUsername       string
	SMTPPassword       string
	SMTPFrom           string
	StoragePath        string
	StorageMaxFileSize int64
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	return Config{
		Port:               getEnv("SERVER_PORT", getEnv("PORT", "8080")),
		Environment:        getEnv("ENV", "development"),
		DatabaseURL:        buildDatabaseURL(),
		DBMaxConnections:   25,
		RedisAddr:          buildRedisAddr(),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            parseInt(getEnv("REDIS_DB", "0"), 0),
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret-key-change-in-production"),
		JWTRefreshSecret:   getEnv("JWT_REFRESH_SECRET", "dev-refresh-secret-key-change-in-production"),
		JWTAccessTTL:       parseDuration(getEnv("JWT_ACCESS_TTL", "15m"), 15*time.Minute),
		JWTRefreshTTL:      parseDuration(getEnv("JWT_REFRESH_TTL", "168h"), 7*24*time.Hour), // 7 days
		RateLimitEnabled:   parseBool(getEnv("RATE_LIMIT_ENABLED", "true"), true),
		RateLimitPerMinute: parseInt(getEnv("RATE_LIMIT_REQUESTS_PER_MINUTE", "100"), 100),
		RateLimitWindow:    parseDuration(getEnv("RATE_LIMIT_WINDOW", "1m"), 1*time.Minute),
		SMTPHost:           getEnv("SMTP_HOST", ""),
		SMTPPort:           getEnv("SMTP_PORT", "587"),
		SMTPUsername:       getEnv("SMTP_USERNAME", ""),
		SMTPPassword:       getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:           getEnv("SMTP_FROM", "noreply@issuetracker.com"),
		StoragePath:        getEnv("STORAGE_PATH", "./uploads"),
		StorageMaxFileSize: parseInt64(getEnv("STORAGE_MAX_FILE_SIZE", "10485760"), 10*1024*1024), // 10MB
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parseDuration(value string, fallback time.Duration) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}

func parseInt64(value string, fallback int64) int64 {
	num, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return num
}

func parseInt(value string, fallback int) int {
	num, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return num
}

func parseBool(value string, fallback bool) bool {
	b, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return b
}

// buildDatabaseURL constructs PostgreSQL connection string from individual env vars
func buildDatabaseURL() string {
	// If DATABASE_URL is provided, use it directly
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	// Otherwise, build from individual components
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "issue_tracker")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname + "?sslmode=" + sslmode
}

// buildRedisAddr constructs Redis address from host and port env vars
func buildRedisAddr() string {
	// If REDIS_ADDR is provided, use it directly
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}

	// Otherwise, build from host and port
	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6379")
	return host + ":" + port
}
