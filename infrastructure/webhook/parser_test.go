package webhook

import (
	"encoding/base64"
	"testing"
)

func TestParseEnvelope_DirectBody(t *testing.T) {
	raw := []byte(`{"data":{"waybill_number":"YT123"},"data_code":"tisPushData"}`)
	env, err := parseEnvelope(raw, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.IsEncrypted || env.IsStandard {
		t.Fatal("expected direct body")
	}
}

func TestParseEnvelope_StandardEnvelope(t *testing.T) {
	raw := []byte(`{"customerCode":"CN","timestamp":"123","body":{"data_code":"tisPushData"}}`)
	env, err := parseEnvelope(raw, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !env.IsStandard {
		t.Fatal("expected standard")
	}
}

func TestParseEnvelope_NilBody(t *testing.T) {
	raw := []byte(`{"customerCode":"CN","timestamp":"123","body":null}`)
	_, err := parseEnvelope(raw, "")
	if err != ErrEmptyBody {
		t.Fatalf("expected ErrEmptyBody, got: %v", err)
	}
}

func TestParseEnvelope_EmptyBody(t *testing.T) {
	_, err := parseEnvelope([]byte{}, "")
	if err != ErrEmptyBody {
		t.Fatalf("expected ErrEmptyBody, got: %v", err)
	}
}

func TestParseEnvelope_InvalidJSON(t *testing.T) {
	_, err := parseEnvelope([]byte(`{not json}`), "")
	if err != ErrInvalidJSON {
		t.Fatalf("expected ErrInvalidJSON, got: %v", err)
	}
}

func TestParseEnvelope_EncryptedPayload(t *testing.T) {
	inner := `{"data_code":"tisPushData","waybill_number":"YT456"}`
	encrypted := encryptForTest(t, []byte(inner), "test-key")
	encoded := base64.StdEncoding.EncodeToString(encrypted)
	raw := []byte(`{"encrypt":"` + encoded + `"}`)

	env, err := parseEnvelope(raw, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !env.IsEncrypted {
		t.Fatal("expected IsEncrypted=true")
	}
}

func TestParseEnvelope_EncryptedStandardEnvelope(t *testing.T) {
	inner := `{"customerCode":"CN","timestamp":"123","body":{"data_code":"tisPushData","tracking_number":"777"}}`
	encrypted := encryptForTest(t, []byte(inner), "test-key")
	encoded := base64.StdEncoding.EncodeToString(encrypted)
	raw := []byte(`{"encrypt":"` + encoded + `"}`)

	env, err := parseEnvelope(raw, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !env.IsEncrypted || !env.IsStandard {
		t.Fatalf("expected encrypted+standard, got enc=%v std=%v", env.IsEncrypted, env.IsStandard)
	}
}

func TestParseEnvelope_FallbackToDirect(t *testing.T) {
	raw := []byte(`{"unknown":"value"}`)
	env, err := parseEnvelope(raw, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.IsStandard || env.IsEncrypted {
		t.Fatal("expected direct fallback")
	}
}

func TestResolveBodyEnvelope(t *testing.T) {
	tests := []struct {
		name        string
		raw         map[string]interface{}
		isEncrypted bool
		wantStd     bool
		wantEnc     bool
		wantErr     bool
	}{
		{"direct", map[string]interface{}{"key": "val"}, false, false, false, false},
		{"standard", map[string]interface{}{"body": map[string]interface{}{"key": "val"}}, false, true, false, false},
		{"encrypted direct", map[string]interface{}{"key": "val"}, true, false, true, false},
		{"encrypted standard", map[string]interface{}{"body": map[string]interface{}{"key": "val"}}, true, true, true, false},
		{"nil body", map[string]interface{}{"body": nil}, false, false, false, true},
		{"empty body map", map[string]interface{}{"body": map[string]interface{}{}}, false, false, false, true},
		{"wrong body type", map[string]interface{}{"body": "string"}, false, false, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := resolveBodyEnvelope(tt.raw, tt.isEncrypted)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if env.IsStandard != tt.wantStd {
				t.Errorf("IsStandard=%v, want %v", env.IsStandard, tt.wantStd)
			}
			if env.IsEncrypted != tt.wantEnc {
				t.Errorf("IsEncrypted=%v, want %v", env.IsEncrypted, tt.wantEnc)
			}
		})
	}
}
