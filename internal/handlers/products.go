package handlers

import (
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"house-bartender-go/internal/app"
	"house-bartender-go/internal/db"

	"github.com/go-chi/chi/v5"
)

type productFormInput struct {
	Name          string
	Category      string
	ABVPercent    *float64
	AllergenFlags string
	Notes         string
	IsAvailable   bool
	StockCount    *int64
}

func (s *Server) ProductCreatePost(w http.ResponseWriter, r *http.Request) {
	search, category, status := inventoryFiltersFromRequest(r)
	in, ok := parseProductFormInput(r)
	if !ok {
		s.App.AddFlash(w, r, app.FlashError, "Name and category are required.")
		s.redirect(w, r, inventoryURL("/bartender/products", search, category, status))
		return
	}

	_, err := s.App.Store().Q.CreateProduct(db.CreateProductParams{
		Name:          in.Name,
		Category:      in.Category,
		ABVPercent:    in.ABVPercent,
		AllergenFlags: in.AllergenFlags,
		Notes:         in.Notes,
		IsAvailable:   in.IsAvailable,
		StockCount:    in.StockCount,
	})
	if err != nil {
		s.App.AddFlash(w, r, app.FlashError, "Could not create ingredient (name might already exist).")
		s.redirect(w, r, inventoryURL("/bartender/products", search, category, status))
		return
	}

	s.broadcastInventory()
	s.App.AddFlash(w, r, app.FlashSuccess, "Ingredient created.")
	s.redirect(w, r, inventoryURL("/bartender/products", search, category, status))
}

func (s *Server) ProductTogglePost(w http.ResponseWriter, r *http.Request) {
	search, category, status := inventoryFiltersFromRequest(r)
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, inventoryURL("/bartender/products", search, category, status))
		return
	}
	_ = r.ParseForm()
	avail := formBool(r, "is_available")
	_ = s.App.Store().Q.ToggleProductAvailability(id, avail)

	s.broadcastInventory()
	if r.Header.Get("HX-Request") != "" {
		s.renderProductsTablePartial(w, r)
		return
	}
	s.redirect(w, r, inventoryURL("/bartender/products", search, category, status))
}

func (s *Server) ProductEditGet(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/bartender/products")
		return
	}

	p, _ := s.App.Store().Q.GetProductByID(id)
	if p == nil {
		s.redirect(w, r, "/bartender/products")
		return
	}

	search, category, status := inventoryFiltersFromRequest(r)
	page := s.buildBartenderProductsPage(search, category, status, productFormStateFromProduct(*p, inventoryURL("/bartender/products/"+idStr+"/edit", search, category, status)))
	s.renderLayout(w, r, "Edit Ingredient", "bartender_products.html", page)
}

func (s *Server) ProductEditPost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/bartender/products")
		return
	}

	search, category, status := inventoryFiltersFromRequest(r)
	existing, _ := s.App.Store().Q.GetProductByID(id)
	if existing == nil {
		s.redirect(w, r, inventoryURL("/bartender/products", search, category, status))
		return
	}

	in, ok := parseProductFormInput(r)
	if !ok {
		s.App.AddFlash(w, r, app.FlashError, "Name and category are required.")
		s.redirect(w, r, inventoryURL("/bartender/products/"+idStr+"/edit", search, category, status))
		return
	}

	if err := s.App.Store().Q.UpdateProduct(db.UpdateProductParams{
		ID:            id,
		Name:          in.Name,
		Category:      in.Category,
		ABVPercent:    in.ABVPercent,
		AllergenFlags: in.AllergenFlags,
		Notes:         in.Notes,
		IsAvailable:   in.IsAvailable,
		StockCount:    in.StockCount,
	}); err != nil {
		s.App.AddFlash(w, r, app.FlashError, "Update failed (name might already exist).")
		s.redirect(w, r, inventoryURL("/bartender/products/"+idStr+"/edit", search, category, status))
		return
	}

	s.broadcastInventory()
	s.App.AddFlash(w, r, app.FlashSuccess, "Ingredient updated.")
	s.redirect(w, r, inventoryURL("/bartender/products", search, category, status))
}

func (s *Server) ProductStockPost(w http.ResponseWriter, r *http.Request) {
	search, category, status := inventoryFiltersFromRequest(r)
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, inventoryURL("/bartender/products", search, category, status))
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
	if r.Header.Get("HX-Request") != "" {
		s.renderProductsTablePartial(w, r)
		return
	}
	s.redirect(w, r, inventoryURL("/bartender/products", search, category, status))
}

