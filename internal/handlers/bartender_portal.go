package handlers

import (
	"net/http"

	"house-bartender-go/internal/app"
	"house-bartender-go/internal/db"
)

type BartenderDashboardPage struct {
	OnDuty         bool
	CountPlaced    int
	CountAccepted  int
	CountInProgress int
	CountReady     int
}

type BartenderProductsPage struct {
	Search   string
	Products []db.Product
}

type BartenderCocktailsPage struct {
	Cocktails []db.Cocktail
}

type BartenderOrdersPage struct {
	Mode      string // "bartender"
	Orders    []db.Order
	Bartenders []db.User
	Events    map[int64][]db.OrderEvent // optional
}

func (s *Server) BartenderDashboardGet(w http.ResponseWriter, r *http.Request) {
	u := s.App.CurrentUser(r)
	if u == nil {
		s.redirect(w, r, "/login")
		return
	}

	queue, _ := s.App.Store().Q.ListOrderQueue()
	var cPlaced, cAcc, cProg, cReady int
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
	}

	page := BartenderDashboardPage{
		OnDuty:          u.OnDuty,
		CountPlaced:     cPlaced,
		CountAccepted:   cAcc,
		CountInProgress: cProg,
		CountReady:      cReady,
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
	search := r.URL.Query().Get("q")
	products, _ := s.App.Store().Q.ListProducts(search)
	page := BartenderProductsPage{Search: search, Products: products}
	s.renderLayout(w, r, "Products", "bartender_products.html", page)
}

func (s *Server) BartenderProductsPartialGet(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	products, _ := s.App.Store().Q.ListProducts(search)
	s.renderPartial(w, r, "products_table.html", BartenderProductsPage{Search: search, Products: products}, "")
}

func (s *Server) BartenderCocktailsGet(w http.ResponseWriter, r *http.Request) {
	cocks, _ := s.App.Store().Q.ListCocktailsComputed(false)
	page := BartenderCocktailsPage{Cocktails: cocks}
	s.renderLayout(w, r, "Cocktails", "bartender_cocktails.html", page)
}

func (s *Server) BartenderCocktailsPartialGet(w http.ResponseWriter, r *http.Request) {
	cocks, _ := s.App.Store().Q.ListCocktailsComputed(false)
	// Path override so cocktails_table renders bartender controls (it checks prefix "/bartender")
	s.renderPartial(w, r, "cocktails_table.html", BartenderCocktailsPage{Cocktails: cocks}, "/bartender/cocktails")
}

func (s *Server) BartenderOrdersGet(w http.ResponseWriter, r *http.Request) {
	page := s.buildBartenderOrdersPage()
	s.renderLayout(w, r, "Orders", "bartender_orders.html", page)
}

func (s *Server) BartenderOrdersPartialGet(w http.ResponseWriter, r *http.Request) {
	page := s.buildBartenderOrdersPage()
	s.renderPartial(w, r, "orders_list.html", page, "/bartender/orders")
}

func (s *Server) buildBartenderOrdersPage() BartenderOrdersPage {
	orders, _ := s.App.Store().Q.ListOrderQueue()
	users, _ := s.App.Store().Q.ListUsers()

	var bartenders []db.User
	for _, u := range users {
		if u.Role == app.RoleBartender && u.IsActive {
			bartenders = append(bartenders, u)
		}
	}

	return BartenderOrdersPage{
		Mode:       "bartender",
		Orders:     orders,
		Bartenders: bartenders,
		Events:     map[int64][]db.OrderEvent{},
	}
}
