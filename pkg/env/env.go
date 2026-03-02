package env

import (
	"os"
	"strconv"
	"time"
)

func GetInt(name string, defaultValue int) int {
	raw, found := os.LookupEnv(name)
	if !found || raw == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}

	return value
}

func GetInt64(name string, defaultValue int64) int64 {
	raw, found := os.LookupEnv(name)
	if !found || raw == "" {
		return defaultValue
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return defaultValue
	}

	return value
}

// 1000000000 = 1s
// 1000000 = 1ms
// Default unit is millisecond.
func GetDuration(name string, defaultValue time.Duration) time.Duration {
	raw, found := os.LookupEnv(name)
	if !found || raw == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}

	return time.Duration(value) * time.Millisecond
}

func GetString(name string, defaultValue string) string {
	value, found := os.LookupEnv(name)
	if !found || value == "" {
		return defaultValue
	}
	return value
}

func GetBool(name string, defaultValue bool) bool {
	raw, found := os.LookupEnv(name)
	if !found || raw == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return defaultValue
	}

	return value
}
