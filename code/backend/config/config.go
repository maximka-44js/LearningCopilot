package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

type (
	// Config -.
	Config struct {
		App     app
		HTTP    http
		Log     log
		PG      pg
		JWT     jwt
		CORS    cors
		S3      s3
		Metrics metrics
		Swagger swagger
		River   river
		APIKeys apiKeys
	}

	// WorkerConfig -.
	WorkerConfig struct {
		LogLevel        string `env:"LOG_LEVEL" envDefault:"info"`
		PGURL           string `env:"PG_URL,required"`
		PGPoolMax       int    `env:"PG_POOL_MAX" envDefault:"5"`
		RiverQueue      string `env:"RIVER_QUEUE" envDefault:"default"`
		RiverMaxWorkers int    `env:"RIVER_MAX_WORKERS" envDefault:"10"`
		S3              s3
	}

	// App -.
	app struct {
		Name    string `env:"APP_NAME,required"`
		Version string `env:"APP_VERSION,required"`
	}

	// HTTP -.
	http struct {
		Port            string        `env:"HTTP_PORT,required"`
		UsePreforkMode  bool          `env:"HTTP_USE_PREFORK_MODE" envDefault:"false"`
		ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"120s"`
		WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"120s"`
		ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"3s"`
		BodyLimitBytes  int           `env:"HTTP_BODY_LIMIT_BYTES" envDefault:"268435456"`
	}

	// Log -.
	log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	// PG -.
	pg struct {
		PoolMax int    `env:"PG_POOL_MAX,required"`
		URL     string `env:"PG_URL,required"`
	}

	// JWT -.
	jwt struct {
		Issuer          string `env:"JWT_ISSUER,required"`
		DiscoveryIssuer string `env:"JWT_DISCOVERY_ISSUER,required"`
		ClientID        string `env:"JWT_CLIENT_ID,required"`
	}

	// CORS -.
	cors struct {
		AllowOrigins     string `env:"CORS_ALLOW_ORIGINS" envDefault:"http://localhost:8080,http://127.0.0.1:8080,http://127.0.0.1:8000,http://localhost:3000,http://127.0.0.1:3000"`
		AllowHeaders     string `env:"CORS_ALLOW_HEADERS" envDefault:"*"`
		AllowMethods     string `env:"CORS_ALLOW_METHODS" envDefault:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
		AllowCredentials bool   `env:"CORS_ALLOW_CREDENTIALS" envDefault:"true"`
	}

	// S3 -.
	s3 struct {
		EndpointHost  string        `env:"S3_ENDPOINT_URL"`
		Port          int           `env:"S3_PORT" envDefault:"443"`
		Region        string        `env:"S3_REGION_NAME,required"`
		AccessKey     string        `env:"S3_ACCESS_KEY,required"`
		SecretKey     string        `env:"S3_SECRET_KEY,required"`
		Bucket        string        `env:"S3_BUCKET_NAME,required"`
		UsePathStyle  bool          `env:"S3_USE_PATH_STYLE" envDefault:"true"`
		PresignExpire time.Duration `env:"S3_PRESIGN_EXPIRE" envDefault:"15m"`
	}

	// Metrics -.
	metrics struct {
		Enabled bool `env:"METRICS_ENABLED" envDefault:"true"`
	}

	// Swagger -.
	swagger struct {
		Enabled bool `env:"SWAGGER_ENABLED" envDefault:"false"`
	}

	// River -.
	river struct {
		Queue      string `env:"RIVER_QUEUE" envDefault:"default"`
		MaxWorkers int    `env:"RIVER_MAX_WORKERS" envDefault:"10"`
	}

	// API keys -.
	apiKeys struct {
		Enabled           bool          `env:"API_KEYS_ENABLED" envDefault:"false"`
		HashPepper        string        `env:"API_KEYS_HASH_PEPPER"`
		MaxPerUser        int           `env:"API_KEYS_MAX_PER_USER" envDefault:"20"`
		RotateGraceWindow time.Duration `env:"API_KEYS_ROTATE_GRACE_WINDOW" envDefault:"24h"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}

// NewWorkerConfig returns worker config.
func NewWorkerConfig() (*WorkerConfig, error) {
	cfg := &WorkerConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("worker config error: %w", err)
	}

	return cfg, nil
}

func (s s3) Endpoint() string {
	if s.EndpointHost == "" {
		return ""
	}
	if strings.HasPrefix(s.EndpointHost, "http://") || strings.HasPrefix(s.EndpointHost, "https://") {
		return s.EndpointHost
	}

	scheme := "https"
	if s.Port == 80 {
		scheme = "http"
	}

	if s.Port > 0 {
		return fmt.Sprintf("%s://%s:%d", scheme, s.EndpointHost, s.Port)
	}

	return fmt.Sprintf("%s://%s", scheme, s.EndpointHost)
}
