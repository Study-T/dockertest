// Package webhook provides Yuntu webhook signature verification,
// AES-CBC decryption, and envelope parsing for the tracking service.
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

const defaultReplayWindowMs = 300_000 // 5 minutes

var (
	ErrMissingTimestamp   = errors.New("missing timestamp header")
	ErrMissingSignature   = errors.New("missing signature header")
	ErrInvalidTimestamp   = errors.New("invalid timestamp format")
	ErrExceedReplay       = errors.New("request timestamp outside replay window")
	ErrSignatureMismatch  = errors.New("signature verification failed")
)

// Verifier handles Yuntu webhook signature verification.
type Verifier struct {
	encryptKey   string
	replayWindow int64
}

// NewVerifier creates a verifier with the given config.
func NewVerifier(encryptKey string, replayWindow int64) *Verifier {
	return &Verifier{encryptKey: encryptKey, replayWindow: replayWindow}
}

// Verify checks the webhook signature.
// Algorithm: SHA256(timestamp + encryptKey + rawBody)
// Bypasses verification when encryptKey is empty (dev mode).
func (v *Verifier) Verify(timestamp, signature string, body []byte) error {
	if v.encryptKey == "" {
		return nil // dev mode: skip verification
	}
	if timestamp == "" {
		return ErrMissingTimestamp
	}
	if signature == "" {
		return ErrMissingSignature
	}

	tsMs, err := parseTimestampMs(timestamp)
	if err != nil {
		return err
	}

	if err := checkReplayWindow(tsMs, v.replayWindow); err != nil {
		return err
	}

	expected := computeSignature(timestamp, v.encryptKey, body)
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return ErrSignatureMismatch
	}
	return nil
}

func parseTimestampMs(timestamp string) (int64, error) {
	if timestamp == "" {
		return 0, ErrInvalidTimestamp
	}
	if len(timestamp) > 15 {
		return 0, ErrInvalidTimestamp
	}
	var ts int64
	for _, c := range timestamp {
		if c < '0' || c > '9' {
			return 0, ErrInvalidTimestamp
		}
		ts = ts*10 + int64(c-'0')
	}
	return ts, nil
}

func checkReplayWindow(tsMs, windowMs int64) error {
	if windowMs <= 0 {
		windowMs = defaultReplayWindowMs
	}
	nowMs := time.Now().UnixMilli()
	if tsMs > nowMs {
		return ErrExceedReplay
	}
	if nowMs-tsMs > windowMs {
		return ErrExceedReplay
	}
	return nil
}

func computeSignature(timestamp, encryptKey string, rawBody []byte) string {
	h := sha256.New()
	h.Write([]byte(timestamp))
	h.Write([]byte(encryptKey))
	h.Write(rawBody)
	return hex.EncodeToString(h.Sum(nil))
}
