package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Token is a signed access token.
type Token struct {
	AccessToken string    `json:"access_token"`
	Expiry      time.Time `json:"expiry"`
}

type claims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

// UserID verifies a signed access token and returns its subject.
func UserID(token, secret string) (uuid.UUID, error) {
	parsed, err := jwt.ParseWithClaims(token, &claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse token: %w", err)
	}

	tokenClaims, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}
	id, err := uuid.Parse(tokenClaims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse subject: %w", err)
	}
	return id, nil
}

// Sign creates an HMAC JWT for the user.
func Sign(userID uuid.UUID, email, secret string, ttl time.Duration) (Token, error) {
	expiry := time.Now().Add(ttl)
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
		Email: email,
	}).SignedString([]byte(secret))
	if err != nil {
		return Token{}, fmt.Errorf("sign token: %w", err)
	}
	return Token{AccessToken: signed, Expiry: expiry}, nil
}
