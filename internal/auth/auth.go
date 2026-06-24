package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   Role   `json:"role"`
	jwt.RegisteredClaims
}

type Authenticator struct {
	secret []byte
}

func New(secret string) (*Authenticator, error) {
	if len(secret) < 16 {
		return nil, fmt.Errorf("secret must be at least 16 characters")
	}
	return &Authenticator{secret: []byte(secret)}, nil
}

func (a *Authenticator) GenerateToken(userID string, role Role, ttl time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secret)
}

func (a *Authenticator) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

type contextKey string

const claimsKey contextKey = "auth_claims"

func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func ClaimsFromContext(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsKey).(*Claims)
	return claims
}

type Middleware struct {
	auth    *Authenticator
	allowed []Role
}

func NewMiddleware(auth *Authenticator, allowedRoles ...Role) *Middleware {
	return &Middleware{auth: auth, allowed: allowedRoles}
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if tokenStr == header {
			http.Error(w, "invalid authorization format", http.StatusUnauthorized)
			return
		}

		claims, err := m.auth.ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		if len(m.allowed) > 0 {
			allowed := false
			for _, role := range m.allowed {
				if claims.Role == role {
					allowed = true
					break
				}
			}
			if !allowed {
				http.Error(w, "insufficient permissions", http.StatusForbidden)
				return
			}
		}

		r = r.WithContext(ContextWithClaims(r.Context(), claims))
		next.ServeHTTP(w, r)
	})
}
