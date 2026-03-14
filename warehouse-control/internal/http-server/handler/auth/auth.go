package auth

import (
	"net/http"
	"time"

	"warehouse-control/internal/grpc/sso"
	"warehouse-control/internal/http-server/handler/auth/dto"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/zlog"
)

type AuthHandler struct {
	ssoClient *sso.Client
	logger    *zlog.Zerolog
}

func NewHandler(ssoClient *sso.Client, logger *zlog.Zerolog) *AuthHandler {
	return &AuthHandler{ssoClient: ssoClient, logger: logger}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input"})
		return
	}
	token, err := h.ssoClient.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
		return
	}
	resp := dto.LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	c.JSON(http.StatusOK, resp)
}
