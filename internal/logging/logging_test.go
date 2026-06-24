package logging_test

import (
	"testing"

	"github.com/aaronbateman02/Arby/internal/logging"
)

func TestNew_ValidLevel(t *testing.T) {
	logger, err := logging.New("debug")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNew_InvalidLevel(t *testing.T) {
	_, err := logging.New("invalid")
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}
}
