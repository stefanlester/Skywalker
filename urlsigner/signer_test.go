package urlsigner

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"strings"
	"testing"
	"time"
)

var signer = Signer{Secret: []byte("abc123abc123abc123abc123abc12345")}

func TestSigner_RoundTripNoQuery(t *testing.T) {
	token := signer.GenerateTokenFromString("https://example.com/reset/abc")

	if !strings.Contains(token, "?hash=") {
		t.Errorf("expected ?hash= separator in token %s", token)
	}

	if !signer.VerifyToken(token) {
		t.Error("valid token did not verify")
	}
}

func TestSigner_RoundTripExistingQuery(t *testing.T) {
	token := signer.GenerateTokenFromString("https://example.com/reset?email=me@here.com")

	if !strings.Contains(token, "&hash=") {
		t.Errorf("expected &hash= separator in token %s", token)
	}

	if !signer.VerifyToken(token) {
		t.Error("valid token did not verify")
	}
}

func TestSigner_TamperedURLFails(t *testing.T) {
	token := signer.GenerateTokenFromString("https://example.com/reset?email=me@here.com")

	tampered := strings.Replace(token, "me@here.com", "you@there.com", 1)
	if signer.VerifyToken(tampered) {
		t.Error("tampered url verified but should not have")
	}
}

func TestSigner_TamperedHashFails(t *testing.T) {
	token := signer.GenerateTokenFromString("https://example.com/reset/abc")

	// Decode the hash payload, flip one bit of the MAC, and re-encode, so the
	// tampering survives base64's ignored trailing padding bits.
	idx := strings.LastIndex(token, hashParam) + len(hashParam)
	payload, err := base64.RawURLEncoding.DecodeString(token[idx:])
	if err != nil {
		t.Fatal(err)
	}
	payload[len(payload)-1] ^= 0x01
	tampered := token[:idx] + base64.RawURLEncoding.EncodeToString(payload)

	if signer.VerifyToken(tampered) {
		t.Error("tampered hash verified but should not have")
	}
}

func TestSigner_TruncatedHashFails(t *testing.T) {
	token := signer.GenerateTokenFromString("https://example.com/reset/abc")

	if signer.VerifyToken(token[:len(token)-4]) {
		t.Error("truncated hash verified but should not have")
	}
}

func TestSigner_WrongSecretFails(t *testing.T) {
	token := signer.GenerateTokenFromString("https://example.com/reset/abc")

	other := Signer{Secret: []byte("some-other-secret-entirely-here!")}
	if other.VerifyToken(token) {
		t.Error("token verified with the wrong secret")
	}
}

func TestSigner_MalformedTokens(t *testing.T) {
	malformed := []string{
		"",
		"https://example.com/no-hash-at-all",
		"https://example.com/reset?hash=",
		"https://example.com/reset?hash=!!!not-base64!!!",
		"https://example.com/reset?hash=" + base64.RawURLEncoding.EncodeToString([]byte("too short")),
	}

	for _, token := range malformed {
		if signer.VerifyToken(token) {
			t.Errorf("malformed token %q verified but should not have", token)
		}
		if !signer.Expired(token, 60) {
			t.Errorf("malformed token %q not treated as expired", token)
		}
	}
}

func TestSigner_FreshTokenNotExpired(t *testing.T) {
	token := signer.GenerateTokenFromString("https://example.com/reset/abc")

	if signer.Expired(token, 60) {
		t.Error("fresh token reported as expired at 60 minutes")
	}
}

func TestSigner_OldTokenExpired(t *testing.T) {
	// Hand-craft a correctly signed token whose timestamp is two hours old.
	prefix := "https://example.com/reset/abc?" + hashParam

	ts := make([]byte, tsLen)
	binary.BigEndian.PutUint64(ts, uint64(time.Now().Add(-2*time.Hour).Unix()))

	h := hmac.New(sha256.New, signer.Secret)
	h.Write([]byte(prefix))
	h.Write(ts)
	token := prefix + base64.RawURLEncoding.EncodeToString(append(ts, h.Sum(nil)...))

	if !signer.VerifyToken(token) {
		t.Error("hand-crafted old token did not verify")
	}

	if !signer.Expired(token, 60) {
		t.Error("two-hour-old token not reported as expired at 60 minutes")
	}
}
