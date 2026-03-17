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
	staticDir := filepath.Join(workDir, "static")
	r.Static("/static", staticDir)

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", auth.Login)
		authGroup.POST("/refresh", auth.Refresh)
	}

	protected := r.Group("/")
	protected.Use(mw.Middleware())
	{
		protected.POST("/auth/register", mw.RequireRole(domain.RoleAdmin), auth.AdminRegisterNewUser)

		itemsGroup := protected.Group("/items")
		{
			itemsGroup.GET("", items.GetItems)
			itemsGroup.GET("/:id", items.GetItemByID)

			writeItems := itemsGroup.Group("")
			writeItems.Use(mw.RequireRole(domain.RoleManager, domain.RoleAdmin))
			{
				writeItems.POST("", items.CreateItem)
				writeItems.PUT("/:id", items.UpdateItem)
				writeItems.DELETE("/:id", items.DeleteItem)
			}

			itemsGroup.DELETE("/bulk", mw.RequireRole(domain.RoleAdmin), items.BulkDeleteItems)
		}

		protected.GET("/history", history.GetHistory)
		protected.GET("/history/item/:id", history.GetItemHistory)
		protected.GET("/history/export", history.ExportHistoryCSV)
		protected.GET("/history/diff/:id", history.GetDiff)
	}

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
