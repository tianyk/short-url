package config

import (
	"os"
	"testing"
)

func setEnv(t *testing.T, key, value string) {
	t.Helper()

	previous, existed := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, previous)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

func TestLoadUsesDefaultPort(t *testing.T) {
	setEnv(t, "APP_ENV", "test")
	setEnv(t, "APP_PORT", "")
	setEnv(t, "APP_PREFIX", "https://example.com/")
	setEnv(t, "APP_TOKEN", "")
	setEnv(t, "APP_STORE_PATH", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != defaultPort {
		t.Fatalf("expected default port %d, got %d", defaultPort, cfg.Port)
	}
	if cfg.Prefix != "https://example.com" {
		t.Fatalf("expected normalized prefix, got %q", cfg.Prefix)
	}
	if cfg.StorePath != defaultStorePath {
		t.Fatalf("expected default store path %q, got %q", defaultStorePath, cfg.StorePath)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	setEnv(t, "APP_ENV", "test")
	setEnv(t, "APP_PORT", "70000")
	setEnv(t, "APP_PREFIX", "not-a-url")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid configuration error")
	}
}

func TestLoadRequiresProductionToken(t *testing.T) {
	setEnv(t, "APP_ENV", "production")
	setEnv(t, "APP_PORT", "5000")
	setEnv(t, "APP_PREFIX", "https://example.com")
	setEnv(t, "APP_TOKEN", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected missing production token error")
	}
}
