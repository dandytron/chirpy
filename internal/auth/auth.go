package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenType string

const (
	// TokenTypeAccess
	TokenTypeAccess TokenType = "chirpy-access"
)

// ErrNoAuthHeaderIncluded -
var ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")

// HashPassword
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	stringifiedHashedPassword := string(hashedPassword)
	return stringifiedHashedPassword, nil
}

// CheckPasswordHash
func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// Make a JSON Web Token
func MakeJWT(
	userID uuid.UUID,
	tokenSecret string,
	expiresIn time.Duration,
) (string, error) {
	signingKey := []byte(tokenSecret)
	newJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})

	return newJWT.SignedString(signingKey)
}

// Validate an incoming JSON Web Token
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return uuid.Nil, err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	id, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return id, nil
}

// GetBearerToken
func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}

	// Check if it starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("authorization header format must be 'Bearer {token}'")
	}

	// Extract the token
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return strings.TrimSpace(token), nil

}

// This generates a random 256-bit (32-byte) hex-encoded string.
func MakeRefreshToken() (string, error) {
	randomData := make([]byte, 32)
	// Read fills randomBytes with cryptographically secure random bytes. It never returns an error, and always fills b entirely.
	// No need for error-handling.
	_, err := rand.Read(randomData)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(randomData), nil
}

// This extracts the API key from the Authorization header, which is
// expected to be in this format: Authorization: ApiKey THE_KEY_HERE

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}

	// Check if it starts with "ApiKey "
	if !strings.HasPrefix(authHeader, "ApiKey ") {
		return "", errors.New("authorization header format must be 'ApiKey {KEY}'")
	}

	key := strings.TrimPrefix(authHeader, "ApiKey ")
	return strings.TrimSpace(key), nil
}
