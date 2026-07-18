// Package urlsigner signs URLs with an embedded timestamp and an HMAC-SHA256
// signature, so links such as password-reset URLs can be verified and expired
// without any server-side state.
package urlsigner

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"strings"
	"time"
)

// hashParam is the query parameter that carries the timestamp and signature.
const hashParam = "hash="

// tsLen is the length in bytes of the big-endian unix timestamp that prefixes
// the MAC inside the encoded hash payload.
const tsLen = 8

// Signer signs and verifies URLs using HMAC-SHA256 keyed with Secret.
type Signer struct {
	Secret []byte
}

// GenerateTokenFromString signs data (typically a URL) by appending a hash
// query parameter whose value is the base64 (raw URL) encoding of an 8-byte
// big-endian unix timestamp followed by an HMAC-SHA256 signature over the URL
// and that timestamp.
func (s *Signer) GenerateTokenFromString(data string) string {
	var prefix string
	if strings.Contains(data, "?") {
		prefix = data + "&" + hashParam
	} else {
		prefix = data + "?" + hashParam
	}

	ts := make([]byte, tsLen)
	binary.BigEndian.PutUint64(ts, uint64(time.Now().Unix()))

	payload := append(ts, s.mac(prefix, ts)...)

	return prefix + base64.RawURLEncoding.EncodeToString(payload)
}

// VerifyToken reports whether token was produced by GenerateTokenFromString
// with the same secret and has not been tampered with. Malformed tokens are
// reported as invalid; they never cause a panic.
func (s *Signer) VerifyToken(token string) bool {
	prefix, ts, mac, ok := parse(token)
	if !ok {
		return false
	}
	return hmac.Equal(mac, s.mac(prefix, ts))
}

// Expired reports whether token was generated more than minutesUntilExpire
// minutes ago. Malformed tokens are treated as expired.
func (s *Signer) Expired(token string, minutesUntilExpire int) bool {
	_, ts, _, ok := parse(token)
	if !ok {
		return true
	}

	created := time.Unix(int64(binary.BigEndian.Uint64(ts)), 0)

	return time.Since(created) > time.Duration(minutesUntilExpire)*time.Minute
}

// mac computes the HMAC-SHA256 of prefix followed by the raw timestamp bytes.
func (s *Signer) mac(prefix string, ts []byte) []byte {
	h := hmac.New(sha256.New, s.Secret)
	h.Write([]byte(prefix))
	h.Write(ts)
	return h.Sum(nil)
}

// parse splits a token into the signed prefix (everything up to and including
// the final hash parameter), the timestamp bytes, and the MAC. ok is false for
// any malformed token.
func parse(token string) (prefix string, ts, mac []byte, ok bool) {
	idx := strings.LastIndex(token, hashParam)
	if idx == -1 {
		return "", nil, nil, false
	}
	prefix = token[:idx+len(hashParam)]

	payload, err := base64.RawURLEncoding.DecodeString(token[idx+len(hashParam):])
	if err != nil || len(payload) != tsLen+sha256.Size {
		return "", nil, nil, false
	}

	return prefix, payload[:tsLen], payload[tsLen:], true
}
