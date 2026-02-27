package router

import (
	"net/http"
	"warehouse-control/internal/domain"
	authH "warehouse-control/internal/http-server/handler/auth"
	historyH "warehouse-control/internal/http-server/handler/history"
	itemsH "warehouse-control/internal/http-server/handler/items"
	"warehouse-control/internal/http-server/middleware"

	"github.com/go-chi/chi/v5"
)

func New(items *itemsH.ItemsHandler, history *historyH.HistoryHandler, auth *authH.AuthHandler, mw *middleware.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RecoveryMiddleware)
	r.Use(middleware.LoggingMiddleware)

	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	r.Post("/auth/login", auth.Login)

	r.Group(func(r chi.Router) {
		r.Use(mw.Middleware)

		r.Route("/items", func(r chi.Router) {
			r.Get("/", items.GetItems)
			r.Get("/{id}", items.GetItemByID)

			r.With(mw.RequireRole(domain.RoleManager, domain.RoleAdmin)).Post("/", items.CreateItem)
			r.With(mw.RequireRole(domain.RoleManager, domain.RoleAdmin)).Put("/{id}", items.UpdateItem)
			r.With(mw.RequireRole(domain.RoleManager, domain.RoleAdmin)).Delete("/{id}", items.DeleteItem)
		})

		r.Route("/history", func(r chi.Router) {
			r.Get("/", history.GetHistory)
			r.Get("/item/{id}", history.GetItemHistory)
			r.Get("/export", history.ExportHistoryCSV)
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	return r
}
