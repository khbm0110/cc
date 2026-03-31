package kms

import (
	"bytes"
	"context"
	"testing"
)

func TestMockKMS_EncryptDecrypt(t *testing.T) {
	km, err := NewMockKMSFromPassphrase("test-passphrase")
	if err != nil {
		t.Fatalf("failed to create KMS: %v", err)
	}

	plaintext := []byte("my-secret-api-key-12345")

	ciphertext, err := km.Encrypt(context.Background(), plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if bytes.Equal(plaintext, ciphertext) {
		t.Error("ciphertext should differ from plaintext")
	}

	decrypted, err := km.Decrypt(context.Background(), ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted text doesn't match: got %s, want %s", decrypted, plaintext)
	}
}

func TestMockKMS_DifferentCiphertexts(t *testing.T) {
	km, _ := NewMockKMSFromPassphrase("test-passphrase")
	plaintext := []byte("same-plaintext")

	ct1, _ := km.Encrypt(context.Background(), plaintext)
	ct2, _ := km.Encrypt(context.Background(), plaintext)

	if bytes.Equal(ct1, ct2) {
		t.Error("two encryptions of same plaintext should produce different ciphertexts (random nonce)")
	}
}

func TestMockKMS_InvalidKey(t *testing.T) {
	_, err := NewMockKMS([]byte("short"))
	if err == nil {
		t.Error("expected error for key shorter than 32 bytes")
	}
}

func TestMockKMS_DecryptWithWrongKey(t *testing.T) {
	km1, _ := NewMockKMSFromPassphrase("key-one")
	km2, _ := NewMockKMSFromPassphrase("key-two")

	ciphertext, _ := km1.Encrypt(context.Background(), []byte("secret"))

	_, err := km2.Decrypt(context.Background(), ciphertext)
	if err == nil {
		t.Error("expected decryption to fail with wrong key")
	}
}

func TestMockKMS_DecryptTooShort(t *testing.T) {
	km, _ := NewMockKMSFromPassphrase("test")

	_, err := km.Decrypt(context.Background(), []byte("short"))
	if err == nil {
		t.Error("expected error for ciphertext shorter than nonce")
	}
}

func TestMockKMS_EmptyPlaintext(t *testing.T) {
	km, _ := NewMockKMSFromPassphrase("test-passphrase")

	ciphertext, err := km.Encrypt(context.Background(), []byte{})
	if err != nil {
		t.Fatalf("encryption of empty plaintext failed: %v", err)
	}

	decrypted, err := km.Decrypt(context.Background(), ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("expected empty decrypted text, got %d bytes", len(decrypted))
	}
}
