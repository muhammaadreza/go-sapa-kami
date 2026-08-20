package utils

import (
	"time"
	"errors"

	"go-sapa-kami/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

// JWT Claims Structs
// Refresh Claims
type RefreshClaims struct {
	UserID uint
	jwt.RegisteredClaims
}

// Access Claims
type AccessClaims struct {
	UserID   uint
	UserRole model.RoleEnum
	jwt.RegisteredClaims
}

// Generate Refresh Token
func GenerateRefreshToken(userID uint, secret []byte, expiry int, issuer string) (string, error) {
	register := jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry) * time.Hour)),
	}

	claims := RefreshClaims{
		UserID: userID,
		RegisteredClaims: register,
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil { return "", err }

	return token, nil
}

// Generate Access Token
func GenerateAccessToken(userID uint, userRole model.RoleEnum, secret []byte, expiry int, issuer string) (string, error) {
	register := jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry) * time.Minute)),
	}

	claims := AccessClaims{
		UserID:   userID,
		UserRole: userRole,
		RegisteredClaims: register,
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil { return "", err }

	return token, nil
}

// Validate Refresh Token
func ValidateRefreshToken(token string, secret []byte, issuer string) (*RefreshClaims, error) {
	tokenClaims := &RefreshClaims{}
	parsedToken, err := jwt.ParseWithClaims(
		token,
		tokenClaims,
		keyFunc(secret),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil { return nil, err }

	claims, ok := parsedToken.Claims.(*RefreshClaims)
	if ok && parsedToken.Valid { return claims, nil }

	return nil, errors.New("invalid refresh token")
}


func ValidateAccessToken(token string, secret []byte, issuer string) (*AccessClaims, error) {
	tokenClaims := &AccessClaims{}
	parsedToken, err := jwt.ParseWithClaims(
		token,
		tokenClaims,
		keyFunc(secret),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(issuer),
	)
	if err != nil { return nil, err }

	claims, ok := parsedToken.Claims.(*AccessClaims)
	if ok && parsedToken.Valid { return claims, nil }

	return nil, errors.New("invalid access token")
}

// KeyFunc for JWT Validation
func keyFunc(secret []byte) jwt.Keyfunc {
    return func(t *jwt.Token) (interface{}, error) { return secret, nil }
}
