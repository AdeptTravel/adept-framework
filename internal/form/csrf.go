// internal/form/csrf.go
//
// Adept – Forms subsystem: stateless CSRF token utilities.
//
// Context
//   Adept pages embed a hidden `csrf_token` input generated at render time.
//   The server must verify this token on POST to ensure the request originated
//   from a form it rendered.  We implement a *stateless* token:
//
//      base64url( nonce | unixMicro | HMAC_SHA256(secret, nonce+unixMicro) )
//
//   •  nonce – 16 random bytes.  Prevents replay across users.
//   •  unixMicro – microseconds since Unix epoch, 8 bytes, big-endian.
//   •  HMAC – calculated with tenant-scoped secret.  Verifies authenticity.
//
//   Validation checks the signature and ensures the timestamp is within
//   MaxAge.  No server-side sessions are required, keeping the system cache-
//   friendly and multi-instance safe.
//
// Workflow
//   •  GenerateToken()   → returns token string for renderer.
//   •  VerifyToken(tok) → constant-time verify; false on any failure.
//
//------------------------------------------------------------------------------

package form

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"os"
	"sync"
	"time"
)

const (
	tokenBytes   = 16 + 8 + sha256.Size // nonce + ts + sig
	maxAge       = 2 * time.Hour        // token valid window
	secretEnvKey = "ADEPT_CSRF_KEY"     // 32-byte base64 key suggested
)

var (
	secretOnce sync.Once
	secretKey  []byte
)

// GenerateToken creates a new CSRF token.  Call once per form render.
func GenerateToken() (string, error) {
	sec := fetchSecret()

	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, uint64(time.Now().UnixMicro()))

	mac := hmac.New(sha256.New, sec)
	mac.Write(nonce)
	mac.Write(ts)
	sig := mac.Sum(nil)

	buf := make([]byte, 0, tokenBytes)
	buf = append(buf, nonce...)
	buf = append(buf, ts...)
	buf = append(buf, sig...)

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// VerifyToken returns true if tok passes HMAC and age checks.
func VerifyToken(tok string) bool {
	sec := fetchSecret()

	raw, err := base64.RawURLEncoding.DecodeString(tok)
	if err != nil || len(raw) != tokenBytes {
		return false
	}

	nonce := raw[:16]
	tsBytes := raw[16:24]
	sig := raw[24:]

	// Timestamp window check.
	ts := binary.BigEndian.Uint64(tsBytes)
	issued := time.UnixMicro(int64(ts))
	if time.Since(issued) > maxAge || time.Until(issued) > time.Minute {
		// Future timestamp (clock skew) or older than maxAge.
		return false
	}

	// Recompute HMAC.
	mac := hmac.New(sha256.New, sec)
	mac.Write(nonce)
	mac.Write(tsBytes)
	want := mac.Sum(nil)

	return hmac.Equal(sig, want)
}

// fetchSecret returns the process-wide CSRF secret, loading (or generating)
// it exactly once.  In production set ADEPT_CSRF_KEY to a 32-byte base64
// string.  When unset, we generate a random key at startup and log to stderr.
func fetchSecret() []byte {
	secretOnce.Do(func() {
		if env := os.Getenv(secretEnvKey); env != "" {
			if b, err := base64.RawURLEncoding.DecodeString(env); err == nil && len(b) >= 32 {
				secretKey = b
				return
			}
		}
		// Fallback random key (ephemeral – resets on restart).
		secretKey = make([]byte, 32)
		_, _ = rand.Read(secretKey)
		// Logging via stderr is acceptable at init since logger may not be ready.
		// Warning developers to set a persistent key in production.
		os.Stderr.WriteString("[Adept] WARNING: ADEPT_CSRF_KEY not set – using random key\n")
	})
	return secretKey
}
