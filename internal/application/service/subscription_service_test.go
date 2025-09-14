// Package service provides tests for the subscription service, focusing on webhook processing.
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"nihongo-api/internal/domain"
	"nihongo-api/internal/ports"
)

type mockSubRepo struct {
	mock.Mock
}

func (m *mockSubRepo) Create(ctx context.Context, sub *domain.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *mockSubRepo) GetByEventID(ctx context.Context, eventID string) (*domain.Subscription, error) {
	args := m.Called(ctx, eventID)
	return args.Get(0).(*domain.Subscription), args.Error(1)
}

func (m *mockSubRepo) UpdateByEventID(ctx context.Context, eventID string, sub *domain.Subscription) error {
	args := m.Called(ctx, eventID, sub)
	return args.Error(0)
}

func (m *mockSubRepo) GetByExternalUserID(ctx context.Context, externalUserID string) ([]*domain.Subscription, error) {
	args := m.Called(ctx, externalUserID)
	return args.Get(0).([]*domain.Subscription), args.Error(1)
}

func (m *mockSubRepo) UpdateInternalUserID(ctx context.Context, externalUserID string, internalUserID string) error {
	args := m.Called(ctx, externalUserID, internalUserID)
	return args.Error(0)
}

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepo) LinkRevenueCatUserID(ctx context.Context, userID, revenueCatUserID string) error {
	args := m.Called(ctx, userID, revenueCatUserID)
	return args.Error(0)
}

func (m *mockUserRepo) GetByRevenueCatID(ctx context.Context, revenueCatID string) (*domain.User, error) {
	args := m.Called(ctx, revenueCatID)
	return args.Get(0).(*domain.User), args.Error(1)
}

type mockUserSvc struct {
	mock.Mock
}

