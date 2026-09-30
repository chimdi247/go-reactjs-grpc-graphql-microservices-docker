package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// The JWT is minted and verified here at the API gateway — account
// service only ever verifies a raw email+password (see Authenticate RPC)
// and knows nothing about tokens, which keeps auth-token concerns at the
// edge, where the rest of the GraphQL API already lives.
var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		// Fine for local/dev docker-compose; set a real secret in prod.
		s = "insecure-dev-secret-change-me"
	}
	return s
}

type contextKey string

const accountIDContextKey contextKey = "accountID"

type claims struct {
	AccountID string `json:"accountId"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

func signToken(accountID, email string) (string, error) {
	c := claims{
		AccountID: accountID,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(jwtSecret)
}

func parseToken(tokenString string) (*claims, error) {
	c := &claims{}
	token, err := jwt.ParseWithClaims(tokenString, c, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return c, nil
}

// authMiddleware extracts an optional "Authorization: Bearer <token>"
// header and, if it's a valid token, stores the account ID in the
// request context. Requests with no token (or an invalid one) still
// proceed — this is what lets public queries (browsing the catalog)
// work without logging in, while resolvers that need auth (me,
// createOrder) check ctx themselves and reject if it's missing.
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			tokenString := strings.TrimPrefix(header, "Bearer ")
			if c, err := parseToken(tokenString); err == nil {
				ctx := context.WithValue(r.Context(), accountIDContextKey, c.AccountID)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func accountIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(accountIDContextKey).(string)
	return id, ok && id != ""
}
