package config

import (
	"os"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	DSN string
}

type HTTPConfig struct {
	Addr string
}

type Config struct {
	DB   DBConfig
	HTTP HTTPConfig
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		DB: DBConfig{
			DSN: getenv("DB_DSN",
				"postgres://pr_service:pr_service@localhost:5432/pr_service?sslmode=disable"),
		},
		HTTP: HTTPConfig{
			Addr: getenv("HTTP_ADDR", ":8080"),
		},
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
