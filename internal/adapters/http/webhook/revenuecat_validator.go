package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	SignatureHeader = "X-RevenueCat-Signature"
)

var (
	ErrMissingSignature = errors.New("missing signature header")
	ErrInvalidSignature = errors.New("invalid signature")
)

// VerifyRevenueCatSignature is a pure function that verifies the HMAC-SHA256 signature
// body: raw request body
// signatureHeader: value of X-RevenueCat-Signature header (expected format: "sha256=HEX")
// secrets: list of accepted secrets (to support rotation)
func VerifyRevenueCatSignature(body []byte, signatureHeader string, secrets []string) error {
	if strings.TrimSpace(signatureHeader) == "" {
		return ErrMissingSignature
	}

	// Expect format: sha256=HEX
	parts := strings.SplitN(signatureHeader, "=", 2)
	if len(parts) != 2 {
		return ErrInvalidSignature
	}
	algo := strings.TrimSpace(parts[0])
	hexSig := strings.TrimSpace(parts[1])
	if algo != "sha256" || hexSig == "" {
		return ErrInvalidSignature
	}

	// Compare against each secret (rotation support)
	for _, secret := range secrets {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expected := hex.EncodeToString(mac.Sum(nil))
		// constant time compare on bytes
		if subtle.ConstantTimeCompare([]byte(hexSig), []byte(expected)) == 1 {
			return nil
		}
	}

	return ErrInvalidSignature
}

// ValidateWebhook is a small wrapper that integrates with fiber.Ctx
func ValidateWebhook(c *fiber.Ctx, secrets []string) error {
	signature := c.Get(SignatureHeader)
	if strings.TrimSpace(signature) == "" {
		return ErrMissingSignature
	}
	body := c.Body()
	return VerifyRevenueCatSignature(body, signature, secrets)
}
