package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	App  AppConfig
	HTTP HTTPConfig
	DB   DBConfig
	JWT  JWTConfig
}

type AppConfig struct {
	Env      string
	LogLevel string
}

type HTTPConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DBConfig struct {
	Type     string
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

type JWTConfig struct {
	Secret string
	Expire time.Duration
}

func Load() *Config {
	return &Config{
		App: AppConfig{
			Env:      env("APP_ENV", "development"),
			LogLevel: env("APP_LOG_LEVEL", "info"),
		},
		HTTP: HTTPConfig{
			Port:            envInt("APP_PORT", 8080),
			ReadTimeout:     envDuration("HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    envDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: envDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		DB: DBConfig{
			Type:     env("DB_TYPE", "sqlite3"),
			Host:     env("DB_HOST", "127.0.0.1"),
			Port:     envInt("DB_PORT", 3306),
			Name:     env("DB_NAME", "itsm_ops.db"),
			User:     env("DB_USER", "itsm"),
			Password: env("DB_PASSWORD", ""),
		},
		JWT: JWTConfig{
			Secret: env("JWT_SECRET", "please-change-me"),
			Expire: envDuration("JWT_EXPIRE", 24*time.Hour),
		},
	}
}

func env(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
