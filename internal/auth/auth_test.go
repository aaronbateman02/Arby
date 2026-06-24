package auth_test

import (
	"testing"
	"time"

	"github.com/aaronbateman02/Arby/internal/auth"
)

func TestGenerateAndValidate(t *testing.T) {
	a, err := auth.New("test-secret-key-32-bytes-long!!")
	if err != nil {
		t.Fatalf("new auth failed: %v", err)
	}

	token, err := a.GenerateToken("user-1", auth.RoleOperator, 15*time.Minute)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	claims, err := a.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate token failed: %v", err)
	}

	if claims.UserID != "user-1" {
		t.Fatalf("expected user-1, got %s", claims.UserID)
	}
	if claims.Role != auth.RoleOperator {
		t.Fatalf("expected operator role, got %s", claims.Role)
	}
}

func TestInvalidToken(t *testing.T) {
	a, _ := auth.New("test-secret-key-32-bytes-long!!")
	_, err := a.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
