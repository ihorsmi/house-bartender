package handlers

import (
	"net/http"
	"strings"

	"house-bartender-go/internal/app"
	"house-bartender-go/internal/db"

	"github.com/go-chi/chi/v5"
)

type UserHomePage struct {
	FilterAlcohol string
	FilterTag     string
	IncludeIng    string
	ExcludeIng    string
	Cocktails     []db.Cocktail
}

type CocktailDetailPage struct {
	Cocktail     db.Cocktail
	Ingredients  []db.CocktailIngredient
	TagList      []string
	IsAvailable  bool
}

type UserOrdersPage struct {
	Mode   string // "user"
	Orders []db.Order
	Events map[int64][]db.OrderEvent
}

func (s *Server) UserHomeGet(w http.ResponseWriter, r *http.Request) {
	u := s.App.CurrentUser(r)
	if u == nil {
		s.redirect(w, r, "/login")
		return
	}

	// Only USER should use the ordering portal.
	// Bartenders/Admins should live in their own portals.
	if u.Role == app.RoleBartender {
		s.redirect(w, r, "/bartender")
		return
	}
	if u.Role == app.RoleAdmin {
		s.redirect(w, r, "/admin/users")
		return
	}

	page := s.buildUserCocktailsPage(r)
	s.renderLayout(w, r, "Cocktails", "user_home.html", page)
}

func (s *Server) UserCocktailsPartialGet(w http.ResponseWriter, r *http.Request) {
	page := s.buildUserCocktailsPage(r)
	s.renderPartial(w, r, "cocktails_table.html", map[string]any{
		"Cocktails": page.Cocktails,
	}, "")
}

func (s *Server) buildUserCocktailsPage(r *http.Request) UserHomePage {
	q := r.URL.Query()
	alc := strings.TrimSpace(q.Get("alc"))
	if alc == "" {
		alc = "all"
	}
	tag := strings.TrimSpace(q.Get("tag"))
	inc := strings.TrimSpace(q.Get("include"))
	exc := strings.TrimSpace(q.Get("exclude"))

	cocks, _ := s.App.Store().Q.ListCocktailsComputed(true) // user sees available
	// Apply filters
	var out []db.Cocktail
	for _, c := range cocks {
		if !matchAlcohol(alc, c.Tags) {
			continue
		}
		if tag != "" && !containsToken(c.Tags, tag) {
			continue
		}

		if inc != "" || exc != "" {
			ings, _ := s.App.Store().Q.GetCocktailIngredients(c.ID)
			if inc != "" && !ingredientMatch(ings, inc) {
				continue
			}
			if exc != "" && ingredientMatch(ings, exc) {
				continue
			}
		}

		out = append(out, c)
	}

	return UserHomePage{
		FilterAlcohol: alc,
		FilterTag:     tag,
		IncludeIng:    inc,
		ExcludeIng:    exc,
		Cocktails:     out,
	}
}

func (s *Server) CocktailDetailGet(w http.ResponseWriter, r *http.Request) {
	if s.App.CurrentUser(r) == nil {
		s.redirect(w, r, "/login")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		http.NotFound(w, r)
		return
	}

	c, err := s.App.Store().Q.GetCocktailByID(id)
	if err != nil || c == nil {
		http.NotFound(w, r)
		return
	}

	ings, _ := s.App.Store().Q.GetCocktailIngredients(c.ID)

	avail := c.IsEnabled
	for _, it := range ings {
		if it.Required && !it.ProductAvail {
			avail = false
			break
		}
	}

	page := CocktailDetailPage{
		Cocktail:    *c,
		Ingredients: ings,
		TagList:     splitCSV(c.Tags),
		IsAvailable: avail,
	}

	s.renderLayout(w, r, c.Name, "cocktail_detail.html", page)
}

func (s *Server) UserOrdersGet(w http.ResponseWriter, r *http.Request) {
	u := s.App.CurrentUser(r)
	if u == nil {
		s.redirect(w, r, "/login")
		return
	}

	page := s.buildUserOrdersPage(u.ID)
	s.renderLayout(w, r, "My Orders", "user_orders.html", page)
}

func (s *Server) UserOrdersPartialGet(w http.ResponseWriter, r *http.Request) {
	u := s.App.CurrentUser(r)
	if u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	page := s.buildUserOrdersPage(u.ID)
	s.renderPartial(w, r, "orders_list.html", page, "")
}

func (s *Server) buildUserOrdersPage(userID int64) UserOrdersPage {
	orders, _ := s.App.Store().Q.ListOrdersForUser(userID)
	events := map[int64][]db.OrderEvent{}
	for _, o := range orders {
		evs, _ := s.App.Store().Q.ListOrderEvents(o.ID)
		events[o.ID] = evs
	}
	return UserOrdersPage{
		Mode:   "user",
		Orders: orders,
		Events: events,
	}
}

/* ---- helpers ---- */

func parseInt64(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	var n int64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, false
		}
		n = n*10 + int64(ch-'0')
	}
	return n, n > 0
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func normToken(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

func containsToken(csvTags, query string) bool {
	q := normToken(query)
	for _, t := range splitCSV(csvTags) {
		if normToken(t) == q {
			return true
		}
		// allow substring match
		if strings.Contains(normToken(t), q) {
			return true
		}
	}
	return false
}

func matchAlcohol(mode string, tags string) bool {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode == "" || mode == "all" {
		return true
	}
	isNon := containsToken(tags, "non-alcoholic") || containsToken(tags, "nonalcoholic") || containsToken(tags, "na")
	if mode == "non" {
		return isNon
	}
	if mode == "alcohol" {
		return !isNon
	}
	return true
}

func ingredientMatch(ings []db.CocktailIngredient, q string) bool {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return false
	}
	for _, it := range ings {
		if strings.Contains(strings.ToLower(it.ProductName), q) {
			return true
		}
	}
	return false
}
