package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"nihongo-api/internal/domain"
	"nihongo-api/internal/ports"

	"github.com/rs/zerolog"
)

type UserSyncer interface {
	SyncRevenueCatUser(ctx context.Context, revenueCatUserID, name, email, password string) (*domain.User, error)
}

type SubscriptionService struct {
	subRepo  ports.SubscriptionRepository
	userRepo ports.UserRepository
	userSvc  UserSyncer
	logger   zerolog.Logger
	// Opcional: progressService ports.ProgressService para actualizar acceso premium
}

func NewSubscriptionService(subRepo ports.SubscriptionRepository, userRepo ports.UserRepository, userSvc UserSyncer, logger zerolog.Logger) *SubscriptionService {
	return &SubscriptionService{
		subRepo:  subRepo,
		userRepo: userRepo,
		userSvc:  userSvc,
		logger:   logger,
	}
}

// ProcessEvent procesa un evento de RevenueCat, verifica idempotencia, actualiza suscripción
func (s *SubscriptionService) ProcessEvent(ctx context.Context, event *ports.RevenueCatEvent) error {
	s.logger.Info().
		Str("event_id", event.ID).
		Str("type", event.Type).
		Str("app_user_id", event.AppUserID).
		Str("store", event.Store).
		Str("product_id", event.ProductID).
		Str("entitlement_id", event.EntitlementID).
		Msg("Processing RevenueCat event")

	event.Timestamp = time.UnixMilli(event.EventTimestampMs)

	// Validar environment (e.g., solo production, configurar via flag si needed)
	if event.Environment != "PRODUCTION" {
		s.logger.Warn().Str("environment", event.Environment).Msg("Ignoring non-production event")
		return nil
	}

	// Verificar si es producto premium (ejemplo: product_id contiene "premium" o entitlement)
	if !isPremiumProduct(event.ProductID, event.EntitlementID) {
		s.logger.Debug().Str("product_id", event.ProductID).Msg("Ignoring non-premium event")
		return nil // Ignorar eventos no premium
	}

	// Check idempotencia
	existingSub, err := s.subRepo.GetByEventID(ctx, event.ID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error checking idempotency")
		return ErrTransient
	}
	if existingSub != nil {
		s.logger.Info().Str("event_id", event.ID).Msg("Event already processed (idempotent)")
		return nil // Ya procesado (idempotent)
	}

	// If event type is unsupported, return early to avoid unnecessary repo calls
	switch event.Type {
	case "INITIAL_PURCHASE", "RENEWAL", "UNCANCELLATION", "CANCELLATION", "REFUND", "TRANSFER", "BILLING_ISSUE", "EXPIRATION", "PRODUCT_CHANGE":
		// supported types
	default:
		s.logger.Warn().Str("type", event.Type).Msg("Unsupported event type")
		return errors.New("unsupported event type")
	}

	// Buscar suscripción existente por external_user_id
	subs, err := s.subRepo.GetByExternalUserID(ctx, event.AppUserID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Error getting existing subscriptions")
		return err
	}

	var sub *domain.Subscription
	var isNew bool
	if len(subs) > 0 {
		sub = subs[0] // Asumir principal; mejorar si multiple
		isNew = false
	} else {
		isNew = true
	}

	// ... user sync moved into purchase handling below

	switch event.Type {
	case "INITIAL_PURCHASE", "RENEWAL", "UNCANCELLATION", "PRODUCT_CHANGE":
		var expiresAt time.Time
		if event.ExpiresAtMs != nil {
			expiresAt = time.UnixMilli(*event.ExpiresAtMs)
		} else {
			// Fallback basado en product/entitlement (simplificado)
			expiresAt = event.Timestamp.Add(12 * 30 * 24 * time.Hour) // Anual default
		}
		// If new subscription or existing without linked user, sync/reconcile user
		if isNew || (sub != nil && sub.InternalUserID == nil) {
			name, okName := event.SubscriberAttributes["name"].(string)
			email, okEmail := event.SubscriberAttributes["email"].(string)
			if !okName || !okEmail {
				s.logger.Warn().Str("app_user_id", event.AppUserID).Msg("Missing attributes for user sync; using defaults")
				name = "Unknown"
				email = "unknown@example.com"
			}
			password := generateSecurePassword()
			_, syncErr := s.userSvc.SyncRevenueCatUser(ctx, event.AppUserID, name, email, password)
			if syncErr != nil {
				s.logger.Error().Err(syncErr).Msg("Failed to sync user")
				return ErrTransient
			}
			s.logger.Info().Str("app_user_id", event.AppUserID).Msg("User synced successfully")
			// Re-fetch sub para link
			subs, _ = s.subRepo.GetByExternalUserID(ctx, event.AppUserID)
			if len(subs) > 0 {
				sub = subs[0]
				isNew = false
			}
		}

		if isNew {
			sub = domain.NewSubscription(event.AppUserID, event.ProductID, event.ID, event.Type, expiresAt)
		} else {
			// For PRODUCT_CHANGE, update product_id
			if event.Type == "PRODUCT_CHANGE" {
				sub.ProductID = event.ProductID
			}
			sub.UpdateStatus(domain.SubscriptionActive, expiresAt)
		}
		sub.EventType = event.Type
	case "CANCELLATION":
		if sub == nil {
			return errors.New("no subscription found for cancellation")
		}
		sub.UpdateStatus(domain.SubscriptionCancelled, time.Time{})
		sub.EventType = event.Type
	case "REFUND":
		if sub == nil {
			return errors.New("no subscription found for refund")
		}
		sub.UpdateStatus(domain.SubscriptionExpired, time.Time{})
		sub.EventType = event.Type
		s.logger.Info().Str("refund_reason", event.RefundReason).Msg("Subscription refunded")
	case "TRANSFER":
		// Log transfer; link si new attributes
		s.logger.Info().Str("app_user_id", event.AppUserID).Msg("Subscription transferred; check linking")
		return nil // No update sub, solo log
	case "BILLING_ISSUE", "EXPIRATION":
		if sub != nil {
			sub.UpdateStatus(domain.SubscriptionExpired, time.Time{})
			sub.EventType = event.Type
		}
		s.logger.Warn().Str("type", event.Type).Msg("Billing issue or expiration handled")
	default:
		s.logger.Warn().Str("type", event.Type).Msg("Unsupported event type")
		return errors.New("unsupported event type")
	}

	// Si active y linked, TODO: update progress
	if sub != nil && sub.Status == domain.SubscriptionActive && sub.InternalUserID != nil {
		s.logger.Debug().Str("user_id", sub.InternalUserID.Hex()).Msg("Active subscription; grant premium access if needed")
		// TODO: progressService.GrantPremiumAccess(ctx, sub.InternalUserID.Hex())
	}

	// Persistir con retry simple (3 attempts)
	err = retryDBOp(3, func() error {
		if isNew || sub.ID.IsZero() {
			return s.subRepo.Create(ctx, sub)
		}
		return s.subRepo.UpdateByEventID(ctx, event.ID, sub)
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to persist subscription")
		// If DB reported duplicate (race), treat as already processed
		if errors.Is(err, ports.ErrDuplicateEvent) {
			return nil
		}
		return ErrTransient
	}

	s.logger.Info().Str("event_id", event.ID).Str("status", string(sub.Status)).Msg("Event processed successfully")
	return nil
}

var (
	ErrAlreadyProcessed = errors.New("already processed")
	ErrTransient        = errors.New("transient error")
)

func isPremiumProduct(productID, entitlementID string) bool {
	// Basado en product o entitlement
	premiumProducts := []string{"premium_monthly", "premium_yearly"}
	for _, p := range premiumProducts {
		if productID == p || entitlementID == p {
			return true
		}
	}
	return false
}

// retryDBOp simple exponential backoff
func retryDBOp(maxRetries int, op func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = op()
		if err == nil {
			return nil
		}
		// Backoff: 100ms * 2^i
		time.Sleep(time.Duration(1<<i) * 100 * time.Millisecond)
	}
	return err
}

// generateSecurePassword generates a cryptographically secure random password
func generateSecurePassword() string {
	bytes := make([]byte, 16) // 32 hex chars
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based if crypto fails (unlikely)
		return hex.EncodeToString([]byte(time.Now().String()[:16]))
	}
	return hex.EncodeToString(bytes)
}
