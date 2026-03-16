package router

import (
	"net/http"
	"os"
	"path/filepath"

	"warehouse-control/internal/config"
	"warehouse-control/internal/domain"
	authH "warehouse-control/internal/http-server/handler/auth"
	historyH "warehouse-control/internal/http-server/handler/history"
	itemsH "warehouse-control/internal/http-server/handler/items"
	"warehouse-control/internal/http-server/middleware"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/zlog"
)

func New(
	items *itemsH.ItemsHandler,
	history *historyH.HistoryHandler,
	auth *authH.AuthHandler,
	mw *middleware.AuthMiddleware,
	cfg *config.Config,
	logger *zlog.Zerolog,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.LoggingMiddleware())

	if cfg.RateLimit.Enabled {
		r.Use(middleware.NewRateLimiterMiddleware(
			cfg.RateLimit.Rate,
			cfg.RateLimit.Capacity,
			logger,
		).Middleware())
	}

	workDir, _ := os.Getwd()
	r.Static("/static", filepath.Join(workDir, "static"))

	r.POST("/auth/login", auth.Login)
	r.POST("/auth/refresh", auth.Refresh)

	protected := r.Group("/")
	protected.Use(mw.Middleware())

	protected.GET("/items", items.GetItems)
	protected.GET("/items/:id", items.GetItemByID)
	protected.POST("/items", mw.RequireRole(domain.RoleManager, domain.RoleAdmin), items.CreateItem)
	protected.PUT("/items/:id", mw.RequireRole(domain.RoleManager, domain.RoleAdmin), items.UpdateItem)
	protected.DELETE("/items/:id", mw.RequireRole(domain.RoleManager, domain.RoleAdmin), items.DeleteItem)
	protected.DELETE("/items/bulk", mw.RequireRole(domain.RoleAdmin), items.BulkDeleteItems)

	protected.GET("/history", history.GetHistory)
	protected.GET("/history/item/:id", history.GetItemHistory)
	protected.GET("/history/export", history.ExportHistoryCSV)

	r.GET("/", serveIndex)
	r.NoRoute(serveIndex)

	return r
}

func serveIndex(c *gin.Context) {
	workDir, _ := os.Getwd()
	indexPath := filepath.Join(workDir, "static", "templates", "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		c.String(http.StatusNotFound, "index.html not found")
		return
	}
	c.File(indexPath)
}
