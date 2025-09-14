package service

import (
	"context"
	"errors"
	"nihongo-api/internal/domain"
	"nihongo-api/internal/ports"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/rs/zerolog"
)

// UserService handles user business logic
type UserService struct {
	userRepo ports.UserRepository
	subRepo  ports.SubscriptionRepository
	logger   zerolog.Logger
}

// NewUserService creates a new user service
func NewUserService(userRepo ports.UserRepository, subRepo ports.SubscriptionRepository, logger zerolog.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		subRepo:  subRepo,
		logger:   logger,
	}
}

// RegisterUser creates a new user
func (s *UserService) RegisterUser(ctx context.Context, name, email, password string) (*domain.User, error) {
	s.logger.Info().Str("email", email).Msg("Registering new user")

	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, email)
	if existingUser != nil {
		s.logger.Warn().Str("email", email).Msg("User already exists")
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to hash password")
		return nil, err
	}

	user := &domain.User{
		Name:      name,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create user")
		return nil, err
	}

	s.logger.Info().Str("user_id", user.ID.Hex()).Msg("User registered successfully")
	return user, nil
}

// AuthenticateUser verifies user credentials
func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (*domain.User, error) {
	s.logger.Info().Str("email", email).Msg("Authenticating user")

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Warn().Str("email", email).Msg("Invalid credentials")
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		s.logger.Warn().Str("email", email).Msg("Invalid credentials")
		return nil, errors.New("invalid credentials")
	}

	s.logger.Info().Str("user_id", user.ID.Hex()).Msg("User authenticated successfully")
	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// SyncRevenueCatUser sincroniza un usuario con RevenueCat, crea si no existe y linkea suscripciones
func (s *UserService) SyncRevenueCatUser(ctx context.Context, revenueCatUserID, name, email, password string) (*domain.User, error) {
	s.logger.Info().Str("revenue_cat_user_id", revenueCatUserID).Msg("Syncing RevenueCat user")

	// Verificar si ya existe por revenueCatID
	user, err := s.userRepo.GetByRevenueCatID(ctx, revenueCatUserID)
	if err == nil {
		s.logger.Debug().Str("user_id", user.ID.Hex()).Msg("User already exists")
		// Ya existe, solo actualizar si necesario
		return user, nil
	}

	s.logger.Info().Str("email", email).Msg("Creating new user for RevenueCat ID")

	// Si no existe, crear nuevo user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to hash password")
		return nil, err
	}

	user = &domain.User{
		ID:               primitive.NewObjectID(),
		Name:             name,
		Email:            email,
		Password:         string(hashedPassword),
		RevenueCatUserID: revenueCatUserID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create user")
		return nil, err
	}

	// Linkear suscripciones huérfanas
	linkErr := s.subRepo.UpdateInternalUserID(ctx, revenueCatUserID, user.ID.Hex())
	if linkErr != nil {
		s.logger.Warn().Err(linkErr).Str("revenue_cat_user_id", revenueCatUserID).Msg("Failed to link orphan subscriptions, but user created")
		// No fail creation
	} else {
		s.logger.Info().Str("user_id", user.ID.Hex()).Msg("Orphan subscriptions linked")
	}

	s.logger.Info().Str("user_id", user.ID.Hex()).Msg("RevenueCat user synced successfully")
	return user, nil
}
