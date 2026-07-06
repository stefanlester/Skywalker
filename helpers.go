package skywalker

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
)

const (
	randomString = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// RandomString generates a random string of length n from values in the
// const randomString, drawing each character uniformly via crypto/rand.
func (c *Skywalker) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomString)

	max := big.NewInt(int64(len(r)))
	for i := range s {
		x, _ := rand.Int(rand.Reader, max)
		s[i] = r[x.Int64()]
	}

	return string(s)
}

// CreateDirIf NotExsit creates a new directory if it doesn't exist.
func (c *Skywalker) CreateDirIfNotExist(path string) error {
	const mode = 0755

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, mode)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateFileIfNotExist creates a new file at path if it does not exist.
func (c *Skywalker) CreateFileIfNotExist(path string) error {
	var _, err = os.Stat(path)

	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return err
		}

		defer func(file *os.File) {
			_ = file.Close()
		}(file)
	}

	return nil
}

// Encryption encrypts and decrypts strings using AES-GCM with the supplied
// Key (16, 24, or 32 bytes for AES-128/192/256).
type Encryption struct {
	Key []byte
}

// Encrypt encrypts text with AES-GCM using a random nonce. The result is
// base64.URLEncoding(nonce || ciphertext), so it is safe to use in URLs and
// cookies. GCM is authenticated, so Decrypt detects any tampering.
func (e *Encryption) Encrypt(text string) (string, error) {
	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt reverses Encrypt. It returns an error if the input is not valid
// base64, is too short to contain a nonce, was encrypted with a different
// key, or has been tampered with (GCM authentication failure).
func (e *Encryption) Decrypt(cryptoText string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", fmt.Errorf("decrypt: invalid base64 input: %w", err)
	}

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", errors.New("decrypt: ciphertext shorter than nonce")
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: authentication failed: %w", err)
	}

	return string(plaintext), nil
}
