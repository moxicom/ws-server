package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerAddr string
	RedisAddr  string
	RedisPass  string
	RedisDB    int
}

func NewConfig() Config {
	return Config{
		ServerAddr: getEnv("SERVER_ADDR", ":8080"),
		RedisAddr:  getEnv("REDIS_ADDR", ":6379"),
		RedisPass:  getEnv("REDIS_PASS", ""),
		RedisDB:    getEnvAsInt("REDIS_DB", 0),
	}
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

func getEnv(key string, defaultVal string) string {
	if value := os.Getenv(key); len(value) > 0 {
		return value
	}

	return defaultVal
}
