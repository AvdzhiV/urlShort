package configs

import (
	"testing"
)

func TestSet_ValidAddress(t *testing.T) {
	cfg := &Config{}
	err := cfg.Set("localhost:9090")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Expected Host to be 'localhost', got '%s'", cfg.Host)
	}

	if cfg.Port != 9090 {
		t.Errorf("Expected Port to be 9090, got '%d'", cfg.Port)
	}
}
