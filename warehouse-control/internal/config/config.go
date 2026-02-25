package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
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
}

func MustLoad() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read environment variables: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) DBDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DB.User, c.DB.Pass, c.DB.Host, c.DB.Port, c.DB.DBName)
}
