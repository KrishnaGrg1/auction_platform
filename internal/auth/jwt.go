package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claim struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
type JWTManager struct {
	secret   string
	TokenTTL time.Duration
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{
		secret:   secret,
		TokenTTL: 7 * 24 * time.Hour,
	}
}

func (m *JWTManager) GenerateToken(UserId, Email string) (string, error) {
	claim := &Claim{
		UserID: UserId,
		Email:  Email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.TokenTTL)),
			Issuer:    "Auction-Platform",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(m.secret))
}

func (m *JWTManager) VerifyToken(tokenString string) (*Claim, error) {
	parsedToken, err := jwt.ParseWithClaims(
		tokenString,
		&Claim{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(m.secret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(*Claim)
	if !ok || !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
