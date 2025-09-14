package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestValidateWebhook(t *testing.T) {
	secret := "test_secret"
	body := []byte(`{"id": "test", "type": "INITIAL_PURCHASE", "app_user_id": "user1"}`)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	expectedSignature := "sha256=" + hex.EncodeToString(h.Sum(nil))

	tests := []struct {
		name        string
		signature   string
		body        []byte
		secret      string
		expectError bool
		expectedMsg string
	}{
		{
			name:        "Valid signature",
			signature:   expectedSignature,
			body:        body,
			secret:      secret,
			expectError: false,
		},
		{
			name:        "Missing signature header",
			body:        body,
			secret:      secret,
			expectError: true,
			expectedMsg: "missing signature header",
		},
		{
			name:        "Invalid signature",
			signature:   "sha256=invalid",
			body:        body,
			secret:      secret,
			expectError: true,
			expectedMsg: "invalid signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			// Handler que llama ValidateWebhook y retorna error si corresponde
			app.Post("/webhook", func(c *fiber.Ctx) error {
				err := ValidateWebhook(c, []string{tt.secret})
				if err != nil {
					return c.Status(fiber.StatusUnauthorized).SendString(err.Error())
				}
				return c.SendStatus(fiber.StatusOK)
			})

			req, _ := http.NewRequest("POST", "/webhook", bytes.NewReader(tt.body))
			if tt.signature != "" {
				req.Header.Set("X-RevenueCat-Signature", tt.signature)
			}

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("app.Test error: %v", err)
			}

			if tt.expectError {
				assert.NotEqual(t, fiber.StatusOK, resp.StatusCode)
				buf := new(bytes.Buffer)
				buf.ReadFrom(resp.Body)
				assert.Contains(t, buf.String(), tt.expectedMsg)
			} else {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			}
		})
	}
}
