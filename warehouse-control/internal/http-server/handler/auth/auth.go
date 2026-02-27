package auth

import (
	"encoding/json"
	"net/http"
	"time"
	"warehouse-control/internal/grpc/sso"
	"warehouse-control/internal/http-server/handler/auth/dto"

	"github.com/wb-go/wbf/zlog"
)

type AuthHandler struct {
	ssoClient *sso.Client
	logger    *zlog.Zerolog
}

func NewHandler(ssoClient *sso.Client, logger *zlog.Zerolog) *AuthHandler {
	return &AuthHandler{ssoClient: ssoClient, logger: logger}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid_input", http.StatusBadRequest)
		return
	}

	token, err := h.ssoClient.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, "invalid_credentials", http.StatusUnauthorized)
		return
	}

	resp := dto.LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
