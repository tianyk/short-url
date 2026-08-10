package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const defaultPort = 5000
const defaultStorePath = "short-url-store"

// Config contains the runtime configuration used by the application.
type Config struct {
	Environment string
	Port        int
	Prefix      string
	Token       string
	StorePath   string
}

// Load reads dotenv files and environment variables, then validates the result.
func Load() (Config, error) {
	environment := os.Getenv("APP_ENV")
	if environment == "" {
		environment = "development"
	}

	loadDotenv(environment)

	port := defaultPort
	if value := os.Getenv("APP_PORT"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf("parse APP_PORT: %w", err)
		}
		port = parsed
	}

	storePath := os.Getenv("APP_STORE_PATH")
	if storePath == "" {
		storePath = defaultStorePath
	}

	cfg := Config{
		Environment: environment,
		Port:        port,
		Prefix:      strings.TrimRight(os.Getenv("APP_PREFIX"), "/"),
		Token:       os.Getenv("APP_TOKEN"),
		StorePath:   storePath,
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func loadDotenv(environment string) {
	// Missing dotenv files are optional; process environment variables take precedence.
	_ = godotenv.Load(".env." + environment + ".local")
	if environment != "test" {
		_ = godotenv.Load(".env.local")
	}
	_ = godotenv.Load(".env." + environment)
	_ = godotenv.Load()
}

func (cfg Config) validate() error {
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("APP_PORT must be between 1 and 65535")
	}

	prefix, err := url.Parse(cfg.Prefix)
	if err != nil || prefix.Host == "" || (prefix.Scheme != "http" && prefix.Scheme != "https") {
		return fmt.Errorf("APP_PREFIX must be an absolute HTTP or HTTPS URL")
	}

	if cfg.Environment == "production" && cfg.Token == "" {
		return fmt.Errorf("APP_TOKEN is required in production")
	}
	if strings.TrimSpace(cfg.StorePath) == "" {
		return fmt.Errorf("APP_STORE_PATH must not be empty")
	}

	return nil
}
