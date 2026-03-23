package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"house-bartender-go/internal/app"
	"house-bartender-go/internal/db"

	"github.com/go-chi/chi/v5"
)

const cocktailLibraryPageSize = 10

type CocktailLibraryPage struct {
	Search       string
	Spirit       string
	Cocktails    []db.Cocktail
	CurrentPage  int
	TotalPages   int
	PageNumbers  []int
	TotalCount   int
	ShowingStart int
	ShowingEnd   int
	HasPrev      bool
	HasNext      bool
	PrevPage     int
	NextPage     int
}

type CocktailDetailPage struct {
	Cocktail    db.Cocktail
	Ingredients []db.CocktailIngredient
	TagList     []string
	IsAvailable bool
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

	page := s.buildCocktailLibraryPage(r, true)
	s.renderLayout(w, r, "Cocktails", "user_home.html", page)
}

func (s *Server) UserCocktailsPartialGet(w http.ResponseWriter, r *http.Request) {
	page := s.buildCocktailLibraryPage(r, true)
	s.renderPartial(w, r, "library_results.html", page, "/")
}

func (s *Server) buildCocktailLibraryPage(r *http.Request, onlyAvailable bool) CocktailLibraryPage {
	q := r.URL.Query()
	search := strings.TrimSpace(q.Get("q"))
	spirit := normalizeSpiritFilter(q.Get("spirit"))
	pageNumber := parsePositiveInt(q.Get("page"), 1)

	cocks, _ := s.App.Store().Q.ListCocktailsComputed(onlyAvailable)
	var (
		out             []db.Cocktail
		ingredientCache = map[int64][]db.CocktailIngredient{}
	)

	loadIngredients := func(cocktailID int64) []db.CocktailIngredient {
		if cached, ok := ingredientCache[cocktailID]; ok {
			return cached
		}
		ings, _ := s.App.Store().Q.GetCocktailIngredients(cocktailID)
		ingredientCache[cocktailID] = ings
		return ings
	}

	for _, c := range cocks {
		var ings []db.CocktailIngredient
		if search != "" || spirit != "" {
			ings = loadIngredients(c.ID)
		}
		if !cocktailMatchesSearch(c, ings, search) {
			continue
		}
		if !cocktailMatchesSpirit(c, ings, spirit) {
			continue
		}
		out = append(out, c)
	}

	totalCount := len(out)
	totalPages := 1
	if totalCount > 0 {
		totalPages = (totalCount + cocktailLibraryPageSize - 1) / cocktailLibraryPageSize
	}
	if pageNumber > totalPages {
		pageNumber = totalPages
	}
	if pageNumber < 1 {
		pageNumber = 1
	}

	start := 0
	end := totalCount
	if totalCount > 0 {
		start = (pageNumber - 1) * cocktailLibraryPageSize
		if start > totalCount {
			start = totalCount
		}
		end = start + cocktailLibraryPageSize
		if end > totalCount {
			end = totalCount
		}
		out = out[start:end]
	}

	pageNumbers := make([]int, 0, totalPages)
	for i := 1; i <= totalPages; i++ {
		pageNumbers = append(pageNumbers, i)
	}

	showingStart := 0
	showingEnd := 0
	if totalCount > 0 {
		showingStart = start + 1
		showingEnd = end
	}

	return CocktailLibraryPage{
		Search:       search,
		Spirit:       spirit,
		Cocktails:    out,
		CurrentPage:  pageNumber,
		TotalPages:   totalPages,
		PageNumbers:  pageNumbers,
		TotalCount:   totalCount,
		ShowingStart: showingStart,
		ShowingEnd:   showingEnd,
		HasPrev:      pageNumber > 1,
		HasNext:      pageNumber < totalPages,
		PrevPage:     pageNumber - 1,
		NextPage:     pageNumber + 1,
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

func parsePositiveInt(s string, fallback int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
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

func cocktailMatchesSearch(c db.Cocktail, ings []db.CocktailIngredient, search string) bool {
	search = strings.ToLower(strings.TrimSpace(search))
	if search == "" {
		return true
	}

	fields := []string{
		c.Name,
		c.Description,
		c.Tags,
		c.Difficulty,
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), search) {
			return true
		}
	}

	for _, it := range ings {
		if strings.Contains(strings.ToLower(it.ProductName), search) {
			return true
		}
	}

	return false
}

func normalizeSpiritFilter(s string) string {
	switch normalized := strings.TrimSpace(strings.ToLower(s)); normalized {
	case "", "all", "all-spirits":
		return ""
	case "whiskey", "gin", "tequila", "rum":
		return normalized
	default:
		return ""
	}
}

func cocktailMatchesSpirit(c db.Cocktail, ings []db.CocktailIngredient, spirit string) bool {
	spirit = normalizeSpiritFilter(spirit)
	if spirit == "" {
		return true
	}

	text := strings.ToLower(strings.Join([]string{c.Name, c.Description, c.Tags}, " "))
	for _, token := range spiritTokens(spirit) {
		if strings.Contains(text, token) {
			return true
		}
	}

	for _, it := range ings {
		ingredientText := strings.ToLower(strings.Join([]string{it.ProductName, it.ProductCategory}, " "))
		for _, token := range spiritTokens(spirit) {
			if strings.Contains(ingredientText, token) {
				return true
			}
		}
	}

	return false
}

func spiritTokens(spirit string) []string {
	switch spirit {
	case "whiskey":
		return []string{"whiskey", "whisky", "bourbon", "rye"}
	case "gin":
		return []string{"gin"}
	case "tequila":
		return []string{"tequila"}
	case "rum":
		return []string{"rum"}
	default:
		return nil
	}
}
