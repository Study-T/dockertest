package webhook

import (
	"encoding/json"
	"errors"
)

var (
	ErrInvalidJSON = errors.New("invalid JSON payload")
	ErrEmptyBody   = errors.New("envelope body is empty")
)

// Envelope represents a parsed Yuntu webhook envelope.
type Envelope struct {
	Body        map[string]interface{}
	IsEncrypted bool
	IsStandard  bool
}

// Parser handles Yuntu webhook envelope parsing.
type Parser struct{}

// NewParser creates a new parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses a raw webhook body through three envelope formats:
//  1. AES-CBC encrypted: {"encrypt": "base64..."} → decrypt → re-parse
//  2. Standard envelope: {"customerCode", "timestamp", "body": {...}}
//  3. Direct body: {"data": {...}, "data_code": "tisPushData"}
func (p *Parser) Parse(rawBody []byte, encryptKey string) (*Envelope, error) {
	return parseEnvelope(rawBody, encryptKey)
}

func parseEnvelope(rawBody []byte, encryptKey string) (*Envelope, error) {
	if len(rawBody) == 0 {
		return nil, ErrEmptyBody
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(rawBody, &raw); err != nil {
		return nil, ErrInvalidJSON
	}

	// Priority 1: encrypted payload
	if enc, ok := raw["encrypt"].(string); ok && enc != "" {
		return parseEncryptedEnvelope(enc, encryptKey)
	}

	return resolveBodyEnvelope(raw, false)
}

func parseEncryptedEnvelope(base64Ciphertext, encryptKey string) (*Envelope, error) {
	plaintext, err := DecryptAES(base64Ciphertext, encryptKey)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(plaintext, &raw); err != nil {
		return nil, ErrInvalidJSON
	}

	return resolveBodyEnvelope(raw, true)
}

func resolveBodyEnvelope(raw map[string]interface{}, isEncrypted bool) (*Envelope, error) {
	bodyVal, hasBody := raw["body"]
	if !hasBody {
		return &Envelope{Body: raw, IsEncrypted: isEncrypted, IsStandard: false}, nil
	}
	if bodyVal == nil {
		return nil, ErrEmptyBody
	}

	bodyMap, ok := bodyVal.(map[string]interface{})
	if !ok || len(bodyMap) == 0 {
		return nil, ErrEmptyBody
	}

	return &Envelope{Body: bodyMap, IsEncrypted: isEncrypted, IsStandard: true}, nil
}
