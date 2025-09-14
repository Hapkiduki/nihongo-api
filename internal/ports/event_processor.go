package ports

import (
	"context"
	"time"
)

// RevenueCatEvent represents the incoming webhook payload from RevenueCat
// Based on: https://docs.revenuecat.com/reference/webhooks
type RevenueCatEvent struct {
	// Common fields
	ID                   string                 `json:"id"`
	Type                 string                 `json:"type"`
	AppUserID            string                 `json:"app_user_id"`
	ProductID            string                 `json:"product_id"`
	Store                string                 `json:"store,omitempty"` // "app_store", "play_store", "amazon", etc.
	Environment          string                 `json:"environment"`
	EventTimestampMs     int64                  `json:"event_timestamp_ms"`
	PurchasedAtMs        *int64                 `json:"purchased_at_ms,omitempty"`
	ExpiresAtMs          *int64                 `json:"expires_at_ms,omitempty"`
	SubscriberAttributes map[string]interface{} `json:"subscriber_attributes,omitempty"`

	// Product/Price fields
	Price                    *float64 `json:"price,omitempty"`
	Currency                 string   `json:"currency,omitempty"`
	PriceInPurchasedCurrency *float64 `json:"price_in_purchased_currency,omitempty"`

	// Transaction fields
	TransactionID         string `json:"transaction_id,omitempty"`
	OriginalTransactionID string `json:"original_transaction_id,omitempty"`
	PeriodType            string `json:"period_type,omitempty"` // "normal", "intro", "trial"

	// Entitlement fields
	EntitlementID  string   `json:"entitlement_id,omitempty"`
	EntitlementIds []string `json:"entitlement_ids,omitempty"`

	// Offering/Presentation fields
	PresentedOfferingID string `json:"presented_offering_id,omitempty"`

	// Cancellation/Refund fields
	CancelReason string `json:"cancel_reason,omitempty"`
	RefundReason string `json:"refund_reason,omitempty"`

	// Billing issue fields
	BillingIssueDetectedAtMs *int64 `json:"billing_issue_detected_at_ms,omitempty"`

	// Internal processing
	Timestamp time.Time `json:"-"`
}

// EventProcessor defines the interface for processing RevenueCat webhook events
type EventProcessor interface {
	ProcessEvent(ctx context.Context, event *RevenueCatEvent) error
}
