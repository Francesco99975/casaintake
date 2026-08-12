package helpers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Francesco99975/casaintake/cmd/boot"
	"golang.org/x/crypto/chacha20poly1305"
)

func ComputeFileChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])
	return checksum
}

func ParseBase64Key(keyStr string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func Encrypt(plaintext []byte, key []byte) (string, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(ciphertext string, key []byte) ([]byte, error) {
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}
	nonceSize := aead.NonceSize()
	nonce, ciphertext2 := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]
	return aead.Open(nil, nonce, ciphertext2, nil)
}

var timeSecret = []byte(boot.Environment.MetricSecret)

func GenerateStartToken() string {
	ts := time.Now().Unix()
	payload := strconv.FormatInt(ts, 10)

	mac := hmac.New(sha256.New, timeSecret)
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return payload + "." + sig
}

func ValidateStartToken(token string, minDuration time.Duration) error {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid token format")
	}

	payload, sig := parts[0], parts[1]

	// Verify signature
	mac := hmac.New(sha256.New, timeSecret)
	mac.Write([]byte(payload))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return fmt.Errorf("invalid signature")
	}

	ts, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp")
	}

	start := time.Unix(ts, 0)
	elapsed := time.Since(start)

	if elapsed < minDuration {
		return fmt.Errorf("submitted too fast")
	}

	// Optional: also reject if too old (e.g. > 30 minutes)
	if elapsed > 30*time.Minute {
		return fmt.Errorf("form expired")
	}

	return nil
}
