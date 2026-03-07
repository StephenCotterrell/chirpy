package auth_test

import (
	"encoding/hex"
	"net/http"
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

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantErr   bool
	}{
		{
			name: "Valid Bearer Token",
			headers: http.Header{
				"Authorization": []string{"Bearer valid_token"},
			},
			wantToken: "valid_token",
			wantErr:   false,
		},
		{
			name:      "Missing Authorization Header",
			headers:   http.Header{},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "Malformed Authorization Header",
			headers: http.Header{
				"Authorization": []string{"InvalidBearer token"},
			},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := auth.GetBearerToken(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() err = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotToken != tt.wantToken {
				t.Errorf("GetBearerToken() gotToken = %v, want %v", gotToken, tt.wantToken)
			}
		})
	}
}

func TestMakeRefreshToken(t *testing.T) {
	tests := []struct {
		name    string
		count   int
		wantLen int
		unique  bool
	}{
		{
			name:    "single_token_has_expected_shape",
			count:   1,
			wantLen: 64,
		},
		{
			name:   "multiple_tokens_are_unique",
			count:  3,
			unique: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.count == 0 {
				tt.count = 1
			}

			tokens := make([]string, 0, tt.count)
			for i := 0; i < tt.count; i++ {
				tokens = append(tokens, auth.MakeRefreshToken())
			}

			first := tokens[0]
			if first == "" {
				t.Fatal("token is empty")
			}
			if tt.wantLen > 0 && len(first) != tt.wantLen {
				t.Fatalf("token length = %d, want %d", len(first), tt.wantLen)
			}
			if _, err := hex.DecodeString(first); err != nil {
				t.Fatalf("token is not valid hex: %v", err)
			}
			if tt.unique {
				seen := make(map[string]struct{}, len(tokens))
				for _, tok := range tokens {
					seen[tok] = struct{}{}
				}
				if len(seen) != len(tokens) {
					t.Fatal("tokens should be unique across calls")
				}
			}
		})
	}
}
