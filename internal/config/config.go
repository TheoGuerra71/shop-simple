package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBSSLMode  string
	JWTSecret  string
	Port       string
	AppEnv     string
}

func Carregar() (*Config, error) {
	// Carrega .env se existir (não falha se não encontrar)
	godotenv.Load()

	cfg := &Config{
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "simple_shop"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", ""),
		Port:       getEnv("PORT", "7000"),
		AppEnv:     getEnv("APP_ENV", "development"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET é obrigatório. Configure no .env ou como variável de ambiente")
	}
	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD é obrigatório. Configure no .env ou como variável de ambiente")
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=%s",
		c.DBUser, c.DBPassword, c.DBName, c.DBHost, c.DBSSLMode)
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
