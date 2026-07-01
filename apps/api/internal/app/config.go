package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime configuration loaded from environment variables.
// No secrets are committed; all values come from Render env / local .env.
type Config struct {
	Port            string
	Environment     string
	FrontendOrigins []string

	DatabaseURL    string
	DatabaseSkip   bool
	DBMaxOpenConns int
	DBMaxIdleConns int

	JWTSigningKey   string
	RefreshTokenKey string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	SupabaseURL            string
	SupabaseServiceRoleKey string
	SupabaseStorageBucket  string

	ResourceStorageType string
	ResourceLocalPath   string
	MaxUploadSize       int64

	RateLimitEnabled bool
	RateLimitRPS     float64
	RateLimitBurst   int
	RateLimitTTL     time.Duration
	RateLimitCleanup time.Duration

	SchedulerEnabled  bool
	SchedulerInterval time.Duration
}

// LoadConfig reads environment variables with safe defaults for local dev.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		FrontendOrigins: splitOrigins(getEnv("FRONTEND_ORIGINS", "http://localhost:5173")),

		DatabaseURL:    getEnv("DATABASE_URL", ""),
		DatabaseSkip:   getEnv("DB_SKIP", "false") == "true",
		DBMaxOpenConns: parseInt(getEnv("DB_MAX_OPEN_CONNS", "5")),
		DBMaxIdleConns: parseInt(getEnv("DB_MAX_IDLE_CONNS", "2")),

		JWTSigningKey:   getEnv("JWT_SIGNING_KEY", ""),
		RefreshTokenKey: getEnv("REFRESH_TOKEN_KEY", ""),
		AccessTokenTTL:  parseDuration(getEnv("ACCESS_TOKEN_TTL", "15m")),
		RefreshTokenTTL: parseDuration(getEnv("REFRESH_TOKEN_TTL", "7d")),

		SupabaseURL:            getEnv("SUPABASE_URL", ""),
		SupabaseServiceRoleKey: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		SupabaseStorageBucket:  getEnv("SUPABASE_STORAGE_BUCKET", "vts-edu-files"),

		ResourceStorageType: getEnv("RESOURCE_STORAGE_TYPE", "local"),
		ResourceLocalPath:   getEnv("RESOURCE_LOCAL_PATH", "/tmp/vts-edu-resources"),
		MaxUploadSize:       parseInt64(getEnv("MAX_UPLOAD_SIZE", "10485760")),

		RateLimitEnabled: getEnv("RATE_LIMIT_ENABLED", "false") == "true",
		RateLimitRPS:     parseFloat(getEnv("RATE_LIMIT_RPS", "10")),
		RateLimitBurst:   parseInt(getEnv("RATE_LIMIT_BURST", "20")),
		RateLimitTTL:     parseDuration(getEnv("RATE_LIMIT_TTL", "5m")),
		RateLimitCleanup: parseDuration(getEnv("RATE_LIMIT_CLEANUP", "1m")),

		SchedulerEnabled:  getEnv("SCHEDULER_ENABLED", "false") == "true",
		SchedulerInterval: time.Duration(parseInt(getEnv("SCHEDULER_INTERVAL_SECONDS", "60"))) * time.Second,
	}

	var missing []string
	if cfg.DatabaseURL == "" && !cfg.DatabaseSkip {
		missing = append(missing, "DATABASE_URL (or set DB_SKIP=true for local dev without DB)")
	}
	if cfg.JWTSigningKey == "" {
		missing = append(missing, "JWT_SIGNING_KEY")
	}
	if cfg.RefreshTokenKey == "" {
		missing = append(missing, "REFRESH_TOKEN_KEY")
	}
	if cfg.ResourceStorageType == "supabase" {
		// Fail fast: the service role key is server-side only and must
		// be present when RESOURCE_STORAGE_TYPE=supabase. The check
		// never logs the value.
		if cfg.SupabaseURL == "" {
			missing = append(missing, "SUPABASE_URL (required when RESOURCE_STORAGE_TYPE=supabase)")
		}
		if cfg.SupabaseServiceRoleKey == "" {
			missing = append(missing, "SUPABASE_SERVICE_ROLE_KEY (required when RESOURCE_STORAGE_TYPE=supabase; server-side only)")
		}
		if cfg.SupabaseStorageBucket == "" {
			missing = append(missing, "SUPABASE_STORAGE_BUCKET (required when RESOURCE_STORAGE_TYPE=supabase)")
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	if n <= 0 {
		return 5
	}
	return n
}

func parseInt64(s string) int64 {
	var n int64
	fmt.Sscanf(s, "%d", &n)
	if n <= 0 {
		return 10 * 1024 * 1024 // 10 MiB
	}
	return n
}

func parseFloat(s string) float64 {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil || n <= 0 {
		return 10.0
	}
	return n
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}

func splitOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
