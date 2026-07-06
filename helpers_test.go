package skywalker

import (
	"encoding/base64"
	"strings"
	"testing"
)

var encKey = []byte("abcdefghijklmnopqrstuvwxyz123456") // 32 bytes -> AES-256

func TestEncryption_EncryptDecryptRoundTrip(t *testing.T) {
	e := Encryption{Key: encKey}

	for _, plaintext := range []string{"", "a", "hello, world", strings.Repeat("x", 4096)} {
		encrypted, err := e.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("Encrypt(%q) returned error: %v", plaintext, err)
		}

		decrypted, err := e.Decrypt(encrypted)
		if err != nil {
			t.Fatalf("Decrypt of Encrypt(%q) returned error: %v", plaintext, err)
		}

		if decrypted != plaintext {
			t.Errorf("round trip mismatch: got %q, want %q", decrypted, plaintext)
		}
	}
}

func TestEncryption_EncryptIsNonDeterministic(t *testing.T) {
	e := Encryption{Key: encKey}

	first, err := e.Encrypt("same input")
	if err != nil {
		t.Fatal(err)
	}
	second, err := e.Encrypt("same input")
	if err != nil {
		t.Fatal(err)
	}

	if first == second {
		t.Error("two encryptions of the same plaintext produced identical output; nonce is not random")
	}
}

func TestEncryption_DecryptTamperedCiphertext(t *testing.T) {
	e := Encryption{Key: encKey}

	encrypted, err := e.Encrypt("some secret")
	if err != nil {
		t.Fatal(err)
	}

	raw, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatal(err)
	}

	// flip one bit in the last byte (part of the GCM tag)
	raw[len(raw)-1] ^= 0x01
	tampered := base64.URLEncoding.EncodeToString(raw)

	if _, err := e.Decrypt(tampered); err == nil {
		t.Error("Decrypt of tampered ciphertext returned nil error; want authentication failure")
	}
}

func TestEncryption_DecryptWrongKey(t *testing.T) {
	e := Encryption{Key: encKey}

	encrypted, err := e.Encrypt("some secret")
	if err != nil {
		t.Fatal(err)
	}

	other := Encryption{Key: []byte("654321zyxwvutsrqponmlkjihgfedcba")}
	if _, err := other.Decrypt(encrypted); err == nil {
		t.Error("Decrypt with wrong key returned nil error; want authentication failure")
	}
}

func TestEncryption_DecryptBadInput(t *testing.T) {
	e := Encryption{Key: encKey}

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"not base64", "!!!not-valid-base64!!!"},
		{"too short for nonce", base64.URLEncoding.EncodeToString([]byte("short"))},
		{"garbage of valid length", base64.URLEncoding.EncodeToString([]byte("this is long enough but garbage"))},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := e.Decrypt(tc.input); err == nil {
				t.Errorf("Decrypt(%q) returned nil error; want error", tc.input)
			}
		})
	}
}

func TestEncryption_InvalidKeyLength(t *testing.T) {
	e := Encryption{Key: []byte("too-short")}

	if _, err := e.Encrypt("anything"); err == nil {
		t.Error("Encrypt with invalid key length returned nil error; want error")
	}

	if _, err := e.Decrypt("anything"); err == nil {
		t.Error("Decrypt with invalid key length returned nil error; want error")
	}
}
