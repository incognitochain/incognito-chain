package utils

import "os"

// GetEnv to get environment variable by key
func GetEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}
