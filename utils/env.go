package utils

import (
	"os"
)

func GetEnv(key, value string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return value
}
