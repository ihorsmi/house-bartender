package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"house-bartender-go/internal/app"
	"house-bartender-go/internal/db"
)

type BartenderDashboardPage struct {
	OnDuty          bool
	CountPlaced     int
	CountAccepted   int
	CountInProgress int
	CountReady      int
	Orders          []db.Order
	AverageOrderAge string
	ClaimedCoverage int
	ReadyCoverage   int
	FeaturedName    string
	FeaturedContext string
	FeaturedImage   string
	Push            BartenderPushPage
}

type BartenderPushPage struct {
	Show               bool
	Configured         bool
	PublicKey          string
	EnabledDeviceCount int
}

type BartenderProductsPage struct {
	Search     string
	Category   string
	Status     string
	Categories []string
	Products   []db.Product
	Form       ProductFormState
}

type ProductFormState struct {
	Mode          string // "new" or "edit"
	Action        string
	Name          string
	Category      string
	ABVPercent    string
	AllergenFlags string
	Notes         string
	StockCount    string
	IsAvailable   bool
}

type BartenderOrdersPage struct {
	Mode       string // "bartender"
	Orders     []db.Order
	Bartenders []db.User
	Events     map[int64][]db.OrderEvent // optional
}

type DashboardOrdersPreviewPage struct {
	Orders []db.Order
}

func (s *Server) BartenderDashboardGet(w http.ResponseWriter, r *http.Request) {
	u := s.App.CurrentUser(r)
	if u == nil {
		s.redirect(w, r, "/login")
		return
	}

	queue := s.listBartenderQueue()
	var cPlaced, cAcc, cProg, cReady int
	var claimed int
	var totalAge time.Duration
	featuredName := "Queue clarity in motion"
	featuredContext := "Open the full queue to assign, advance, and clear orders without breaking live updates."
	featuredImage := ""
	for _, o := range queue {
		switch o.Status {
		case "PLACED":
			cPlaced++
		case "ACCEPTED":
			cAcc++
		case "IN_PROGRESS":
			cProg++
		case "READY":
			cReady++
		}

		if o.Status != "PLACED" || o.AssignedBartenderID != nil {
			claimed++
		}
		totalAge += time.Since(o.CreatedAt)

		if featuredImage == "" && strings.TrimSpace(o.CocktailImagePath) != "" {
			featuredImage = o.CocktailImagePath
			featuredName = o.CocktailName
			if location := strings.TrimSpace(o.Location); location != "" {
				featuredContext = location
			}
		}
	}

	totalOrders := len(queue)
	page := BartenderDashboardPage{
		OnDuty:          u.OnDuty,
		CountPlaced:     cPlaced,
		CountAccepted:   cAcc,
		CountInProgress: cProg,
		CountReady:      cReady,
		Orders:          limitOrders(queue, 5),
		AverageOrderAge: averageOrderAge(totalAge, totalOrders),
		ClaimedCoverage: percent(claimed, totalOrders),
		ReadyCoverage:   percent(cReady, totalOrders),
		FeaturedName:    featuredName,
		FeaturedContext: featuredContext,
		FeaturedImage:   featuredImage,
	}
	s.renderLayout(w, r, "Bartender", "bartender_dashboard.html", page)
}

func (s *Server) BartenderDutyPost(w http.ResponseWriter, r *http.Request) {
	u := s.App.CurrentUser(r)
	if u == nil {
		s.redirect(w, r, "/login")
		return
	}
	newDuty := !u.OnDuty
	_ = s.App.Store().Q.SetUserDuty(u.ID, newDuty)
	if newDuty {
		s.App.AddFlash(w, r, app.FlashSuccess, "You are now On Duty.")
	} else {
		s.App.AddFlash(w, r, app.FlashInfo, "You are now Off Duty.")
	}
	s.redirect(w, r, "/bartender")
}

func (s *Server) BartenderProductsGet(w http.ResponseWriter, r *http.Request) {
	search, category, status := inventoryFiltersFromRequest(r)
	page := s.buildBartenderProductsPage(search, category, status, defaultProductFormState(inventoryURL("/bartender/products", search, category, status)))
	s.renderLayout(w, r, "Ingredients", "bartender_products.html", page)
}

func (s *Server) BartenderProductsPartialGet(w http.ResponseWriter, r *http.Request) {
	search, category, status := inventoryFiltersFromRequest(r)
	page := s.buildBartenderProductsPage(search, category, status, ProductFormState{})
	s.renderPartial(w, r, "products_table.html", page, "/bartender/products")
}

func defaultProductFormState(action string) ProductFormState {
	return ProductFormState{
		Mode:        "new",
		Action:      action,
		IsAvailable: true,
	}
}

func (s *Server) BartenderCocktailsGet(w http.ResponseWriter, r *http.Request) {
	page := s.buildCocktailLibraryPage(r, false)
	s.renderLayout(w, r, "Cocktails", "bartender_cocktails.html", page)
}

func (s *Server) BartenderCocktailsPartialGet(w http.ResponseWriter, r *http.Request) {
	page := s.buildCocktailLibraryPage(r, false)
	s.renderPartial(w, r, "library_results.html", page, "/bartender/cocktails")
}

func (s *Server) BartenderOrdersGet(w http.ResponseWriter, r *http.Request) {
	page := s.buildBartenderOrdersPage("bartender")
	s.renderLayout(w, r, "Orders", "bartender_orders.html", page)
}

func (s *Server) BartenderOrdersPartialGet(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(r.URL.Query().Get("view")) == "dashboard" {
		page := DashboardOrdersPreviewPage{Orders: limitOrders(s.listBartenderQueue(), 5)}
		s.renderPartial(w, r, "dashboard_orders_preview.html", page, "/bartender")
		return
	}

	page := s.buildBartenderOrdersPage("bartender")
	s.renderPartial(w, r, "orders_list.html", page, "/bartender/orders")
}

func (s *Server) buildBartenderOrdersPage(mode string) BartenderOrdersPage {
	orders := s.listBartenderQueue()
	users, _ := s.App.Store().Q.ListUsers()

	var bartenders []db.User
	for _, u := range users {
		if u.Role == app.RoleBartender && u.IsActive {
			bartenders = append(bartenders, u)
		}
	}

	return BartenderOrdersPage{
		Mode:       mode,
		Orders:     orders,
		Bartenders: bartenders,
		Events:     map[int64][]db.OrderEvent{},
	}
}

func (s *Server) listBartenderQueue() []db.Order {
	orders, _ := s.App.Store().Q.ListOrderQueue()
	return orders
}

func limitOrders[T any](items []T, limit int) []T {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

func percent(value, total int) int {
	if value <= 0 || total <= 0 {
		return 0
	}
	return (value*100 + total/2) / total
}

func averageOrderAge(totalAge time.Duration, count int) string {
	if count <= 0 {
		return "No queue"
	}
	return compactDuration(totalAge / time.Duration(count))
}

func compactDuration(d time.Duration) string {
	if d < time.Minute {
		return "Just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Round(time.Minute)/time.Minute))
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if minutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func (s *Server) buildBartenderProductsPage(search, category, status string, form ProductFormState) BartenderProductsPage {
	products, _ := s.App.Store().Q.ListProducts(search)
	allProducts, _ := s.App.Store().Q.ListProducts("")

	return BartenderProductsPage{
		Search:     search,
		Category:   category,
		Status:     status,
		Categories: productCategories(allProducts),
		Products:   filterProducts(products, category, status),
		Form:       form,
	}
}
