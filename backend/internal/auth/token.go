package auth

import (
	"errors"
	"time"

	"devtracker/backend/internal/config"
	apperrors "devtracker/backend/pkg/errors"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenInput struct {
	UserID string
	Email  string
	Name   string
	Role   string
}

type TokenManager struct {
	secret string
	issuer string
	ttl    time.Duration
}

func NewTokenManager(cfg config.JWTConfig) *TokenManager {
	return &TokenManager{
		secret: cfg.Secret,
		issuer: cfg.Issuer,
		ttl:    cfg.AccessTokenTTL,
	}
}

func (m *TokenManager) Generate(input TokenInput) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.ttl)

	claims := Claims{
		UserID: input.UserID,
		Email:  input.Email,
		Name:   input.Name,
		Role:   input.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   input.UserID,
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func (m *TokenManager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	options := []jwt.ParserOption{
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	}

	if m.issuer != "" {
		options = append(options, jwt.WithIssuer(m.issuer))
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(m.secret), nil
	}, options...)
	if err != nil {
		return nil, apperrors.Unauthorized("invalid or expired token")
	}

	if !token.Valid {
		return nil, apperrors.Unauthorized("invalid or expired token")
	}

	if claims.UserID == "" || claims.Subject == "" {
		return nil, apperrors.Unauthorized("invalid token claims")
	}

	if claims.UserID != claims.Subject {
		return nil, apperrors.Unauthorized("invalid token subject")
	}

	return claims, nil
}

func TokenFromError(err error) *apperrors.AppError {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return apperrors.Unauthorized("invalid or expired token")
}
