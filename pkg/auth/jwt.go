package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = func() []byte {
	password := os.Getenv("TODO_PASSWORD")
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

func GenerateToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(8 * time.Hour).Unix(),
		"pwd": hex.EncodeToString(secretKey()),
	})

	return token.SignedString(secretKey())
}

func ValidateToken(tokenString string) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey(), nil
	})

	if err != nil || !token.Valid {
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	storedHash, ok := claims["pwd"].(string)
	if !ok {
		return false
	}

	return storedHash == hex.EncodeToString(secretKey())
}
