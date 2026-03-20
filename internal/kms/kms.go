package kms

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

var (
	ErrDecryptionFailed = errors.New("decryption failed")
	ErrEncryptionFailed = errors.New("encryption failed")
)

// KeyManager defines the interface for encrypting/decrypting sensitive data.
type KeyManager interface {
	Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
}

// MockKMS implements KeyManager using AES-256-GCM with a static key.
// Suitable for local development and testing. Replace with real KMS
// (AWS KMS, GCP KMS, HashiCorp Vault) in production.
type MockKMS struct {
	key []byte // 32 bytes for AES-256
}

// NewMockKMS creates a new MockKMS with the given 32-byte key.
func NewMockKMS(key []byte) (*MockKMS, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes, got %d", len(key))
	}
	return &MockKMS{key: key}, nil
}

// NewMockKMSFromPassphrase creates a MockKMS by zero-padding a passphrase to 32 bytes.
// Only for development/testing.
func NewMockKMSFromPassphrase(passphrase string) (*MockKMS, error) {
	key := make([]byte, 32)
	copy(key, []byte(passphrase))
	return &MockKMS{key: key}, nil
}

func (m *MockKMS) Encrypt(_ context.Context, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(m.key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	return aesGCM.Seal(nonce, nonce, plaintext, nil), nil
}

func (m *MockKMS) Decrypt(_ context.Context, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(m.key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("%w: ciphertext too short", ErrDecryptionFailed)
	}

	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}
