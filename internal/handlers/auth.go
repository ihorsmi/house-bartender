package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"house-bartender-go/internal/app"
	"house-bartender-go/internal/db"
)

type Server struct {
	App *app.App
}

type ViewData struct {
	Title        string
	Path         string
	User         *db.User
	Flashes      []app.Flash
	PageTemplate string
	Page         any
	Now          time.Time
}

func (s *Server) renderLayout(w http.ResponseWriter, r *http.Request, title, pageTemplate string, page any) {
	data := ViewData{
		Title:        title,
		Path:         r.URL.Path,
		User:         s.App.CurrentUser(r),
		Flashes:      s.App.PopFlashes(w, r),
		PageTemplate: pageTemplate,
		Page:         page,
		Now:          time.Now(),
	}
	_ = s.App.Templates().ExecuteTemplate(w, "layout.html", data)
}

func (s *Server) renderPartial(w http.ResponseWriter, r *http.Request, templateName string, page any, pathOverride string) {
	path := r.URL.Path
	if pathOverride != "" {
		path = pathOverride
	}
	data := ViewData{
		Title:   "",
		Path:    path,
		User:    s.App.CurrentUser(r),
		Flashes: nil,
		Page:    page,
		Now:     time.Now(),
	}
	_ = s.App.Templates().ExecuteTemplate(w, templateName, data)
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request, to string) {
	// Optional HTMX-style redirect header
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", to)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, to, http.StatusSeeOther)
}

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	if err := s.App.Store().Ping(); err != nil {
		http.Error(w, "db not ok", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok"))
}

/* ---------------- Login / Logout ---------------- */

func (s *Server) LoginGet(w http.ResponseWriter, r *http.Request) {
	if s.App.CurrentUser(r) != nil {
		s.redirect(w, r, "/")
		return
	}
	s.renderLayout(w, r, "Login", "login.html", map[string]any{})
}

func (s *Server) LoginPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	email := app.NormalizeEmail(r.FormValue("email"))
	pw := r.FormValue("password")

	u, err := s.App.Store().Q.GetUserByEmail(email)
	if err != nil || u == nil || !u.IsActive || !app.CheckPassword(u.PasswordHash, pw) {
		s.App.AddFlash(w, r, app.FlashError, "Invalid credentials.")
		s.redirect(w, r, "/login")
		return
	}

	_ = s.App.SetSessionUser(w, r, u.ID)
	s.App.AddFlash(w, r, app.FlashSuccess, "Welcome back, "+u.DisplayName+"!")
	s.redirect(w, r, "/")
}

func (s *Server) LogoutPost(w http.ResponseWriter, r *http.Request) {
	_ = s.App.ClearSession(w, r)
	s.App.AddFlash(w, r, app.FlashInfo, "Logged out.")
	s.redirect(w, r, "/login")
}

/* ---------------- Onboarding ---------------- */

func (s *Server) OnboardingGet(w http.ResponseWriter, r *http.Request) {
	// DB is the source of truth: onboarding only when no admin exists.
	hasAdmin, err := s.App.Store().Q.HasAnyAdmin()
	if err == nil && hasAdmin {
		s.redirect(w, r, "/login")
		return
	}
	s.renderLayout(w, r, "Onboarding", "onboarding.html", map[string]any{})
}

func (s *Server) OnboardingPost(w http.ResponseWriter, r *http.Request) {
	// DB is the source of truth: allow creating first admin only if none exists.
	hasAdmin, err := s.App.Store().Q.HasAnyAdmin()
	if err == nil && hasAdmin {
		s.App.AddFlash(w, r, app.FlashInfo, "Admin already exists. Please log in.")
		s.redirect(w, r, "/login")
		return
	}

	_ = r.ParseForm()
	email := app.NormalizeEmail(r.FormValue("email"))
	name := strings.TrimSpace(r.FormValue("display_name"))
	pw := r.FormValue("password")

	if email == "" || name == "" {
		s.App.AddFlash(w, r, app.FlashError, "Email and display name are required.")
		s.redirect(w, r, "/onboarding")
		return
	}

	hash, err := app.HashPassword(pw)
	if err != nil {
		s.App.AddFlash(w, r, app.FlashError, "Password must be at least 8 characters.")
		s.redirect(w, r, "/onboarding")
		return
	}

	id, err := s.App.Store().Q.CreateUser(db.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
		Role:         app.RoleAdmin,
		DisplayName:  name,
		IsActive:     true,
		OnDuty:       false,
	})
	if err != nil {
		// If another request created the admin first, guide user to login.
		// (we can't reliably detect the exact constraint error across sqlite drivers here)
		s.App.AddFlash(w, r, app.FlashError, "Could not create admin (it may already exist). Try logging in.")
		s.redirect(w, r, "/login")
		return
	}

	_ = s.App.SetSessionUser(w, r, id)
	s.App.AddFlash(w, r, app.FlashSuccess, "Admin created. You're in.")
	s.redirect(w, r, "/admin/users")
}

func parseIDParam(r *http.Request, key string) (int64, bool) {
	v := strings.TrimSpace(key)
	if v == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(v, 10, 64)
	return id, err == nil && id > 0
}
