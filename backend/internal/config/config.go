package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv             string
	Port               string
	BaseURL            string
	SupabaseURL        string
	SupabaseKey        string
	RedisURL           string
	RedisMachineIDKey  string
	CacheTTL           time.Duration
	ShutdownTimeout    time.Duration
	ReadHeaderTimeout  time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	RateLimitWriteLimit int
	RateLimitWriteBurst int
	RateLimitReadLimit  int
	RateLimitReadBurst  int
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		Port:              getEnv("PORT", "8080"),
		BaseURL:           getEnv("BASE_URL", "http://localhost:8080"),
		SupabaseURL:       strings.TrimSpace(os.Getenv("SUPABASE_URL")),
		SupabaseKey:       strings.TrimSpace(os.Getenv("SUPABASE_KEY")),
		RedisURL:          strings.TrimSpace(os.Getenv("REDIS_URL")),
		RedisMachineIDKey: getEnv("REDIS_MACHINE_ID_KEY", "makeitshort:node:registration:counter"),
	}

	cacheTTLSeconds, err := getEnvInt("CACHE_TTL_SECONDS", 172800)
	if err != nil {
		return Config{}, err
	}
	cfg.CacheTTL = time.Duration(cacheTTLSeconds) * time.Second

	shutdownTimeoutSeconds, err := getEnvInt("SHUTDOWN_TIMEOUT_SECONDS", 10)
	if err != nil {
		return Config{}, err
	}
	cfg.ShutdownTimeout = time.Duration(shutdownTimeoutSeconds) * time.Second

	readHeaderTimeoutSeconds, err := getEnvInt("HTTP_READ_HEADER_TIMEOUT_SECONDS", 5)
	if err != nil {
		return Config{}, err
	}
	cfg.ReadHeaderTimeout = time.Duration(readHeaderTimeoutSeconds) * time.Second

	readTimeoutSeconds, err := getEnvInt("HTTP_READ_TIMEOUT_SECONDS", 10)
	if err != nil {
		return Config{}, err
	}
	cfg.ReadTimeout = time.Duration(readTimeoutSeconds) * time.Second

	writeTimeoutSeconds, err := getEnvInt("HTTP_WRITE_TIMEOUT_SECONDS", 10)
	if err != nil {
		return Config{}, err
	}
	cfg.WriteTimeout = time.Duration(writeTimeoutSeconds) * time.Second

	idleTimeoutSeconds, err := getEnvInt("HTTP_IDLE_TIMEOUT_SECONDS", 60)
	if err != nil {
		return Config{}, err
	}
	cfg.IdleTimeout = time.Duration(idleTimeoutSeconds) * time.Second

	cfg.RateLimitWriteLimit, err = getEnvInt("RATE_LIMIT_WRITE_LIMIT", 5)
	if err != nil {
		return Config{}, err
	}
	cfg.RateLimitWriteBurst, err = getEnvInt("RATE_LIMIT_WRITE_BURST", 5)
	if err != nil {
		return Config{}, err
	}
	cfg.RateLimitReadLimit, err = getEnvInt("RATE_LIMIT_READ_LIMIT", 100)
	if err != nil {
		return Config{}, err
	}
	cfg.RateLimitReadBurst, err = getEnvInt("RATE_LIMIT_READ_BURST", 200)
	if err != nil {
		return Config{}, err
	}

	if cfg.SupabaseURL == "" || cfg.SupabaseKey == "" {
		return Config{}, fmt.Errorf("SUPABASE_URL and SUPABASE_KEY are required")
	}

	if cfg.RedisURL == "" {
		return Config{}, fmt.Errorf("REDIS_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}

	if value <= 0 {
		return 0, fmt.Errorf("%s must be > 0", key)
	}

	return value, nil
}


