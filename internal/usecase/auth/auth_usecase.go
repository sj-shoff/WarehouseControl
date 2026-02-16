package auth_usecase

import (
	"context"
	"errors"
	"fmt"
	"warehouse-control/internal/domain"
	customErr "warehouse-control/internal/domain/errors"

	"github.com/wb-go/wbf/zlog"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	repo   usersRepository
	logger *zlog.Zerolog
}

func NewService(repo usersRepository, logger *zlog.Zerolog) *AuthUsecase {
	return &AuthUsecase{
		repo:   repo,
		logger: logger,
	}
}

func (s *AuthUsecase) Authenticate(ctx context.Context, username, password string) (*domain.User, error) {
	if username == "" || password == "" {
		return nil, customErr.ErrInvalidCredentials
	}
	s.logger.Info().Str("username", username).Msg("Authenticating user")
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		s.logger.Warn().Err(err).Str("username", username).Msg("User not found")
		if errors.Is(err, customErr.ErrUserNotFound) {
			return nil, customErr.ErrInvalidCredentials
		}
		if errors.Is(err, customErr.ErrDatabase) {
			return nil, customErr.ErrDatabase
		}
		return nil, fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		s.logger.Warn().Str("username", username).Msg("Invalid password")
		return nil, customErr.ErrInvalidCredentials
	}
	s.logger.Info().Str("username", username).Str("role", string(user.Role)).Msg("User authenticated")
	return user, nil
}

func (s *AuthUsecase) GetUsers(ctx context.Context) ([]*domain.User, error) {
	s.logger.Info().Msg("Getting users")
	users, err := s.repo.GetUsers(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get users")
		if errors.Is(err, customErr.ErrDatabase) {
			return nil, customErr.ErrDatabase
		}
		return nil, fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Int("count", len(users)).Msg("Users retrieved")
	return users, nil
}
