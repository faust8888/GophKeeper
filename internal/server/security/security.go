package security

import (
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

func HashPassword(password string) string {
	pass := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("error hashed password: %w", err)
	}
	return string(hashedPassword)
}

func CompareHash(hash string, password string) error {
	h := []byte(hash)
	return bcrypt.CompareHashAndPassword(h, []byte(password))
}

func Authenticate(token string, authKey string) (string, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		return []byte(authKey), nil
	})
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid token")
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "invalid token claims")
	}
	return sub, nil
}
