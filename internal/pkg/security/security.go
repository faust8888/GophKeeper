package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

type CryptoService struct {
	key []byte
}

const (
	HASH_BYTE_SIZE    = 32
	PBKDF2_ITERATIONS = 15000
)

func New(masterPassword string, salt string) *CryptoService {
	key := pbkdf2.Key([]byte(masterPassword), []byte(salt), PBKDF2_ITERATIONS, HASH_BYTE_SIZE, sha256.New)
	return &CryptoService{key: key}
}

func (c *CryptoService) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func (c *CryptoService) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(data) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil,
		data[:gcm.NonceSize()],
		data[gcm.NonceSize():],
		nil,
	)
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

func HashPassword(password string) (hash, salt string, err error) {
	if password == "" {
		return "", "", errors.New("password is empty")
	}

	salt = GenerateSalt()

	hashBytes := sha256.Sum256([]byte(password + salt))
	hash = base64.StdEncoding.EncodeToString(hashBytes[:])

	return hash, salt, nil
}

func VerifyPassword(password, hash, salt string) bool {
	hashBytes := sha256.Sum256([]byte(password + salt))
	computedHash := base64.StdEncoding.EncodeToString(hashBytes[:])
	return computedHash == hash
}

func GenerateSalt() string {
	saltBytes := make([]byte, 16)
	_, err := rand.Read(saltBytes)
	if err != nil {
		return uuid.New().String()
	}
	return base64.StdEncoding.EncodeToString(saltBytes)
}