func (s *Server) ProductDeletePost(w http.ResponseWriter, r *http.Request) {
	search, category, status := inventoryFiltersFromRequest(r)
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, inventoryURL("/bartender/products", search, category, status))
		return
	}
	if err := s.App.Store().Q.DeleteProduct(id); err != nil {
		s.App.AddFlash(w, r, app.FlashError, "Delete failed (ingredient might be used by a cocktail).")
		s.redirect(w, r, inventoryURL("/bartender/products", search, category, status))
		return
	}

	s.broadcastInventory()
	if r.Header.Get("HX-Request") != "" {
		s.renderProductsTablePartial(w, r)
		return
	}
	s.App.AddFlash(w, r, app.FlashSuccess, "Ingredient deleted.")
	s.redirect(w, r, inventoryURL("/bartender/products", search, category, status))
}

func (s *Server) renderProductsTablePartial(w http.ResponseWriter, r *http.Request) {
	search, category, status := inventoryFiltersFromRequest(r)
	s.renderPartial(w, r, "products_table.html", s.buildBartenderProductsPage(search, category, status, ProductFormState{}), "/bartender/products")
}

func parseProductFormInput(r *http.Request) (productFormInput, bool) {
	_ = r.ParseForm()
	in := productFormInput{
		Name:          strings.TrimSpace(r.FormValue("name")),
		Category:      strings.TrimSpace(r.FormValue("category")),
		AllergenFlags: strings.TrimSpace(r.FormValue("allergen_flags")),
		Notes:         strings.TrimSpace(r.FormValue("notes")),
		IsAvailable:   formBool(r, "is_available"),
	}

	if in.Name == "" || in.Category == "" {
		return in, false
	}

	if abvStr := strings.TrimSpace(r.FormValue("abv_percent")); abvStr != "" {
		if f, err := strconv.ParseFloat(abvStr, 64); err == nil {
			in.ABVPercent = &f
		}
	}

	if stockStr := strings.TrimSpace(r.FormValue("stock_count")); stockStr != "" {
		if n, err := strconv.ParseInt(stockStr, 10, 64); err == nil {
			in.StockCount = &n
		}
	}

	return in, true
}

func productFormStateFromProduct(p db.Product, action string) ProductFormState {
	return ProductFormState{
		Mode:          "edit",
		Action:        action,
		Name:          p.Name,
		Category:      p.Category,
		ABVPercent:    floatPtrToString(p.ABVPercent),
		AllergenFlags: p.AllergenFlags,
		Notes:         p.Notes,
		StockCount:    int64PtrToString(p.StockCount),
		IsAvailable:   p.IsAvailable,
	}
}

func floatPtrToString(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}

func int64PtrToString(v *int64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatInt(*v, 10)
}

func inventoryFiltersFromRequest(r *http.Request) (string, string, string) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	return q, "", ""
}

func normalizeInventoryStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "", "all":
		return ""
	case "available", "low", "out", "unavailable":
		return strings.TrimSpace(strings.ToLower(raw))
	default:
		return ""
	}
}

func productCategories(products []db.Product) []string {
	seen := map[string]struct{}{}
	var categories []string
	for _, product := range products {
		category := strings.TrimSpace(product.Category)
		if category == "" {
			continue
		}
		if _, ok := seen[category]; ok {
			continue
		}
		seen[category] = struct{}{}
		categories = append(categories, category)
	}
	sort.Strings(categories)
	return categories
}

func filterProducts(products []db.Product, category, status string) []db.Product {
	category = strings.TrimSpace(category)
	status = normalizeInventoryStatus(status)
	if category == "" && status == "" {
		return products
	}

	filtered := make([]db.Product, 0, len(products))
	for _, product := range products {
		if category != "" && !strings.EqualFold(strings.TrimSpace(product.Category), category) {
			continue
		}

		label := inventoryStatusLabel(product.StockCount, product.ComputedAvail)
		switch status {
		case "available":
			if label != "Available" {
				continue
			}
		case "low":
			if label != "Low Stock" {
				continue
			}
		case "out":
			if label != "Out of Stock" {
				continue
			}
		case "unavailable":
			if label != "Unavailable" {
				continue
			}
		}

		filtered = append(filtered, product)
	}

	return filtered
}

func inventoryURL(path, search, category, status string) string {
	values := url.Values{}
	if search = strings.TrimSpace(search); search != "" {
		values.Set("q", search)
	}
	if category = strings.TrimSpace(category); category != "" {
		values.Set("category", category)
	}
	if status = normalizeInventoryStatus(status); status != "" {
		values.Set("status", status)
	}
	if encoded := values.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}

func inventoryStatusLabel(stock *int64, available bool) string {
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
