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
	userRepo  userRepository
	logger    *zlog.Zerolog
	jwtSecret string
	tokenTTL  time.Duration
}

func NewAuthUsecase(
	userRepo userRepository,
	logger *zlog.Zerolog,
	jwtSecret string,
	tokenTTL time.Duration,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:  userRepo,
		logger:    logger,
		jwtSecret: jwtSecret,
		tokenTTL:  tokenTTL,
	}
}

func (s *AuthUsecase) Login(ctx context.Context, username, password string) (*domain.UserClaim, string, time.Time, error) {
	const op = "auth_usecase.Login"

	if username == "" || password == "" {
		return nil, "", time.Time{}, customErr.ErrInvalidInput
	}

	s.logger.Info().Str("username", username).Msg("Login attempt")

	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		s.logger.Warn().Err(err).Str("username", username).Msg("User not found")
		if errors.Is(err, customErr.ErrUserNotFound) {
			return nil, "", time.Time{}, customErr.ErrInvalidCredentials
		}
		return nil, "", time.Time{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn().Str("username", username).Msg("Invalid password")
		return nil, "", time.Time{}, customErr.ErrInvalidCredentials
	}

	expiresAt := time.Now().Add(s.tokenTTL)
	token, err := jwt.NewToken(user, s.jwtSecret, s.tokenTTL)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate token")
		return nil, "", time.Time{}, fmt.Errorf("%s: %w", op, customErr.ErrInternal)
	}

	claim := &domain.UserClaim{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}

	s.logger.Info().Str("username", username).Str("role", string(user.Role)).Msg("Login successful")
	return claim, token, expiresAt, nil
}

func (s *AuthUsecase) ValidateToken(ctx context.Context, token string) (*domain.UserClaim, error) {
	claim, err := jwt.ValidateToken(token, s.jwtSecret)
	if err != nil {
		return nil, customErr.ErrTokenInvalid
	}
	return claim, nil
}

func (s *AuthUsecase) CreateUser(ctx context.Context, req *domain.CreateUserRequest) (int64, error) {
	const op = "auth_usecase.CreateUser"

	if req.Username == "" || req.Password == "" || !domain.IsValidRole(string(req.Role)) {
		return 0, customErr.ErrInvalidInput
	}

	s.logger.Info().Str("username", req.Username).Str("role", string(req.Role)).Msg("Create user attempt")

	passHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to hash password")
		return 0, fmt.Errorf("%s: %w", op, customErr.ErrInternal)
	}

	user := &domain.User{
		Username:     req.Username,
		PasswordHash: string(passHash),
		Role:         req.Role,
	}

	id, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		s.logger.Error().Err(err).Str("username", req.Username).Msg("Failed to create user")
		if errors.Is(err, customErr.ErrUserExists) {
			return 0, customErr.ErrUserExists
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	s.logger.Info().Int64("user_id", id).Str("username", req.Username).Msg("User created")
	return id, nil
}

func (s *AuthUsecase) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	if id <= 0 {
		return nil, customErr.ErrInvalidInput
	}
	return s.userRepo.GetUserByID(ctx, id)
}

func (s *AuthUsecase) GetUsers(ctx context.Context) ([]*domain.User, error) {
	s.logger.Info().Msg("Getting users")
	users, err := s.userRepo.GetUsers(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get users")
		return nil, fmt.Errorf("auth_usecase.GetUsers: %w", err)
	}
	s.logger.Info().Int("count", len(users)).Msg("Users retrieved")
	return users, nil
}

func (s *AuthUsecase) UpdateUser(ctx context.Context, req *domain.UpdateUserRequest) error {
	const op = "auth_usecase.UpdateUser"

	if req.ID <= 0 || req.Username == "" || !domain.IsValidRole(string(req.Role)) {
		return customErr.ErrInvalidInput
	}

	s.logger.Info().Int64("user_id", req.ID).Str("username", req.Username).Msg("Update user attempt")

	err := s.userRepo.UpdateUser(ctx, req.ID, req.Username, req.Role)
	if err != nil {
		s.logger.Error().Err(err).Int64("user_id", req.ID).Msg("Failed to update user")
		if errors.Is(err, customErr.ErrUserNotFound) {
			return customErr.ErrUserNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	s.logger.Info().Int64("user_id", req.ID).Msg("User updated")
	return nil
}

func (s *AuthUsecase) DeleteUser(ctx context.Context, id int64) error {
	if id <= 0 {
		return customErr.ErrInvalidInput
	}
	s.logger.Info().Int64("user_id", id).Msg("Deleting user")
	return s.userRepo.DeleteUser(ctx, id)
}

func (s *AuthUsecase) UpdateUserRole(ctx context.Context, id int64, role domain.UserRole) error {
	if id <= 0 || !domain.IsValidRole(string(role)) {
		return customErr.ErrInvalidInput
	}
	s.logger.Info().Int64("user_id", id).Str("role", string(role)).Msg("Updating user role")
	return s.userRepo.UpdateUserRole(ctx, id, role)
}
