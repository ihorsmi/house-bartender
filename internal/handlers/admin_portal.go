package handlers

import (
	"net/http"
	"strings"

	"house-bartender-go/internal/app"
	"house-bartender-go/internal/db"

	"github.com/go-chi/chi/v5"
)

type AdminUsersPage struct {
	Users []db.User
}

type AdminSettingsPage struct {
	DBPath  string
	DataDir string
	Counts  string
}

func (s *Server) AdminUsersGet(w http.ResponseWriter, r *http.Request) {
	users, _ := s.App.Store().Q.ListUsers()
	s.renderLayout(w, r, "Users", "admin_users.html", AdminUsersPage{Users: users})
}

func (s *Server) AdminUserCreatePost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	email := app.NormalizeEmail(r.FormValue("email"))
	name := strings.TrimSpace(r.FormValue("display_name"))
	role := strings.TrimSpace(r.FormValue("role"))
	pw := r.FormValue("password")

	if email == "" || name == "" || role == "" {
		s.App.AddFlash(w, r, app.FlashError, "Email, name and role are required.")
		s.redirect(w, r, "/admin/users")
		return
	}
	if role != app.RoleUser && role != app.RoleBartender && role != app.RoleAdmin {
		s.App.AddFlash(w, r, app.FlashError, "Invalid role.")
		s.redirect(w, r, "/admin/users")
		return
	}
	hash, err := app.HashPassword(pw)
	if err != nil {
		s.App.AddFlash(w, r, app.FlashError, "Password must be at least 8 characters.")
		s.redirect(w, r, "/admin/users")
		return
	}

	_, err = s.App.Store().Q.CreateUser(db.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
		Role:         role,
		DisplayName:  name,
		IsActive:     true,
		OnDuty:       role == app.RoleBartender,
	})
	if err != nil {
		s.App.AddFlash(w, r, app.FlashError, "Could not create user (email might already exist).")
		s.redirect(w, r, "/admin/users")
		return
	}

	s.App.AddFlash(w, r, app.FlashSuccess, "User created.")
	s.redirect(w, r, "/admin/users")
}

func (s *Server) AdminUserUpdatePost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/admin/users")
		return
	}

	_ = r.ParseForm()
	email := app.NormalizeEmail(r.FormValue("email"))
	name := strings.TrimSpace(r.FormValue("display_name"))
	role := strings.TrimSpace(r.FormValue("role"))
	pw := r.FormValue("password")

	if email == "" || name == "" || role == "" {
		s.App.AddFlash(w, r, app.FlashError, "Email, name and role are required.")
		s.redirect(w, r, "/admin/users")
		return
	}

	target, _ := s.App.Store().Q.GetUserByID(id)
	if target == nil {
		s.redirect(w, r, "/admin/users")
		return
	}

	// Prevent removing the last active admin
	if target.Role == app.RoleAdmin && role != app.RoleAdmin {
		if !s.hasAnotherActiveAdmin(id) {
			s.App.AddFlash(w, r, app.FlashError, "Cannot remove the last active admin.")
			s.redirect(w, r, "/admin/users")
			return
		}
	}

	if err := s.App.Store().Q.UpdateUser(db.UpdateUserParams{
		ID:          id,
		Email:       email,
		Role:        role,
		DisplayName: name,
	}); err != nil {
		s.App.AddFlash(w, r, app.FlashError, "Update failed (email might already exist).")
		s.redirect(w, r, "/admin/users")
		return
	}

	if strings.TrimSpace(pw) != "" {
		hash, err := app.HashPassword(pw)
		if err != nil {
			s.App.AddFlash(w, r, app.FlashError, "Password must be at least 8 characters.")
			s.redirect(w, r, "/admin/users")
			return
		}
		_ = s.App.Store().Q.SetUserPassword(id, hash)
	}

	s.App.AddFlash(w, r, app.FlashSuccess, "User updated.")
	s.redirect(w, r, "/admin/users")
}

func (s *Server) AdminUserTogglePost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/admin/users")
		return
	}

	_ = r.ParseForm()
	active := strings.TrimSpace(r.FormValue("active")) == "1"

	target, _ := s.App.Store().Q.GetUserByID(id)
	if target == nil {
		s.redirect(w, r, "/admin/users")
		return
	}

	if target.Role == app.RoleAdmin && !active {
		if !s.hasAnotherActiveAdmin(id) {
			s.App.AddFlash(w, r, app.FlashError, "Cannot disable the last active admin.")
			s.redirect(w, r, "/admin/users")
			return
		}
	}

	_ = s.App.Store().Q.SetUserActive(id, active)
	s.App.AddFlash(w, r, app.FlashSuccess, "User status updated.")
	s.redirect(w, r, "/admin/users")
}

func (s *Server) AdminUserDutyPost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/admin/users")
		return
	}
	_ = r.ParseForm()
	onDuty := strings.TrimSpace(r.FormValue("on_duty")) == "1"
	_ = s.App.Store().Q.SetUserDuty(id, onDuty)
	s.App.AddFlash(w, r, app.FlashSuccess, "Duty updated.")
	s.redirect(w, r, "/admin/users")
}

func (s *Server) AdminSettingsGet(w http.ResponseWriter, r *http.Request) {
	counts, _ := s.App.Store().Q.DebugCounts()
	cfg := s.App.Config()
	page := AdminSettingsPage{
		DBPath:  cfg.DBPath,
		DataDir: cfg.DataDir,
		Counts:  counts,
	}
	s.renderLayout(w, r, "Settings", "admin_settings.html", page)
}

func (s *Server) AdminSettingsSeedPost(w http.ResponseWriter, r *http.Request) {
	if err := db.SeedCatalog(s.App.Store().DB); err != nil {
		s.App.AddFlash(w, r, app.FlashError, "Seed failed: "+err.Error())
		s.redirect(w, r, "/admin/settings")
		return
	}
	s.App.AddFlash(w, r, app.FlashSuccess, "Catalog seed ran (idempotent).")
	s.redirect(w, r, "/admin/settings")
}

func (s *Server) hasAnotherActiveAdmin(excludeID int64) bool {
	users, _ := s.App.Store().Q.ListUsers()
	n := 0
	for _, u := range users {
		if u.ID == excludeID {
			continue
		}
		if u.Role == app.RoleAdmin && u.IsActive {
			n++
		}
	}
	return n > 0
}
