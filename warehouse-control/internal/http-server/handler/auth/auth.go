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
		c.JSON(http.StatusBadRequest, gin.H{"error": customErr.ErrInvalidInput})
		return
	}
	token, err := h.ssoClient.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": customErr.ErrInvalidCredentials})
		return
	}
	resp := dto.LoginResponse{
		Token:     token,
		ExpiresAt: h.config.JWT.ExpHours,
	}
	c.JSON(http.StatusOK, resp)
}
