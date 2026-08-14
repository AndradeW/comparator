package config

import (
	"os"
	"testing"
	"time"
)

func TestGetTimeout_default(t *testing.T) {
	// Asegurar que la env no está seteada
	old, ok := os.LookupEnv("HTTP_TIMEOUT")
	_ = os.Unsetenv("HTTP_TIMEOUT")
	defer func() {
		if ok {
			_ = os.Setenv("HTTP_TIMEOUT", old)
		}
	}()

	if got := GetTimeout(); got != defaultTimeout {
		t.Fatalf("GetTimeout() = %v, se esperaba %v", got, defaultTimeout)
	}
}

func TestGetTimeout_envValida(t *testing.T) {
	t.Setenv("HTTP_TIMEOUT", "5s")

	if got := GetTimeout(); got != 5*time.Second {
		t.Fatalf("GetTimeout() = %v, se esperaba 5s", got)
	}
}

func TestGetTimeout_envInvalida(t *testing.T) {
	t.Setenv("HTTP_TIMEOUT", "no-es-una-duracion")

	if got := GetTimeout(); got != defaultTimeout {
		t.Fatalf("GetTimeout() = %v, se esperaba default %v", got, defaultTimeout)
	}
}

func TestGetMaxBodySize_default(t *testing.T) {
	old, ok := os.LookupEnv("MAX_BODY_SIZE")
	_ = os.Unsetenv("MAX_BODY_SIZE")
	defer func() {
		if ok {
			_ = os.Setenv("MAX_BODY_SIZE", old)
		}
	}()

	if got := GetMaxBodySize(); got != defaultMaxBodySize {
		t.Fatalf("GetMaxBodySize() = %d, se esperaba %d", got, defaultMaxBodySize)
	}
}

func TestGetMaxBodySize_envValida(t *testing.T) {
	t.Setenv("MAX_BODY_SIZE", "2048")

	if got := GetMaxBodySize(); got != 2048 {
		t.Fatalf("GetMaxBodySize() = %d, se esperaba 2048", got)
	}
}

func TestGetMaxBodySize_envInvalida(t *testing.T) {
	t.Setenv("MAX_BODY_SIZE", "no-es-un-numero")

	if got := GetMaxBodySize(); got != defaultMaxBodySize {
		t.Fatalf("GetMaxBodySize() = %d, se esperaba default %d", got, defaultMaxBodySize)
	}
}

func TestGetMaxBodySize_envCero(t *testing.T) {
	t.Setenv("MAX_BODY_SIZE", "0")

	if got := GetMaxBodySize(); got != defaultMaxBodySize {
		t.Fatalf("GetMaxBodySize() = %d, se esperaba default %d", got, defaultMaxBodySize)
	}
}

func TestGetPort_default(t *testing.T) {
	old, ok := os.LookupEnv("PORT")
	_ = os.Unsetenv("PORT")
	defer func() {
		if ok {
			_ = os.Setenv("PORT", old)
		}
	}()

	if got := GetPort(); got != ":"+PORT {
		t.Fatalf("GetPort() = %q, se esperaba %q", got, ":"+PORT)
	}
}

func TestGetPort_env(t *testing.T) {
	t.Setenv("PORT", "9090")

	if got := GetPort(); got != ":9090" {
		t.Fatalf("GetPort() = %q, se esperaba :9090", got)
	}
}

func TestGetMaxResponseSize_default(t *testing.T) {
	old, ok := os.LookupEnv("MAX_RESPONSE_SIZE")
	_ = os.Unsetenv("MAX_RESPONSE_SIZE")
	defer func() {
		if ok {
			_ = os.Setenv("MAX_RESPONSE_SIZE", old)
		}
	}()

	if got := GetMaxResponseSize(); got != defaultMaxResponseSize {
		t.Fatalf("GetMaxResponseSize() = %d, se esperaba %d", got, defaultMaxResponseSize)
	}
}

func TestGetMaxResponseSize_envValida(t *testing.T) {
	t.Setenv("MAX_RESPONSE_SIZE", "4096")

	if got := GetMaxResponseSize(); got != 4096 {
		t.Fatalf("GetMaxResponseSize() = %d, se esperaba 4096", got)
	}
}

func TestGetMaxResponseSize_envInvalida(t *testing.T) {
	t.Setenv("MAX_RESPONSE_SIZE", "no-es-un-numero")

	if got := GetMaxResponseSize(); got != defaultMaxResponseSize {
		t.Fatalf("GetMaxResponseSize() = %d, se esperaba default %d", got, defaultMaxResponseSize)
	}
}

func TestGetCORSAllowedOrigin_sinConfigurar(t *testing.T) {
	old, ok := os.LookupEnv("CORS_ALLOWED_ORIGIN")
	_ = os.Unsetenv("CORS_ALLOWED_ORIGIN")
	defer func() {
		if ok {
			_ = os.Setenv("CORS_ALLOWED_ORIGIN", old)
		}
	}()

	if got := GetCORSAllowedOrigin(); got != "" {
		t.Fatalf("GetCORSAllowedOrigin() = %q, se esperaba \"\" (CORS deshabilitado)", got)
	}
}

func TestGetCORSAllowedOrigin_env(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGIN", "https://example.com")

	if got := GetCORSAllowedOrigin(); got != "https://example.com" {
		t.Fatalf("GetCORSAllowedOrigin() = %q, se esperaba https://example.com", got)
	}
}
