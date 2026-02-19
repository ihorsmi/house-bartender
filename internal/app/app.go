package app

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"house-bartender-go/internal/db"
)

type Config struct {
	Addr    string
	BaseURL string

	DataDir   string
	DBPath    string
	UploadDir string

	SessionHashKey  []byte
	SessionBlockKey []byte

	BootstrapAdminEmail    string
	BootstrapAdminPassword string
	BootstrapAdminName     string
}

type App struct {
	cfg       Config
	store     *db.Store
	log       *slog.Logger
	templates *template.Template
	sseHub    *SSEHub

	// Kept for backward compatibility; onboarding gating is enforced via DB in middleware.
	needsOnboarding bool
}

func New(cfg Config, logger *slog.Logger) (*App, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080"
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "/data"
	}
	if cfg.DBPath == "" {
		cfg.DBPath = filepath.Join(cfg.DataDir, "housebartender.db")
	}
	if cfg.UploadDir == "" {
		cfg.UploadDir = filepath.Join(cfg.DataDir, "uploads")
	}

	// NOTE: /data is a Docker volume; ensure paths exist.
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir data dir: %w", err)
	}
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir upload dir: %w", err)
	}

	// Load keys from env (hex) if present
	if len(cfg.SessionHashKey) == 0 {
		if hk := strings.TrimSpace(os.Getenv("SESSION_HASH_KEY_HEX")); hk != "" {
			b, err := hex.DecodeString(hk)
			if err != nil {
				return nil, fmt.Errorf("SESSION_HASH_KEY_HEX invalid hex: %w", err)
			}
			cfg.SessionHashKey = b
		}
	}
	if len(cfg.SessionBlockKey) == 0 {
		if bk := strings.TrimSpace(os.Getenv("SESSION_BLOCK_KEY_HEX")); bk != "" {
			b, err := hex.DecodeString(bk)
			if err != nil {
				return nil, fmt.Errorf("SESSION_BLOCK_KEY_HEX invalid hex: %w", err)
			}
			cfg.SessionBlockKey = b
		}
	}

	// Ensure we have a signing key (stable recommended via env, but generate if missing)
	if len(cfg.SessionHashKey) < 32 {
		cfg.SessionHashKey = make([]byte, 32)
		_, _ = rand.Read(cfg.SessionHashKey)
		logger.Warn("SESSION_HASH_KEY_HEX not set (or too short) — generating ephemeral session key; sessions will reset on restart")
	}

	store, err := db.Open(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Migrate(store.DB); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	a := &App{
		cfg:    cfg,
		store:  store,
		log:    logger,
		sseHub: NewSSEHub(logger),
	}

	// Templates
	funcs := template.FuncMap{
		"fmtTime": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Local().Format("2006-01-02 15:04")
		},
		"since": func(t time.Time, now time.Time) string {
			if t.IsZero() {
				return ""
			}
			d := now.Sub(t)
			if d < 0 {
				d = -d
			}
			if d < time.Minute {
				return "just now"
			}
			if d < time.Hour {
				return fmt.Sprintf("%dm", int(d.Minutes()))
			}
			if d < 24*time.Hour {
				return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
			}
			return fmt.Sprintf("%dd %dh", int(d.Hours())/24, int(d.Hours())%24)
		},
		"splitCSV": func(s string) []string {
			var out []string
			for _, p := range strings.Split(s, ",") {
				p = strings.TrimSpace(p)
				if p != "" {
					out = append(out, p)
				}
			}
			return out
		},
		"hasPrefix": strings.HasPrefix,
		"fmtQty": func(q *float64) string {
			if q == nil {
				return ""
			}
			// render without trailing zeros (150 -> "150", 0.5 -> "0.5")
			return strconv.FormatFloat(*q, 'f', -1, 64)
		},
	}

	tpl := template.New("all").Funcs(funcs)
	tpl, err = tpl.ParseGlob("views/templates/*.html")
	if err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	tpl, err = tpl.ParseGlob("views/partials/*.html")
	if err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("parse partials: %w", err)
	}
	a.templates = tpl

	// Bootstrap admin if none exists (only once).
	hasAdmin, err := store.Q.HasAnyAdmin()
	if err != nil {
		_ = store.Close()
		return nil, err
	}
	if !hasAdmin {
		email := strings.TrimSpace(cfg.BootstrapAdminEmail)
		pass := strings.TrimSpace(cfg.BootstrapAdminPassword)
		name := strings.TrimSpace(cfg.BootstrapAdminName)

		if email != "" && pass != "" && name != "" {
			hash, err := HashPassword(pass)
			if err != nil {
				_ = store.Close()
				return nil, err
			}
			_, err = store.Q.CreateUser(db.CreateUserParams{
				Email:        NormalizeEmail(email),
				PasswordHash: hash,
				Role:         RoleAdmin,
				DisplayName:  name,
				IsActive:     true,
				OnDuty:       false,
			})
			if err != nil {
				_ = store.Close()
				return nil, fmt.Errorf("bootstrap admin: %w", err)
			}
			a.log.Info("bootstrapped admin user", "email", NormalizeEmail(email))
			a.needsOnboarding = false
		} else {
			a.needsOnboarding = true
		}
	} else {
		a.needsOnboarding = false
	}

	// Seed catalog ONLY if empty (never touches users).
	empty, err := isCatalogEmpty(store.DB)
	if err != nil {
		a.log.Warn("catalog empty check failed", "err", err)
	} else if empty {
		if err := db.SeedCatalog(store.DB); err != nil {
			a.log.Warn("catalog seed failed", "err", err)
		} else {
			a.log.Info("catalog seeded")
		}
	}

	return a, nil
}

func isCatalogEmpty(dbh *sql.DB) (bool, error) {
	var pc int
	if err := dbh.QueryRow(`SELECT COUNT(1) FROM products;`).Scan(&pc); err != nil {
		return false, err
	}
	var cc int
	if err := dbh.QueryRow(`SELECT COUNT(1) FROM cocktails;`).Scan(&cc); err != nil {
		return false, err
	}
	return pc == 0 && cc == 0, nil
}

func (a *App) Close() error {
	if a == nil {
		return nil
	}
	if a.store != nil {
		return a.store.Close()
	}
	return nil
}

func (a *App) Store() *db.Store              { return a.store }
func (a *App) Templates() *template.Template { return a.templates }
func (a *App) SSE() *SSEHub                  { return a.sseHub }
func (a *App) Config() Config                { return a.cfg }
func (a *App) NeedsOnboarding() bool         { return a.needsOnboarding }
func (a *App) ClearOnboarding()              { a.needsOnboarding = false }
