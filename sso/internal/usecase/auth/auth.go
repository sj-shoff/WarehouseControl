package auth_usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"sso/internal/domain"
	customErr "sso/internal/domain/errors"
	"sso/internal/lib/jwt"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/zlog"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepo   UserRepository
	tokenRepo  TokenRepository
	appRepo    AppRepository
	logger     *zlog.Zerolog
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthUsecase(
	userRepo UserRepository,
	tokenRepo TokenRepository,
	appRepo AppRepository,
	logger *zlog.Zerolog,
	jwtSecret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		appRepo:    appRepo,
		logger:     logger,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *AuthUsecase) Login(ctx context.Context, username, password string, appID int, appSecret string) (*domain.UserClaim, string, string, time.Time, error) {
	if username == "" || password == "" || appID <= 0 || appSecret == "" {
		return nil, "", "", time.Time{}, customErr.ErrInvalidInput
	}

	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		s.logger.Warn().Int("app_id", appID).Msg("access attempt to non-existent app")
		return nil, "", "", time.Time{}, customErr.ErrInvalidInput
	}

	if app.Secret != appSecret {
		s.logger.Warn().Int("app_id", appID).Msg("invalid app secret provided during login")
		return nil, "", "", time.Time{}, customErr.ErrInvalidInput
	}

	user, err := s.userRepo.GetUserByUsernameAndApp(ctx, username, appID)
	if err != nil {
		if errors.Is(err, customErr.ErrUserNotFound) {
			s.logger.Warn().Str("username", username).Msg("invalid credentials (user not found in app)")
			return nil, "", "", time.Time{}, customErr.ErrInvalidCredentials
		}
		return nil, "", "", time.Time{}, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn().Str("username", username).Msg("invalid password attempt")
		return nil, "", "", time.Time{}, customErr.ErrInvalidCredentials
	}

	accessToken, err := jwt.NewToken(user, s.jwtSecret, s.accessTTL)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to generate access token")
		return nil, "", "", time.Time{}, customErr.ErrInternal
	}

	refreshToken := uuid.New().String()
	if err := s.tokenRepo.SaveRefreshToken(ctx, user.ID, s.hashToken(refreshToken), appID, time.Now().Add(s.refreshTTL)); err != nil {
		s.logger.Error().Err(err).Msg("failed to save refresh token")
		return nil, "", "", time.Time{}, fmt.Errorf("save refresh: %w", err)
	}

	s.logger.Info().Str("username", username).Int64("uid", user.ID).Msg("user logged in")

	return &domain.UserClaim{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}, accessToken, refreshToken, time.Now().Add(s.accessTTL), nil
}

func (s *AuthUsecase) Refresh(ctx context.Context, oldRefreshToken string, appID int) (string, string, error) {
	tokenHash := s.hashToken(oldRefreshToken)

	userID, savedAppID, expiresAt, err := s.tokenRepo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, customErr.ErrInvalidCredentials) {
			return "", "", customErr.ErrInvalidCredentials
		}
		return "", "", fmt.Errorf("get refresh token: %w", err)
	}

	_ = s.tokenRepo.DeleteRefreshToken(ctx, tokenHash)

	if time.Now().After(expiresAt) || savedAppID != appID {
		s.logger.Warn().Int64("uid", userID).Msg("invalid or expired refresh token attempt")
		return "", "", customErr.ErrInvalidCredentials
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("get user for refresh: %w", err)
	}

	newAccessToken, err := jwt.NewToken(user, s.jwtSecret, s.accessTTL)
	if err != nil {
		return "", "", customErr.ErrInternal
	}

	newRefreshToken := uuid.New().String()
	if err := s.tokenRepo.SaveRefreshToken(ctx, user.ID, s.hashToken(newRefreshToken), appID, time.Now().Add(s.refreshTTL)); err != nil {
		return "", "", fmt.Errorf("save new refresh: %w", err)
	}

	s.logger.Info().Int64("uid", userID).Msg("tokens refreshed")

	return newAccessToken, newRefreshToken, nil
}

func (s *AuthUsecase) RegisterNewUser(ctx context.Context, username, password string, role domain.UserRole, appID int) (int64, error) {
	if username == "" || password == "" || appID <= 0 {
		return 0, customErr.ErrInvalidInput
	}

	if _, err := s.appRepo.GetByID(ctx, appID); err != nil {
		return 0, customErr.ErrInvalidInput
	}

	if !domain.IsValidRole(string(role)) {
		role = domain.RoleViewer
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error().Err(err).Msg("password hashing failed")
		return 0, customErr.ErrInternal
	}

	id, err := s.userRepo.CreateUser(ctx, &domain.User{
		Username:     username,
		PasswordHash: string(passHash),
		Role:         role,
		AppID:        appID,
	})

	if err != nil {
		if errors.Is(err, customErr.ErrUserExists) {
			return 0, customErr.ErrUserExists
		}
		return 0, fmt.Errorf("create user: %w", err)
	}

	s.logger.Info().Int64("uid", id).Str("username", username).Msg("new user registered")
	return id, nil
}

func (s *AuthUsecase) GetUsers(ctx context.Context) ([]*domain.User, error) {
	users, err := s.userRepo.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch users list: %w", err)
	}
	return users, nil
}

func (s *AuthUsecase) UpdateUserRole(ctx context.Context, userID int64, role domain.UserRole) error {
	if userID <= 0 || !domain.IsValidRole(string(role)) {
		return customErr.ErrInvalidInput
	}

	if err := s.userRepo.UpdateUserRole(ctx, userID, role); err != nil {
		return fmt.Errorf("update role for user %d: %w", userID, err)
	}

	s.logger.Info().Int64("uid", userID).Str("new_role", string(role)).Msg("user role updated")
	return nil
}

func (s *AuthUsecase) hashToken(t string) string {
	h := sha256.Sum256([]byte(t))
	return hex.EncodeToString(h[:])
}

func (u *AuthUsecase) InitialBootstrap(
	ctx context.Context,
	appName,
	appSecret,
	adminUser,
	adminPass string,
) (int64, int32, error) {
	appID, err := u.appRepo.UpsertApp(ctx, appName, appSecret)
	if err != nil {
		u.logger.Error().Err(err).Str("app", appName).Msg("failed to upsert app")
		return 0, 0, fmt.Errorf("app bootstrap failed: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to hash password: %w", err)
	}

	uid, err := u.userRepo.CreateAdminIfNotExist(ctx, adminUser, string(hash), int(appID))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to ensure admin: %w", err)
	}

	u.logger.Info().
		Str("app", appName).
		Int32("app_id", appID).
		Msg("application bootstrap completed")

	return uid, appID, nil
}
