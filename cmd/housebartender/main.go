package main

import (
	"context"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"house-bartender-go/internal/app"
	"house-bartender-go/internal/handlers"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg := app.Config{
		Addr:    getenv("ADDR", ":8080"),
		BaseURL: getenv("BASE_URL", "http://localhost:8080"),

		DataDir:   getenv("DATA_DIR", "/data"),
		DBPath:    getenv("DB_PATH", "/data/housebartender.db"),
		UploadDir: getenv("UPLOAD_DIR", "/data/uploads"),

		BootstrapAdminEmail:    os.Getenv("BOOTSTRAP_ADMIN_EMAIL"),
		BootstrapAdminPassword: os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"),
		BootstrapAdminName:     os.Getenv("BOOTSTRAP_ADMIN_NAME"),
	}

	// Optional: allow keys as hex in env
	if hk := strings.TrimSpace(os.Getenv("SESSION_HASH_KEY_HEX")); hk != "" {
		if b, err := hex.DecodeString(hk); err == nil {
			cfg.SessionHashKey = b
		}
	}
	if bk := strings.TrimSpace(os.Getenv("SESSION_BLOCK_KEY_HEX")); bk != "" {
		if b, err := hex.DecodeString(bk); err == nil {
			cfg.SessionBlockKey = b
		}
	}

	a, err := app.New(cfg, logger)
	if err != nil {
		logger.Error("app init failed", "err", err)
		os.Exit(1)
	}
	defer a.Close()

	// Build router here (avoids app<->handlers import cycle)
	r := chi.NewRouter()
	r.Use(chimw.RealIP)
	r.Use(chimw.RequestID)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))

	r.Use(a.MiddlewareNoCacheForHTMX)
	r.Use(a.MiddlewareLoadCurrentUser)
	r.Use(a.MiddlewareOnboardingGate)

	h := &handlers.Server{App: a}

	// Public
	r.Get("/health", h.Health)
	r.Get("/login", h.LoginGet)
	r.Post("/login", h.LoginPost)
	r.Post("/logout", h.LogoutPost)

	r.Get("/onboarding", h.OnboardingGet)
	r.Post("/onboarding", h.OnboardingPost)

	// Static + uploads
	fileServer(r, "/static", http.Dir("static"))
	fileServer(r, "/uploads", http.Dir(a.Config().UploadDir))

	// Authenticated common
	r.Group(func(ar chi.Router) {
		ar.Use(a.RequireAuth)

		ar.Get("/", h.UserHomeGet)
		ar.Get("/cocktails/{id}", h.CocktailDetailGet)
		ar.Post("/orders", h.OrderCreatePost)
		ar.Get("/orders", h.UserOrdersGet)

		ar.Get("/partials/user/cocktails", h.UserCocktailsPartialGet)
		ar.Get("/partials/user/orders", h.UserOrdersPartialGet)

		ar.Get("/sse", h.SSEGet)
	})

	// Bartender (admin allowed)
	r.Route("/bartender", func(br chi.Router) {
		br.Use(a.RequireAnyRole(app.RoleBartender, app.RoleAdmin))

		br.Get("/", h.BartenderDashboardGet)
		br.Post("/duty", h.BartenderDutyPost)

		br.Get("/products", h.BartenderProductsGet)
		br.Post("/products", h.ProductCreatePost)
		br.Get("/products/{id}/edit", h.ProductEditGet)
		br.Post("/products/{id}/edit", h.ProductEditPost)
		br.Post("/products/{id}/toggle", h.ProductTogglePost)
		br.Post("/products/{id}/stock", h.ProductStockPost)
		br.Post("/products/{id}/delete", h.ProductDeletePost)

		br.Get("/cocktails", h.BartenderCocktailsGet)
		br.Get("/cocktails/new", h.CocktailNewGet)
		br.Post("/cocktails/new", h.CocktailNewPost)
		br.Get("/cocktails/{id}/edit", h.CocktailEditGet)
		br.Post("/cocktails/{id}/edit", h.CocktailEditPost)
		br.Post("/cocktails/{id}/toggle", h.CocktailTogglePost)
		br.Post("/cocktails/{id}/delete", h.CocktailDeletePost)

		br.Get("/orders", h.BartenderOrdersGet)
		br.Post("/orders/{id}/accept", h.OrderAcceptPost)
		br.Post("/orders/{id}/assign", h.OrderAssignPost)
		br.Post("/orders/{id}/status", h.OrderStatusPost)
		br.Post("/orders/{id}/cancel", h.OrderCancelPost)

		br.Get("/partials/products", h.BartenderProductsPartialGet)
		br.Get("/partials/cocktails", h.BartenderCocktailsPartialGet)
		br.Get("/partials/orders", h.BartenderOrdersPartialGet)
	})

	// Admin
	r.Route("/admin", func(ad chi.Router) {
		ad.Use(a.RequireRole(app.RoleAdmin))

		ad.Get("/users", h.AdminUsersGet)
		ad.Post("/users", h.AdminUserCreatePost)
		ad.Post("/users/{id}", h.AdminUserUpdatePost)
		ad.Post("/users/{id}/toggle", h.AdminUserTogglePost)
		ad.Post("/users/{id}/duty", h.AdminUserDutyPost)

		ad.Get("/settings", h.AdminSettingsGet)
		ad.Post("/settings/seed", h.AdminSettingsSeedPost)
	})

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  90 * time.Second,
	}

	go func() {
		logger.Info("listening", "addr", cfg.Addr, "base_url", cfg.BaseURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	logger.Info("shutdown complete")
}

func getenv(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("fileServer does not permit URL params")
	}
	fs := http.StripPrefix(path, http.FileServer(root))
	if path != "/" && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	r.Get(path+"/*", func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}
