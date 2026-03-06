package auth_usecase

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/domain"
	customErr "sso/internal/domain/errors"
	"sso/internal/lib/jwt"
	"time"

	"github.com/wb-go/wbf/zlog"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepo   userRepository
	appUsecase appUsecase
	logger     *zlog.Zerolog
	jwtSecret  string
	tokenTTL   time.Duration
}

func NewAuthUsecase(
	userRepo userRepository,
	appUsecase appUsecase,
	logger *zlog.Zerolog,
	jwtSecret string,
	tokenTTL time.Duration,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:   userRepo,
		appUsecase: appUsecase,
		logger:     logger,
		jwtSecret:  jwtSecret,
		tokenTTL:   tokenTTL,
	}
}

func (s *AuthUsecase) Login(ctx context.Context, username, password string, appID int) (*domain.UserClaim, string, time.Time, error) {
	const op = "auth_usecase.Login"

	if username == "" || password == "" || appID <= 0 {
		return nil, "", time.Time{}, customErr.ErrInvalidInput
	}

	s.logger.Info().Str("username", username).Int("app_id", appID).Msg("login attempt")

	if _, err := s.appUsecase.GetByID(ctx, appID); err != nil {
		s.logger.Warn().Int("app_id", appID).Msg("app not found")
		return nil, "", time.Time{}, customErr.ErrInvalidInput
	}

	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		s.logger.Warn().Err(err).Str("username", username).Msg("user not found")
		if errors.Is(err, customErr.ErrUserNotFound) {
			return nil, "", time.Time{}, customErr.ErrInvalidCredentials
		}
		return nil, "", time.Time{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn().Str("username", username).Msg("invalid password")
		return nil, "", time.Time{}, customErr.ErrInvalidCredentials
	}

	expiresAt := time.Now().Add(s.tokenTTL)
	token, err := jwt.NewToken(user, s.jwtSecret, s.tokenTTL)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to generate token")
		return nil, "", time.Time{}, fmt.Errorf("%s: %w", op, customErr.ErrInternal)
	}

	s.logger.Info().Str("username", username).Str("role", string(user.Role)).Msg("login successful")
	return &domain.UserClaim{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}, token, expiresAt, nil
}

func (s *AuthUsecase) RegisterNewUser(ctx context.Context, username, password string, role domain.UserRole, appID int) (int64, error) {
	const op = "auth_usecase.RegisterNewUser"

	if username == "" || password == "" || appID <= 0 {
		return 0, customErr.ErrInvalidInput
	}

	if _, err := s.appUsecase.GetByID(ctx, appID); err != nil {
		return 0, customErr.ErrInvalidInput
	}

	if !domain.IsValidRole(string(role)) {
		role = domain.RoleViewer
	}

	s.logger.Info().Str("username", username).Str("role", string(role)).Int("app_id", appID).Msg("register attempt")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to hash password")
		return 0, fmt.Errorf("%s: %w", op, customErr.ErrInternal)
	}

	user := &domain.User{
		Username:     username,
		PasswordHash: string(passHash),
		Role:         role,
	}

	id, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		s.logger.Error().Err(err).Str("username", username).Msg("failed to create user")
		if errors.Is(err, customErr.ErrUserExists) {
			return 0, customErr.ErrUserExists
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	s.logger.Info().Int64("user_id", id).Str("username", username).Msg("user registered")
	return id, nil
}

func (s *AuthUsecase) GetUsers(ctx context.Context) ([]*domain.User, error) {
	const op = "auth_usecase.GetUsers"
	s.logger.Info().Msg("getting users")
	users, err := s.userRepo.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	s.logger.Info().Int("count", len(users)).Msg("users retrieved")
	return users, nil
}

func (s *AuthUsecase) UpdateUserRole(ctx context.Context, userID int64, role domain.UserRole) error {
	const op = "auth_usecase.UpdateUserRole"
	if userID <= 0 || !domain.IsValidRole(string(role)) {
		s.logger.Error().Int64("userID", userID).Msg("failed to update user")
		return fmt.Errorf("%s: %w", op, customErr.ErrInvalidInput)
	}
	s.logger.Info().Int64("userID", userID).Msg("user updated")
	return s.userRepo.UpdateUserRole(ctx, userID, role)
}
