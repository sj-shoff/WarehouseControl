package router

import (
	"warehouse-control/internal/config"
	"warehouse-control/internal/domain"
	authH "warehouse-control/internal/http-server/handler/auth"
	historyH "warehouse-control/internal/http-server/handler/history"
	itemsH "warehouse-control/internal/http-server/handler/items"

	"warehouse-control/internal/http-server/middleware"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/zlog"
)

func New(items *itemsH.ItemsHandler,
	history *historyH.HistoryHandler,
	auth *authH.AuthHandler,
	mw *middleware.AuthMiddleware,
	cfg *config.Config,
	logger *zlog.Zerolog) *gin.Engine {

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.LoggingMiddleware())
	if cfg.RateLimit.Enabled {
		rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(
			cfg.RateLimit.Rate,
			cfg.RateLimit.Capacity,
			logger,
		)
		r.Use(rateLimiterMiddleware.Middleware())
	}

	r.HandleMethodNotAllowed = true

	r.Static("/static", "./static")
	r.POST("/auth/login", auth.Login)

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

	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	return r
}
