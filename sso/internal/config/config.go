package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/wb-go/wbf/retry"
)

type Config struct {
	DB struct {
		Host            string        `env:"POSTGRES_HOST" validate:"required"`
		Port            int           `env:"POSTGRES_PORT" validate:"required"`
		User            string        `env:"POSTGRES_USER" validate:"required"`
		Pass            string        `env:"POSTGRES_PASSWORD" validate:"required"`
		DBName          string        `env:"POSTGRES_DB" validate:"required"`
		MaxOpenConns    int           `env:"POSTGRES_MAX_OPEN_CONNS"`
		MaxIdleConns    int           `env:"POSTGRES_MAX_IDLE_CONNS"`
		ConnMaxLifetime time.Duration `env:"POSTGRES_CONN_MAX_LIFETIME"`
	}
	GRPC struct {
		Port    string        `env:"GRPC_PORT" validate:"required"`
		Timeout time.Duration `env:"GRPC_TIMEOUT" env-default:"10s"`
	}
	JWT struct {
		Secret          string        `env:"JWT_SECRET" validate:"required"`
		AccessTokenTTL  time.Duration `env:"JWT_TOKEN_TTL" env-default:"24h"`
		RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env-default:"720h"`
	}
	Retries struct {
		Attempts int     `env:"RETRIES_ATTEMPTS" validate:"required"`
		DelayMs  int     `env:"RETRIES_DELAY_MS" validate:"required"`
		Backoff  float64 `env:"RETRIES_BACKOFF" validate:"required"`
	}
}

func MustLoad() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read env: %w", err)
	}
	if err := validator.New().Struct(&cfg); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	return &cfg, nil
}

func (c *Config) DBDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DB.User, c.DB.Pass, c.DB.Host, c.DB.Port, c.DB.DBName)
}

func (c *Config) DefaultRetryStrategy() retry.Strategy {
	return retry.Strategy{
		Attempts: c.Retries.Attempts,
		Delay:    time.Duration(c.Retries.DelayMs) * time.Millisecond,
		Backoff:  c.Retries.Backoff,
	}
}
