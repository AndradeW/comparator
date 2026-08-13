package config

import (
	"os"
	"strconv"
	"time"
)

const PORT = "8080"

const (
	defaultTimeout    = 10 * time.Second
	defaultMaxBodySize = int64(1 << 20) // 1 MiB
)

func GetPort() string {
	if port := os.Getenv("PORT"); port == "" {
		port = PORT
	}
	return ":" + PORT
}

// GetTimeout devuelve el timeout del cliente HTTP.
// Se configura con la env HTTP_TIMEOUT (formato de time.ParseDuration, ej: "10s").
func GetTimeout() time.Duration {
	envTimeout := os.Getenv("HTTP_TIMEOUT")
	if envTimeout == "" {
		return defaultTimeout
	}

	timeout, err := time.ParseDuration(envTimeout)
	if err != nil {
		return defaultTimeout
	}

	return timeout
}

// GetMaxBodySize devuelve el límite en bytes del request entrante.
// Se configura con la env MAX_BODY_SIZE (en bytes).
func GetMaxBodySize() int64 {
	envSize := os.Getenv("MAX_BODY_SIZE")
	if envSize == "" {
		return defaultMaxBodySize
	}

	size, err := strconv.ParseInt(envSize, 10, 64)
	if err != nil || size <= 0 {
		return defaultMaxBodySize
	}

	return size
}
