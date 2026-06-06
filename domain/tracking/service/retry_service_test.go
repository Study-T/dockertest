package service

import (
	"testing"
	"time"
)

func TestShouldRetry_RetryableEvent(t *testing.T) {
	s := &RetryService{}
	evt := &RawEventForTest{
		RetryCount: 2,
		MaxRetries: 5,
		UpdatedAt:  time.Now().Add(-10 * time.Minute), // well past backoff
	}
	if !s.shouldRetryFromTest(evt.RetryCount, evt.MaxRetries, evt.UpdatedAt) {
		t.Error("expected shouldRetry=true")
	}
}

func TestShouldRetry_ExceededRetries(t *testing.T) {
	s := &RetryService{}
	if s.shouldRetryFromTest(5, 5, time.Now().Add(-1*time.Hour)) {
		t.Error("expected shouldRetry=false when retries exceeded")
	}
}

func TestShouldRetry_TooSoon(t *testing.T) {
	s := &RetryService{}
	// retry_count=0, backoff=30s, but updated 1s ago
	if s.shouldRetryFromTest(0, 5, time.Now().Add(-1*time.Second)) {
		t.Error("expected shouldRetry=false when too soon")
	}
}

func TestShouldRetry_BackoffSchedule(t *testing.T) {
	s := &RetryService{}
	tests := []struct {
		retryCount int
		elapsed    time.Duration
		want       bool
	}{
		{0, 31 * time.Second, true},   // 30s backoff, waited 31s
		{0, 29 * time.Second, false},   // 30s backoff, waited 29s
		{1, 121 * time.Second, true},   // 120s backoff
		{2, 601 * time.Second, true},   // 600s backoff
		{3, 3601 * time.Second, true},  // 3600s backoff
		{4, 3601 * time.Second, true},  // max backoff (capped at 3600s)
	}
	for _, tt := range tests {
		got := s.shouldRetryFromTest(tt.retryCount, 5, time.Now().Add(-tt.elapsed))
		if got != tt.want {
			t.Errorf("retryCount=%d elapsed=%v: got %v, want %v", tt.retryCount, tt.elapsed, got, tt.want)
		}
	}
}

// Helper to test shouldRetry without needing entity import
func (s *RetryService) shouldRetryFromTest(retryCount, maxRetries int, updatedAt time.Time) bool {
	if retryCount >= maxRetries {
		return false
	}
	backoffIdx := retryCount
	if backoffIdx >= len(backoffSchedule) {
		backoffIdx = len(backoffSchedule) - 1
	}
	return time.Since(updatedAt) >= backoffSchedule[backoffIdx]
}

type RawEventForTest struct {
	RetryCount int
	MaxRetries int
	UpdatedAt  time.Time
}

func TestParseJSONToMap(t *testing.T) {
	tests := []struct {
		input   string
		wantNil bool
		wantKey string
	}{
		{`{"key":"value"}`, false, "key"},
		{`{not json}`, true, ""},
		{"", true, ""},
		{`{"nested":{"inner":1}}`, false, "nested"},
	}
	for _, tt := range tests {
		got := parseJSONToMap(tt.input)
		if tt.wantNil {
			if got != nil {
				t.Errorf("parseJSONToMap(%q) expected nil", tt.input)
			}
			continue
		}
		if got == nil {
			t.Errorf("parseJSONToMap(%q) expected non-nil", tt.input)
			continue
		}
		if _, ok := got[tt.wantKey]; !ok {
			t.Errorf("parseJSONToMap(%q) missing key %q", tt.input, tt.wantKey)
		}
	}
}

func TestBackoffSchedule(t *testing.T) {
	expected := []time.Duration{
		30 * time.Second,
		120 * time.Second,
		600 * time.Second,
		3600 * time.Second,
	}
	if len(backoffSchedule) != len(expected) {
		t.Fatalf("backoffSchedule length = %d, want %d", len(backoffSchedule), len(expected))
	}
	for i, got := range backoffSchedule {
		if got != expected[i] {
			t.Errorf("backoffSchedule[%d] = %v, want %v", i, got, expected[i])
		}
	}
}
