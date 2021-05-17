package utils

import (
	"os"
)

// GetEnv to get environment variable by key
func GetEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		//log.Println("[prefix] fallback:", fallback)
		return fallback
	}
	//log.Println("[prefix] value:", value)
	return value
}
