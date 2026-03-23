package app

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"house-bartender-go/internal/catalog"
	"house-bartender-go/internal/db"
	"house-bartender-go/internal/services/push"
)

type Config struct {
	Addr    string
	BaseURL string

	DataDir   string
	DBPath    string
	UploadDir string

	SessionHashKey  []byte
	SessionBlockKey []byte

	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string

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
	push      *push.Service

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
		logger.Warn("SESSION_HASH_KEY_HEX not set (or too short) - generating ephemeral session key; sessions will reset on restart")
	}

	store, err := db.Open(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Migrate(store.DB); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	pushService, err := push.New(store.Q, logger, push.Config{
		PublicKey:  strings.TrimSpace(cfg.VAPIDPublicKey),
		PrivateKey: strings.TrimSpace(cfg.VAPIDPrivateKey),
		Subject:    strings.TrimSpace(cfg.VAPIDSubject),
	})
	if err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("init push service: %w", err)
	}

	a := &App{
		cfg:    cfg,
		store:  store,
		log:    logger,
		sseHub: NewSSEHub(logger),
		push:   pushService,
	}

	// Templates
	humanizeEnum := func(s string) string {
		s = strings.TrimSpace(strings.ReplaceAll(strings.ToLower(s), "_", " "))
		if s == "" {
			return ""
		}
		return strings.ToUpper(s[:1]) + s[1:]
	}
	cocktailPlaceholders := []string{
		"https://lh3.googleusercontent.com/aida-public/AB6AXuAsy80ghZEu-9aP4aXCjTfW0ap5I97HoZccF7aJRizrlUaTyc1NwhP4x0sRTGu9Cgk-FAbUdtpTt-wMwcquIM9czHZTfqJrfEsrwUNPo2SSJzySQvM7BllZmVRQHHftVuk7PsZwGFWcCMLNyjHBCFYKgSqzB9F8L8MHOYypATR1KGYJNdH7NeE2AKvqrtsB0Lj_Ic4_Fzh_t3kyd9tUvbsKULlWt-QLfZbFT5_DbFQo90uZUAthxeTQ-r3I1SR9lbIrfveDUVCung",
		"https://lh3.googleusercontent.com/aida-public/AB6AXuAK7fiiCou66uYd8pXsa8pGu94tWmueX0U5FkOSmxgaG0lmlGOsJKNFYwsQGIowGBU9Ce08f4ZQP2jQrjfiY9OSkPmmi--NoPsVK0eX8CiuIzsONgZowMYGRccmq5we8a5_akJoRfjrpWDH_KiyDyPf6G7CvLffppyQ_PqIWkiQftfZsTSN9PSL7ulyeWVUZ1baV3qVvhOy2fgFP46JGoMk6N4RgoULoEunSMg_XwS3Lwm8Ur_Gdjy2k9C0-5hbFurM2qF0Pc_LmQ",
		"https://lh3.googleusercontent.com/aida-public/AB6AXuCGxMJ3Vbn1ub7cB0g-yov0XHsB82EN_gm6WvfPze3mDMWt0Be4PrZfBP3srme5gMthARGjCM4g80FJUrWjIcr85GgLdcZY09dgEDhUvwXdTYqaHkQltOJ1Szx0R7RqiGFYPlAJSsOzNLvoP6uyt_kuFgOJHioX26lYbZRszRdgIzHxtfUZRbpkbBkfG7zQwH_1q1EZKOWLXUgAzXoTtO5caAljbUNTP9Ql6eQywxSLDl_D6qWWAlWwvaiWXv-_syp6JUs42kjjCw",
		"https://lh3.googleusercontent.com/aida-public/AB6AXuDDXAZaNJFduvT28G8m_w2rzPjosnZahGKBv7HTEP9JD_NW3xYeah3jUjtAHKOO2mLgnneif6CxCrOZ7yOX7ntf488bAYBpBXHjYIrEu_T-FnBPdgPZm20MLQX5i3LC9LHVwmItRbf3B4WyyRBFjiamSmbhvHkD6_iqpxKCKDph43qZZ407BL-IPJZjs2aGv09r_XnSqOoXVJOEXXMr0Dv1Bnjoqc9UwMnPwUadY2iL5-nu6_meVxtVKP_foar17ueIcTOAphtdSA",
		"https://lh3.googleusercontent.com/aida-public/AB6AXuDEKxJLXMTHIj5B1TZnM8dm0CUvNQ0_WOPoVJkvtUR5AiuPL2ochaB7S634ANaQ3oQk-9_Ncx5I8nHYfu6heX0jmHSLpwsRGAE6NJjUIxhjB4EeP40avzjVY6iRH5ZaS083jrWfD3AZmDmMJ5r6U9nNML6canj0Wp8aLA0MD9oI3rheZsXtsuXj6hMkYgVzX-RCE8t9B3KK_geFCfYpLyumzDHPKRd3QvvYAHlwXtCpBRnEJlrBzDJzbh43cLKvqBMRSaFUqeGyjA",
		"https://lh3.googleusercontent.com/aida-public/AB6AXuBI5A0mRC9gQ3EmvDOinUJNDL_0vUr092g8b3UTtslU2aEIa-vgsV_cv-0hcUQY_6YMQpPdr51pYiDtKfSwldOiZFgUMeRnVcbnteq4qlR4NtmpbzC1ifYYWbNCpzOgd75iFH2hip3pBSMx8OnFNIHczw-fBHWfQsWjkfuisweEkTXTpuwGP9KpZ0S9cwqtRslCKAYHA6R1FTajaYtlX4Al_yEQEB9tdi_hGw19lUJJiX8-hSdly8rs_DUOXbZ9yuiwA9t8gSVlJg",
		"https://lh3.googleusercontent.com/aida-public/AB6AXuDqAmy2uC-lW5Vd0QGvhFW0lHe-l4VHgkaM26SJyc4TjBo7Iiekt9U2gWQZKbAp0EqR05Of7Vaq0FDsDSv3bBwISxY-n-XaSdtJUYpgt0RJC5IyVUlM14dSqtY2TxPLR6RtpbERPkKLP6_zr46gjIlHqeYyd84nPF10PYMQhvUIgO64elcBXTqZGLK6RcpZu9tlz1OtzYoiylM30oIqG7NhW2u1KMV7K9HimeRksgPp_x59KSO4CEjgQYK9VQEWhKiEpZoyPl_kSA",
	}
	featurePlaceholderImage := "https://lh3.googleusercontent.com/aida-public/AB6AXuBHPvk-qLDXw1r7qLlHjrHFhbGlpEAxjPcnr9K5PK2uD1ef5Hr9TzQfd_bOWlwuRGX0nQXNvTCF8ftRVvOPO80TDNYNy6ioeW32rZaOMfpbZ1DVZLu1QgZIV_MXQgHfykZ6WcPbeeIx2YEYYOzt95W6mJnQ4EjFYOVl2FoL8ZbvxvGZZvGHFpdLee4hEd_36gPOBoQ2ZxoGj29TfDIwaQQb8Ggk9eay2K7u_lzHqrIbydVQIMByf8rgVPyl_WRlBig0AelLlaZ3Sw"
	queuePlaceholderImage := "https://lh3.googleusercontent.com/aida-public/AB6AXuCruTFtZ0rFJZzQf9g4t0Comgto8fOdn1Mifi7jqYDg_YgLXrtwONsF69kgq6rwYvG9zcEID5hWm_SrwqLUWDwL5T1XemYYHFP4f34myTJ56-A_9cXmHe2M1vcwXRkr5ZClj9oUX04lsW6RDjCJ4UI_q_dnGoisvT0kkvgrjm45PfYFzv4BxOhyjvXj7D6UPnzeRz47zUM53vDVg_MAx9Wpv7Vk8w1MbcQX4ieDP-gmB7lzBpVGnqUHLcs4pkrlPzCwVjFWR6vlKg"
	cocktailPlaceholder := func(idx int) string {
		if len(cocktailPlaceholders) == 0 {
			return featurePlaceholderImage
		}
		if idx < 0 {
			idx = -idx
		}
		return cocktailPlaceholders[idx%len(cocktailPlaceholders)]
	}
	cocktailCardImage := func(name, imagePath string, idx int) string {
		if imagePath = strings.TrimSpace(imagePath); imagePath != "" {
			return imagePath
		}
		if imagePath = catalog.StitchCocktailImageFor(name); imagePath != "" {
			return imagePath
		}
		return cocktailPlaceholder(idx)
	}
	cocktailHeroImage := func(name, imagePath string) string {
		if imagePath = strings.TrimSpace(imagePath); imagePath != "" {
			return imagePath
		}
		if imagePath = catalog.StitchCocktailImageFor(name); imagePath != "" {
			return imagePath
		}
		return featurePlaceholderImage
	}
	orderCocktailImage := func(name, imagePath string) string {
		if imagePath = strings.TrimSpace(imagePath); imagePath != "" {
			return imagePath
		}
		if imagePath = catalog.StitchCocktailImageFor(name); imagePath != "" {
			return imagePath
		}
		return queuePlaceholderImage
	}
	cocktailAlcoholLabel := func(tags string) string {
		tags = strings.ToLower(tags)
		if strings.Contains(tags, "non-alcoholic") || strings.Contains(tags, "nonalcoholic") || strings.Contains(tags, "non alc") || strings.Contains(tags, "non-alc") || strings.Contains(tags, "na") {
			return "Non-Alc"
		}
		return "Alcoholic"
	}
	productStatusLabel := func(stock *int64, available bool) string {
		if stock != nil {
			if *stock <= 0 {
				return "Out of Stock"
			}
			if *stock <= 2 {
				return "Low Stock"
			}
			return "Available"
		}
		if available {
			return "Available"
		}
		return "Unavailable"
	}
	productUnit := func(name, category string) string {
		v := strings.ToLower(strings.TrimSpace(name + " " + category))
		switch {
		case strings.Contains(v, "ice"):
			return "kg"
		case strings.Contains(v, "mint"), strings.Contains(v, "garnish"), strings.Contains(v, "fresh"):
			return "packs"
		default:
			return "units"
		}
	}
	productIcon := func(name, category string) string {
		v := strings.ToLower(strings.TrimSpace(name + " " + category))
		switch {
		case strings.Contains(v, "bitters"), strings.Contains(v, "wine"), strings.Contains(v, "modifier"):
			return "wine_bar"
		case strings.Contains(v, "vermouth"), strings.Contains(v, "amaro"), strings.Contains(v, "aperitif"):
			return "local_bar"
		case strings.Contains(v, "mint"), strings.Contains(v, "fresh"), strings.Contains(v, "garnish"), strings.Contains(v, "fruit"):
			return "nutrition"
		case strings.Contains(v, "ice"):
			return "ac_unit"
		default:
			return "liquor"
		}
	}
	nextOrderStatus := func(status string) string {
		switch strings.TrimSpace(strings.ToUpper(status)) {
		case "PLACED":
			return "ACCEPTED"
		case "ACCEPTED":
			return "IN_PROGRESS"
		case "IN_PROGRESS":
			return "READY"
		case "READY":
			return "DELIVERED"
		default:
			return ""
		}
	}
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
		"hasPrefix":    strings.HasPrefix,
		"humanizeEnum": humanizeEnum,
		"fmtQty": func(q *float64) string {
			if q == nil {
				return ""
			}
			// render without trailing zeros (150 -> "150", 0.5 -> "0.5")
			return strconv.FormatFloat(*q, 'f', -1, 64)
		},
		"initials": func(name string) string {
			parts := strings.Fields(strings.TrimSpace(name))
			if len(parts) == 0 {
				return "HB"
			}
			if len(parts) == 1 {
				runes := []rune(parts[0])
				if len(runes) >= 2 {
					return strings.ToUpper(string(runes[:2]))
				}
				return strings.ToUpper(parts[0])
			}
			first := []rune(parts[0])
			last := []rune(parts[len(parts)-1])
			return strings.ToUpper(string(first[0]) + string(last[0]))
		},
		"cocktailPlaceholderImage": func(idx int) string {
			return cocktailPlaceholder(idx)
		},
		"cocktailCardImage": func(name, imagePath string, idx int) string {
			return cocktailCardImage(name, imagePath, idx)
		},
		"cocktailHeroImage": func(name, imagePath string) string {
			return cocktailHeroImage(name, imagePath)
		},
		"orderCocktailImage": func(name, imagePath string) string {
			return orderCocktailImage(name, imagePath)
		},
		"dashboardFeatureImage": func(name, imagePath string) string {
			return cocktailHeroImage(name, imagePath)
		},
		"stitchCocktailLabel": func(name string) string {
			return catalog.StitchCocktailLabelFor(name)
		},
		"stitchCocktailAlt": func(name string) string {
			return catalog.StitchCocktailAltFor(name)
		},
		"featurePlaceholderImage": func() string {
			return featurePlaceholderImage
		},
		"queuePlaceholderImage": func() string {
			return queuePlaceholderImage
		},
		"cocktailAlcoholLabel": cocktailAlcoholLabel,
		"stockValue": func(stock *int64) int64 {
			if stock == nil {
				return 0
			}
			return *stock
		},
		"stockIncrement": func(stock *int64) int64 {
			if stock == nil {
				return 1
			}
			return *stock + 1
		},
		"stockDecrement": func(stock *int64) int64 {
			if stock == nil || *stock <= 0 {
				return 0
			}
			return *stock - 1
		},
		"productStatusLabel": productStatusLabel,
		"productUnit":        productUnit,
		"productIcon":        productIcon,
		"nextOrderStatus":    nextOrderStatus,
		"orderPrimaryLabel": func(status string) string {
			switch strings.TrimSpace(strings.ToUpper(status)) {
			case "PLACED":
				return "Accept"
			case "ACCEPTED":
				return "Start"
			case "IN_PROGRESS":
				return "Ready"
			case "READY":
				return "Deliver"
			default:
				return ""
			}
		},
		"orderPrimaryIcon": func(status string) string {
			switch strings.TrimSpace(strings.ToUpper(status)) {
			case "PLACED":
				return "play_arrow"
			case "ACCEPTED":
				return "sports_bar"
			case "IN_PROGRESS":
				return "check"
			case "READY":
				return "done_all"
			default:
				return "more_horiz"
			}
		},
		"orderRailIcon": func(status string) string {
			switch strings.TrimSpace(strings.ToUpper(status)) {
			case "PLACED":
				return "hourglass_empty"
			case "ACCEPTED", "IN_PROGRESS":
				return "schedule"
			case "READY":
				return "check_circle"
			case "DELIVERED":
				return "done_all"
			case "CANCELLED":
				return "close"
			default:
				return "schedule"
			}
		},
		"orderStatusLabel": func(status string) string {
			switch strings.TrimSpace(strings.ToUpper(status)) {
			case "IN_PROGRESS":
				return "Preparing"
			default:
				return humanizeEnum(status)
			}
		},
		"navSearchPlaceholder": func(path string) string {
			switch {
			case path == "/", strings.HasPrefix(path, "/bartender/cocktails"):
				return "Search library..."
			case path == "/bartender", path == "/orders", strings.HasPrefix(path, "/bartender/orders"):
				return "Search orders..."
			case strings.HasPrefix(path, "/admin/users"):
				return "Search users..."
			default:
				return "Search..."
			}
		},
		"showShellSearch": func(path string) bool {
			switch {
			case path == "/", path == "/orders", path == "/bartender", strings.HasPrefix(path, "/bartender/cocktails"), strings.HasPrefix(path, "/bartender/orders"), strings.HasPrefix(path, "/admin/users"):
				return true
			default:
				return false
			}
		},
		"libraryURL": func(path, search, spirit string, page int) string {
			values := url.Values{}
			if search = strings.TrimSpace(search); search != "" {
				values.Set("q", search)
			}
			switch spirit = strings.TrimSpace(strings.ToLower(spirit)); spirit {
			case "", "all", "all-spirits":
				spirit = ""
			case "whiskey", "gin", "tequila", "rum":
			default:
				spirit = ""
			}
			if spirit != "" {
				values.Set("spirit", spirit)
			}
			if page > 1 {
				values.Set("page", strconv.Itoa(page))
			}
			if encoded := values.Encode(); encoded != "" {
				return path + "?" + encoded
			}
			return path
		},
		"inventoryURL": func(path, search, category, status string) string {
			values := url.Values{}
			if search = strings.TrimSpace(search); search != "" {
				values.Set("q", search)
			}
			if category = strings.TrimSpace(category); category != "" {
				values.Set("category", category)
			}
			switch status = strings.TrimSpace(strings.ToLower(status)); status {
			case "", "all":
				status = ""
			case "available", "low", "out", "unavailable":
			default:
				status = ""
			}
			if status != "" {
				values.Set("status", status)
			}
			if encoded := values.Encode(); encoded != "" {
				return path + "?" + encoded
			}
			return path
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

	// Sync catalog defaults on every startup so newly introduced seed items
	// appear in existing installations without overriding live availability edits.
	if err := db.SeedCatalog(store.DB); err != nil {
		a.log.Warn("catalog sync failed", "err", err)
	} else {
		a.log.Info("catalog synced")
	}

	return a, nil
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
func (a *App) Push() *push.Service           { return a.push }
func (a *App) Config() Config                { return a.cfg }
func (a *App) NeedsOnboarding() bool         { return a.needsOnboarding }
func (a *App) ClearOnboarding()              { a.needsOnboarding = false }
