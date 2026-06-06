package webhook

import (
	"fmt"
	"testing"
	"time"
)

func currentTimestampMs() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}

func TestVerifier_Verify_Valid(t *testing.T) {
	v := NewVerifier("test-key", 300_000)
	body := []byte(`{"data_code":"tisPushData"}`)
	ts := currentTimestampMs()
	sig := computeSignature(ts, "test-key", body)
	if err := v.Verify(ts, sig, body); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestVerifier_Verify_InvalidSignature(t *testing.T) {
	v := NewVerifier("test-key", 300_000)
	body := []byte(`{}`)
	ts := currentTimestampMs()
	if err := v.Verify(ts, "wrong", body); err != ErrSignatureMismatch {
		t.Fatalf("expected ErrSignatureMismatch, got: %v", err)
	}
}

func TestVerifier_Verify_MissingTimestamp(t *testing.T) {
	v := NewVerifier("key", 300_000)
	if err := v.Verify("", "sig", []byte{}); err != ErrMissingTimestamp {
		t.Fatalf("expected ErrMissingTimestamp, got: %v", err)
	}
}

func TestVerifier_Verify_MissingSignature(t *testing.T) {
	v := NewVerifier("key", 300_000)
	if err := v.Verify(currentTimestampMs(), "", []byte{}); err != ErrMissingSignature {
		t.Fatalf("expected ErrMissingSignature, got: %v", err)
	}
}

func TestVerifier_Verify_ReplayAttack(t *testing.T) {
	v := NewVerifier("key", 300_000)
	if err := v.Verify("946684800000", "sig", []byte{}); err != ErrExceedReplay {
		t.Fatalf("expected ErrExceedReplay, got: %v", err)
	}
}

func TestVerifier_Verify_FutureTimestamp(t *testing.T) {
	v := NewVerifier("key", 300_000)
	if err := v.Verify("4102444800000", "sig", []byte{}); err != ErrExceedReplay {
		t.Fatalf("expected ErrExceedReplay, got: %v", err)
	}
}

func TestVerifier_Verify_TamperedBody(t *testing.T) {
	v := NewVerifier("test-key", 300_000)
	ts := currentTimestampMs()
	sig := computeSignature(ts, "test-key", []byte(`original`))
	if err := v.Verify(ts, sig, []byte(`tampered`)); err != ErrSignatureMismatch {
		t.Fatalf("expected ErrSignatureMismatch, got: %v", err)
	}
}

func TestComputeSignature_Deterministic(t *testing.T) {
	sig1 := computeSignature("12345", "key", []byte(`test`))
	sig2 := computeSignature("12345", "key", []byte(`test`))
	if sig1 != sig2 {
		t.Fatal("expected deterministic output")
	}
}

func TestParseTimestampMs(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"1717312800000", 1717312800000, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
		{"99999999999999999", 0, true},
	}
	for _, tt := range tests {
		got, err := parseTimestampMs(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseTimestampMs(%q) err=%v, wantErr=%v", tt.input, err, tt.wantErr)
		}
		if got != tt.want {
			t.Errorf("parseTimestampMs(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
