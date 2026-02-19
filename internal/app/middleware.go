package app

import (
	"context"
	"net/http"
	"strings"

	"house-bartender-go/internal/db"
)

type ctxKey string

const ctxKeyUser ctxKey = "user"

func (a *App) middlewareLoadCurrentUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := a.GetSessionUserID(r)
		if ok {
			u, err := a.store.Q.GetUserByID(userID)
			if err == nil && u != nil && u.IsActive {
				ctx := context.WithValue(r.Context(), ctxKeyUser, u)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// middlewareOnboardingGate enforces:
// - If NO admin exists -> only /onboarding + static/uploads + /health are accessible (everything else redirects to /onboarding)
// - If admin exists -> /onboarding is disabled (redirect to /login)
func (a *App) middlewareOnboardingGate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Always allow health & static assets
		if path == "/health" ||
			strings.HasPrefix(path, "/static/") ||
			strings.HasPrefix(path, "/uploads/") {
			next.ServeHTTP(w, r)
			return
		}

		hasAdmin, err := a.store.Q.HasAnyAdmin()
		if err != nil {
			// Fail safe: if DB check fails, allow request through (better than hard-locking the UI).
			// You can switch this to redirect-to-onboarding if you prefer "fail closed".
			if a.log != nil {
				a.log.Error("onboarding gate: HasAnyAdmin failed", "err", err)
			}
			next.ServeHTTP(w, r)
			return
		}

		if !hasAdmin {
			// Before first admin exists, force onboarding.
			if path == "/onboarding" {
				next.ServeHTTP(w, r)
				return
			}
			http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
			return
		}

		// Admin exists: onboarding should not be reachable anymore.
		if path == "/onboarding" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *App) middlewareNoCacheForHTMX(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") != "" {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) CurrentUser(r *http.Request) *db.User {
	u, _ := r.Context().Value(ctxKeyUser).(*db.User)
	return u
}

// Exported wrappers so router wiring can live outside the app package (no handlers import cycle).
func (a *App) MiddlewareNoCacheForHTMX(next http.Handler) http.Handler {
	return a.middlewareNoCacheForHTMX(next)
}

func (a *App) MiddlewareLoadCurrentUser(next http.Handler) http.Handler {
	return a.middlewareLoadCurrentUser(next)
}

func (a *App) MiddlewareOnboardingGate(next http.Handler) http.Handler {
	return a.middlewareOnboardingGate(next)
}
