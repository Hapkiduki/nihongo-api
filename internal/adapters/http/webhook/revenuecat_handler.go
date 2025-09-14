package webhook

import (
	"errors"
	"nihongo-api/internal/application/service"
	"nihongo-api/internal/ports"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type RevenueCatHandler struct {
	eventProcessor ports.EventProcessor
	secrets        []string
	logger         zerolog.Logger
}

func NewRevenueCatHandler(eventProcessor ports.EventProcessor, secrets []string, logger zerolog.Logger) *RevenueCatHandler {
	return &RevenueCatHandler{
		eventProcessor: eventProcessor,
		secrets:        secrets,
		logger:         logger,
	}
}

func (h *RevenueCatHandler) Handle(c *fiber.Ctx) error {
	h.logger.Info().Msg("Received RevenueCat webhook")

	// Validar HMAC
	if err := ValidateWebhook(c, h.secrets); err != nil {
		h.logger.Error().Err(err).Msg("Invalid webhook signature")
		// Differentiate missing vs invalid for clearer logs but both 401 for clients
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid webhook signature"})
	}

	// Parsear body a RevenueCatEvent
	var event ports.RevenueCatEvent
	if err := c.BodyParser(&event); err != nil {
		h.logger.Error().Err(err).Msg("Invalid JSON payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	h.logger.Info().Str("event_id", event.ID).Str("type", event.Type).Msg("Parsed event")

	// Procesar evento
	if err := h.eventProcessor.ProcessEvent(c.Context(), &event); err != nil {
		h.logger.Error().Err(err).Str("event_id", event.ID).Msg("Error processing event")
		// Classify errors from service: if it's already processed -> 200, if transient -> 5xx to allow retries
		if errors.Is(err, service.ErrAlreadyProcessed) {
			return c.SendStatus(fiber.StatusOK)
		}
		if errors.Is(err, service.ErrTransient) {
			return c.Status(fiber.StatusInternalServerError).SendString("temporary error")
		}
		// default: return 400 Bad Request for domain/permanent errors
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	h.logger.Info().Str("event_id", event.ID).Msg("Event processed successfully")
	return c.SendStatus(fiber.StatusOK)
}
