package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := Load()
	if cfg.App.Env != "development" {
		t.Errorf("App.Env = %s, want development", cfg.App.Env)
	}
	if cfg.HTTP.Port != 8080 {
		t.Errorf("HTTP.Port = %d, want 8080", cfg.HTTP.Port)
	}
	if cfg.DB.Type != "sqlite3" {
		t.Errorf("DB.Type = %s, want sqlite3", cfg.DB.Type)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Setenv("APP_PORT", "9090")
	defer os.Unsetenv("APP_ENV")
	defer os.Unsetenv("APP_PORT")

	cfg := Load()
	if cfg.App.Env != "production" {
		t.Errorf("App.Env = %s, want production", cfg.App.Env)
	}
	if cfg.HTTP.Port != 9090 {
		t.Errorf("HTTP.Port = %d, want 9090", cfg.HTTP.Port)
	}
}
