package rpccaller

import (
	"os"
)

// GetENV to get environment variable by key
func GetENV(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}