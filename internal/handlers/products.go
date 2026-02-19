package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"house-bartender-go/internal/app"
	"house-bartender-go/internal/db"

	"github.com/go-chi/chi/v5"
)

func (s *Server) ProductCreatePost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := strings.TrimSpace(r.FormValue("name"))
	category := strings.TrimSpace(r.FormValue("category"))
	abvStr := strings.TrimSpace(r.FormValue("abv_percent"))
	allergens := strings.TrimSpace(r.FormValue("allergen_flags"))
	notes := strings.TrimSpace(r.FormValue("notes"))
	isAvail := strings.TrimSpace(r.FormValue("is_available")) == "1"
	stockStr := strings.TrimSpace(r.FormValue("stock_count"))

	if name == "" || category == "" {
		s.App.AddFlash(w, r, app.FlashError, "Name and category are required.")
		s.redirect(w, r, "/bartender/products")
		return
	}

	var abv *float64
	if abvStr != "" {
		if f, err := strconv.ParseFloat(abvStr, 64); err == nil {
			abv = &f
		}
	}

	var stock *int64
	if stockStr != "" {
		if n, err := strconv.ParseInt(stockStr, 10, 64); err == nil {
			stock = &n
		}
	}

	_, err := s.App.Store().Q.CreateProduct(db.CreateProductParams{
		Name:          name,
		Category:      category,
		ABVPercent:    abv,
		AllergenFlags: allergens,
		Notes:         notes,
		IsAvailable:   isAvail,
		StockCount:    stock,
	})
	if err != nil {
		s.App.AddFlash(w, r, app.FlashError, "Could not create product (name might already exist).")
		s.redirect(w, r, "/bartender/products")
		return
	}

	s.broadcastInventory()
	s.App.AddFlash(w, r, app.FlashSuccess, "Product created.")
	s.redirect(w, r, "/bartender/products")
}

func (s *Server) ProductTogglePost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/bartender/products")
		return
	}
	_ = r.ParseForm()
	avail := strings.TrimSpace(r.FormValue("is_available")) == "1"
	_ = s.App.Store().Q.ToggleProductAvailability(id, avail)

	s.broadcastInventory()
	s.redirect(w, r, "/bartender/products")
}

func (s *Server) ProductStockPost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/bartender/products")
		return
	}
	_ = r.ParseForm()
	stockStr := strings.TrimSpace(r.FormValue("stock_count"))

	var stock *int64
	if stockStr != "" {
		if n, err := strconv.ParseInt(stockStr, 10, 64); err == nil {
			stock = &n
		}
	}
	_ = s.App.Store().Q.SetProductStock(id, stock)

	s.broadcastInventory()
	s.redirect(w, r, "/bartender/products")
}

func (s *Server) ProductDeletePost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/bartender/products")
		return
	}
	if err := s.App.Store().Q.DeleteProduct(id); err != nil {
		s.App.AddFlash(w, r, app.FlashError, "Delete failed (product might be used by a cocktail).")
		s.redirect(w, r, "/bartender/products")
		return
	}

	s.broadcastInventory()
	s.App.AddFlash(w, r, app.FlashSuccess, "Product deleted.")
	s.redirect(w, r, "/bartender/products")
}

