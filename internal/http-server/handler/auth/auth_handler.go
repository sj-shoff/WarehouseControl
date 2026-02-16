package auth_handler

import (
	"encoding/json"
	"net/http"
	"time"
	"warehouse-control/internal/http-server/handler/auth/dto"
	"warehouse-control/internal/http-server/middleware"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wb-go/wbf/zlog"
)

type AuthHandler struct {
	authUsecase authUsecase
	logger      *zlog.Zerolog
	jwtSecret   string
	jwtExpHours int
}

func NewHandler(authUsecase authUsecase, logger *zlog.Zerolog, jwtSecret string, jwtExpHours int) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		logger:      logger,
		jwtSecret:   jwtSecret,
		jwtExpHours: jwtExpHours,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode request")
		http.Error(w, "invalid_input", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		h.logger.Warn().Msg("Missing credentials")
		http.Error(w, "invalid_input", http.StatusBadRequest)
		return
	}
	user, err := h.authUsecase.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		h.logger.Warn().Err(err).Msg("Authentication failed")
		http.Error(w, "invalid_credentials", http.StatusUnauthorized)
		return
	}
	expiresAt := time.Now().Add(time.Duration(h.jwtExpHours) * time.Hour)
	claims := middleware.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to sign token")
		http.Error(w, "internal_error", http.StatusInternalServerError)
		return
	}
	resp := dto.LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		User: dto.UserClaim{
			UserID:   user.ID,
			Username: user.Username,
			Role:     string(user.Role),
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	h.logger.Info().Str("username", user.Username).Msg("User logged in")
}

func (h *AuthHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.authUsecase.GetUsers(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get users")
		http.Error(w, "internal_error", http.StatusInternalServerError)
		return
	}
	resp := dto.UsersResponse{
		Users: make([]*dto.UserResponse, len(users)),
	}
	for i, user := range users {
		resp.Users[i] = &dto.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Role:     string(user.Role),
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
