package config

import (
	"os"
	"testing"
	"time"
)

func TestGetTimeout_default(t *testing.T) {
	// Asegurar que la env no está seteada
	old, ok := os.LookupEnv("HTTP_TIMEOUT")
	os.Unsetenv("HTTP_TIMEOUT")
	defer func() {
		if ok {
			os.Setenv("HTTP_TIMEOUT", old)
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
	os.Unsetenv("MAX_BODY_SIZE")
	defer func() {
		if ok {
			os.Setenv("MAX_BODY_SIZE", old)
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
