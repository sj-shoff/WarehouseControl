package auth

import (
	"net/http"

	"warehouse-control/internal/config"
	customErr "warehouse-control/internal/domain/errors"
	"warehouse-control/internal/grpc/sso"
	"warehouse-control/internal/http-server/handler/auth/dto"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/zlog"
)

type AuthHandler struct {
	ssoClient *sso.Client
	config    *config.Config
	logger    *zlog.Zerolog
}

func NewHandler(ssoClient *sso.Client, config *config.Config, logger *zlog.Zerolog) *AuthHandler {
	return &AuthHandler{ssoClient: ssoClient, config: config, logger: logger}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": customErr.ErrInvalidInput.Error()})
		return
	}

	accessToken, refreshToken, expiresAt, err := h.ssoClient.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.logger.Warn().Err(err).Str("username", req.Username).Msg("login failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": customErr.ErrInvalidCredentials.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": customErr.ErrInvalidInput.Error()})
		return
	}

	newAccess, newRefresh, err := h.ssoClient.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.logger.Warn().Err(err).Msg("refresh failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  newAccess,
		"refresh_token": newRefresh,
	})
}