func (m *mockUserSvc) RegisterUser(ctx context.Context, name, email, password string) (*domain.User, error) {
	args := m.Called(ctx, name, email, password)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserSvc) AuthenticateUser(ctx context.Context, email, password string) (*domain.User, error) {
	args := m.Called(ctx, email, password)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserSvc) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserSvc) SyncRevenueCatUser(ctx context.Context, revenueCatUserID, name, email, password string) (*domain.User, error) {
	args := m.Called(ctx, revenueCatUserID, name, email, password)
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestSubscriptionService_ProcessEvent(t *testing.T) {
	logger := zerolog.Nop()

	tests := []struct {
		name           string
		event          *ports.RevenueCatEvent
		setupMocks     func(*mockSubRepo, *mockUserRepo, *mockUserSvc)
		expectError    bool
		expectedStatus domain.SubscriptionStatus
	}{
		{
			name: "Ignore non-premium product",
			event: &ports.RevenueCatEvent{
				ProductID:        "non_premium",
				Environment:      "PRODUCTION",
				EventTimestampMs: 1234567890,
			},
			setupMocks:  func(sub *mockSubRepo, user *mockUserRepo, usvc *mockUserSvc) {},
			expectError: false,
		},
		{
			name: "Idempotent event already processed",
			event: &ports.RevenueCatEvent{
				ID:               "dup_id",
				Type:             "INITIAL_PURCHASE",
				AppUserID:        "user1",
				ProductID:        "premium_monthly",
				Environment:      "PRODUCTION",
				EventTimestampMs: 1234567890,
			},
			setupMocks: func(sub *mockSubRepo, user *mockUserRepo, usvc *mockUserSvc) {
				sub.On("GetByEventID", mock.Anything, "dup_id").Return(&domain.Subscription{}, nil)
			},
			expectError: false,
		},
		{
			name: "New purchase, user sync",
			event: &ports.RevenueCatEvent{
				ID:                   "new_id",
				Type:                 "INITIAL_PURCHASE",
				AppUserID:            "user1",
				ProductID:            "premium_monthly",
				Environment:          "PRODUCTION",
				EventTimestampMs:     1234567890,
				ExpiresAtMs:          ptr(1234567890000),
				SubscriberAttributes: map[string]interface{}{"name": "Test", "email": "test@example.com"},
			},
			setupMocks: func(sub *mockSubRepo, user *mockUserRepo, usvc *mockUserSvc) {
				sub.On("GetByEventID", mock.Anything, "new_id").Return((*domain.Subscription)(nil), nil)
				sub.On("GetByExternalUserID", mock.Anything, "user1").Return([]*domain.Subscription{}, nil)
				usvc.On("SyncRevenueCatUser", mock.Anything, "user1", "Test", "test@example.com", mock.AnythingOfType("string")).Return(&domain.User{ID: primitive.NewObjectID()}, nil)
				sub.On("Create", mock.Anything, mock.AnythingOfType("*domain.Subscription")).Return(nil)
			},
			expectError:    false,
			expectedStatus: domain.SubscriptionActive,
		},
		{
			name: "Cancellation, update status",
			event: &ports.RevenueCatEvent{
				ID:               "cancel_id",
				Type:             "CANCELLATION",
				AppUserID:        "user1",
				ProductID:        "premium_yearly",
				Environment:      "PRODUCTION",
				EventTimestampMs: 1234567890,
			},
			setupMocks: func(sub *mockSubRepo, user *mockUserRepo, usvc *mockUserSvc) {
				sub.On("GetByEventID", mock.Anything, "cancel_id").Return((*domain.Subscription)(nil), nil)
				sub.On("GetByExternalUserID", mock.Anything, "user1").Return([]*domain.Subscription{&domain.Subscription{ID: primitive.NewObjectID(), Status: domain.SubscriptionActive}}, nil)
				sub.On("UpdateByEventID", mock.Anything, "cancel_id", mock.AnythingOfType("*domain.Subscription")).Return(nil)
			},
			expectError:    false,
			expectedStatus: domain.SubscriptionCancelled,
		},
		{
			name: "Invalid environment (sandbox)",
			event: &ports.RevenueCatEvent{
				Type:             "INITIAL_PURCHASE",
				ProductID:        "premium_monthly",
				Environment:      "SANDBOX",
				EventTimestampMs: 1234567890,
			},
			setupMocks:  func(sub *mockSubRepo, user *mockUserRepo, usvc *mockUserSvc) {},
			expectError: false,
		},
		{
			name: "Unsupported event type",
			event: &ports.RevenueCatEvent{
				ID:               "unsupported_id",
				Type:             "UNKNOWN",
				AppUserID:        "user1",
				ProductID:        "premium_monthly",
				Environment:      "PRODUCTION",
				EventTimestampMs: 1234567890,
			},
			setupMocks: func(sub *mockSubRepo, user *mockUserRepo, usvc *mockUserSvc) {
				sub.On("GetByEventID", mock.Anything, "unsupported_id").Return((*domain.Subscription)(nil), nil)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subRepo := new(mockSubRepo)
			userRepo := new(mockUserRepo)
			userSvc := new(mockUserSvc)

			tt.setupMocks(subRepo, userRepo, userSvc)

			s := NewSubscriptionService(subRepo, userRepo, userSvc, logger)

			err := s.ProcessEvent(context.Background(), tt.event)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			subRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			userSvc.AssertExpectations(t)
		})
	}
}

func ptr(i int64) *int64 {
	return &i
}

func Test_isPremiumProduct(t *testing.T) {
	tests := []struct {
		name          string
		productID     string
		entitlementID string
		want          bool
	}{
		{"premium product", "premium_monthly", "", true},
		{"premium entitlement", "", "premium_yearly", true},
		{"non-premium", "basic", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isPremiumProduct(tt.productID, tt.entitlementID))
		})
	}
}

func Test_retryDBOp(t *testing.T) {
	// Test successful on first try
	successOp := func() error { return nil }
	err := retryDBOp(3, successOp)
	assert.NoError(t, err)

	// Test failure on all retries
	failOp := func() error { return errors.New("db error") }
	err = retryDBOp(3, failOp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")

	// Test success on second try (simulate retry)
	callCount := 0
	op := func() error {
		callCount++
		if callCount == 2 {
			return nil
		}
		return errors.New("transient error")
	}
	err = retryDBOp(3, op)
	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}
