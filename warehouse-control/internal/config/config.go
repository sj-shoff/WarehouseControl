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
	Server struct {
		Addr            string        `env:"SERVER_PORT" validate:"required"`
		ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" validate:"required"`
		WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" validate:"required"`
		IdleTimeout     time.Duration `env:"SERVER_IDLE_TIMEOUT" validate:"required"`
		ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" validate:"required"`
	}
	JWT struct {
		Secret   string `env:"JWT_SECRET" validate:"required"`
		ExpHours int    `env:"JWT_EXP_HOURS" validate:"required"`
	}
	SSO struct {
		GRPCAddr      string        `env:"SSO_GRPC_ADDR" validate:"required"`
		ClientTimeout time.Duration `env:"SSO_CLIENT_TIMEOUT" validate:"required"`
	}
	Retries struct {
		Attempts int     `env:"RETRIES_ATTEMPTS" validate:"required"`
		DelayMs  int     `env:"RETRIES_DELAY_MS" validate:"required"`
		Backoff  float64 `env:"RETRIES_BACKOFF" validate:"required"`
	}
	RateLimit struct {
		Enabled  bool `env:"RATE_LIMIT_ENABLED" env-default:"true"`
		Rate     int  `env:"RATE_LIMIT_RATE" env-default:"5"`
		Capacity int  `env:"RATE_LIMIT_CAPACITY" env-default:"10"`
	}
}

func MustLoad() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read environment variables: %w", err)
	}
	if err := validator.New().Struct(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
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
