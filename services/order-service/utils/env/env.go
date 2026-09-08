package env

import (
	"os"
	"strconv"
	"time"
)

func GetString(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func GetInt(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return fallback
}

func GetBool(key string, fallback bool) bool {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return fallback
}

func GetDuration(key string, fallback time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if durationVal, err := time.ParseDuration(val); err == nil {
			return durationVal
		}
	}
	return fallback
}
