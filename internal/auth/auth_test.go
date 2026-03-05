package auth_test

import (
	"testing"
	"time"

	"github.com/StephenCotterrell/chirpy/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		userID := uuid.New()
		tokenSecret := "test-secret"
		expiresIn := time.Minute

		before := time.Now().UTC()
		tokenString, err := auth.MakeJWT(userID, tokenSecret, expiresIn)
		after := time.Now().UTC()
		if err != nil {
			t.Fatalf("MakeJWT() failed: %v", err)
		}
		if tokenString == "" {
			t.Fatal("MakeJWT() returned empty token")
		}

		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(tokenSecret), nil
		})
		if err != nil {
			t.Fatalf("ParseWithClaims() failed: %v", err)
		}
		if !token.Valid {
			t.Fatal("parsed token is invalid")
		}

		if claims.Issuer != "chirpy-access" {
			t.Errorf("Issuer = %q, want %q", claims.Issuer, "chirpy-access")
		}
		if claims.Subject != userID.String() {
			t.Errorf("Subject = %q, want %q", claims.Subject, userID.String())
		}
		if claims.IssuedAt == nil {
			t.Fatal("IssuedAt is nil")
		}
		issuedAt := claims.IssuedAt.Time
		issuedLower := before.Add(-time.Second)
		issuedUpper := after.Add(time.Second)
		if issuedAt.Before(issuedLower) || issuedAt.After(issuedUpper) {
			t.Errorf("IssuedAt = %v, want between %v and %v", issuedAt, issuedLower, issuedUpper)
		}
		if claims.ExpiresAt == nil {
			t.Fatal("ExpiresAt is nil")
		}
		expiresAt := claims.ExpiresAt.Time
		minExp := before.Add(expiresIn).Add(-time.Second)
		maxExp := after.Add(expiresIn).Add(time.Second)
		if expiresAt.Before(minExp) || expiresAt.After(maxExp) {
			t.Errorf("ExpiresAt = %v, want between %v and %v", expiresAt, minExp, maxExp)
		}
	})
}

func TestValidateJWT(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		userID := uuid.New()
		tokenSecret := "test-secret"
		expiresIn := time.Minute

		tokenString, err := auth.MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("MakeJWT() failed: %v", err)
		}

		got, err := auth.ValidateJWT(tokenString, tokenSecret)
		if err != nil {
			t.Fatalf("ValidateJWT() failed: %v", err)
		}
		if got != userID {
			t.Errorf("ValidateJWT() = %v, want %v", got, userID)
		}
	})
}
