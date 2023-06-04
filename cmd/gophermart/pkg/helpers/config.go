package helpers

import "os"

func GetStringEnv(key string, fallback *string) *string {
	if value := os.Getenv(key); len(value) > 0 {
		return &value
	}

	return fallback
}
