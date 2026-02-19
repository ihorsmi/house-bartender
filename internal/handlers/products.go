package handlers

import (
	"net/http"
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
	in, ok := parseProductFormInput(r)
	if !ok {
		s.App.AddFlash(w, r, app.FlashError, "Name and category are required.")
		s.redirect(w, r, "/bartender/products")
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
	avail := formBool(r, "is_available")
	_ = s.App.Store().Q.ToggleProductAvailability(id, avail)

	s.broadcastInventory()
	s.redirect(w, r, "/bartender/products")
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

	search := r.URL.Query().Get("q")
	products, _ := s.App.Store().Q.ListProducts(search)
	page := BartenderProductsPage{
		Search:   search,
		Products: products,
		Form:     productFormStateFromProduct(*p),
	}
	s.renderLayout(w, r, "Edit Product", "bartender_products.html", page)
}

func (s *Server) ProductEditPost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/bartender/products")
		return
	}

	existing, _ := s.App.Store().Q.GetProductByID(id)
	if existing == nil {
		s.redirect(w, r, "/bartender/products")
		return
	}

	in, ok := parseProductFormInput(r)
	if !ok {
		s.App.AddFlash(w, r, app.FlashError, "Name and category are required.")
		s.redirect(w, r, "/bartender/products/"+idStr+"/edit")
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
		s.redirect(w, r, "/bartender/products/"+idStr+"/edit")
		return
	}

	s.broadcastInventory()
	s.App.AddFlash(w, r, app.FlashSuccess, "Product updated.")
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

func productFormStateFromProduct(p db.Product) ProductFormState {
	return ProductFormState{
		Mode:          "edit",
		Action:        "/bartender/products/" + strconv.FormatInt(p.ID, 10) + "/edit",
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

