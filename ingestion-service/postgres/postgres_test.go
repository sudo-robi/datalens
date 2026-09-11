package postgres

import (
	"testing"
)

func TestNewClientInvalidConnString(t *testing.T) {
	_, err := NewClient("postgres://invalid:invalid@localhost:99999/nonexistent?sslmode=disable")
	if err == nil {
		t.Error("expected error for invalid connection string, got nil")
	}
}

func TestNewClientEmptyConnString(t *testing.T) {
	_, err := NewClient("")
	if err == nil {
		t.Error("expected error for empty connection string, got nil")
	}
}

func TestNewClientUnreachableHost(t *testing.T) {
	_, err := NewClient("postgres://user:pass@localhost:1/nonexistent?sslmode=disable&connect_timeout=1")
	if err == nil {
		t.Error("expected error for unreachable host, got nil")
	}
}

func TestNewClientInvalidFormat(t *testing.T) {
	_, err := NewClient("not-a-valid-connection-string")
	if err == nil {
		t.Error("expected error for malformed connection string, got nil")
	}
}
