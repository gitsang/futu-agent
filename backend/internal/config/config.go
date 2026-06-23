package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerPort     int
	DatabaseURL    string
	FutuOpendHost  string
	FutuOpendPort  int
	LLMBaseURL     string
	LLMModel       string
	LLMAPIKey      string
	HTTPProxy      string
	TradingEnabled bool
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:     getEnvAsInt("SERVER_PORT", 8080),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/futu_agent?sslmode=disable"),
		FutuOpendHost:  getEnv("FUTU_OPEND_HOST", "localhost"),
		FutuOpendPort:  getEnvAsInt("FUTU_OPEND_PORT", 11111),
		LLMBaseURL:     getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
		LLMModel:       getEnv("LLM_MODEL", "gpt-4"),
		LLMAPIKey:      getEnv("LLM_API_KEY", ""),
		HTTPProxy:      getEnv("HTTP_PROXY", ""),
		TradingEnabled: getEnvAsBool("TRADING_ENABLED", false),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(strings.TrimSpace(value)); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
