package service

import (
	"testing"
)

func TestGrayscale_Disabled(t *testing.T) {
	gs := NewGrayscaleService(GrayscaleConfig{Enabled: false})
	if gs.ShouldProcessByGo("YT123") {
		t.Error("expected false when disabled")
	}
}

func TestGrayscale_AllMode(t *testing.T) {
	gs := NewGrayscaleService(GrayscaleConfig{Enabled: true, Mode: GrayscaleModeAll})
	if !gs.ShouldProcessByGo("YT123") {
		t.Error("expected true in all mode")
	}
}

func TestGrayscale_Whitelist(t *testing.T) {
	gs := NewGrayscaleService(GrayscaleConfig{
		Enabled:   true,
		Mode:      GrayscaleModeWhitelist,
		Whitelist: []string{"YT100", "YT200", "YT300"},
	})
	tests := []struct {
		id   string
		want bool
	}{
		{"YT100", true},
		{"YT200", true},
		{"YT999", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := gs.ShouldProcessByGo(tt.id); got != tt.want {
			t.Errorf("ShouldProcessByGo(%q) = %v, want %v", tt.id, got, tt.want)
		}
	}
}

func TestGrayscale_Percentage(t *testing.T) {
	gs := NewGrayscaleService(GrayscaleConfig{
		Enabled:    true,
		Mode:       GrayscaleModePercentage,
		Percentage: 50,
	})
	// With 50%, roughly half should pass
	passCount := 0
	for i := 0; i < 1000; i++ {
		if gs.ShouldProcessByGo("YT" + string(rune(i%10+'0'))) {
			passCount++
		}
	}
	// Allow some variance (30%-70%)
	if passCount < 300 || passCount > 700 {
		t.Errorf("expected ~500 pass, got %d", passCount)
	}
}

func TestGrayscale_PercentageZero(t *testing.T) {
	gs := NewGrayscaleService(GrayscaleConfig{
		Enabled:    true,
		Mode:       GrayscaleModePercentage,
		Percentage: 0,
	})
	if gs.ShouldProcessByGo("YT123") {
		t.Error("expected false with 0%")
	}
}

func TestGrayscale_PercentageHundred(t *testing.T) {
	gs := NewGrayscaleService(GrayscaleConfig{
		Enabled:    true,
		Mode:       GrayscaleModePercentage,
		Percentage: 100,
	})
	if !gs.ShouldProcessByGo("YT123") {
		t.Error("expected true with 100%")
	}
}

func TestGrayscale_UnknownMode(t *testing.T) {
	gs := NewGrayscaleService(GrayscaleConfig{
		Enabled: true,
		Mode:    "unknown",
	})
	if gs.ShouldProcessByGo("YT123") {
		t.Error("expected false for unknown mode")
	}
}

func TestGrayscale_WhitelistSize(t *testing.T) {
	gs := NewGrayscaleService(GrayscaleConfig{
		Whitelist: []string{"A", "B", "C"},
	})
	if got := gs.WhitelistSize(); got != 3 {
		t.Errorf("WhitelistSize() = %d, want 3", got)
	}
}
