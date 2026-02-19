package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"house-bartender-go/internal/app"
	"house-bartender-go/internal/db"

	"github.com/go-chi/chi/v5"
)

type CocktailFormRow struct {
	ProductID   int64
	QuantityStr string
	Unit        string
	Required    bool
}

type CocktailFormPage struct {
	Mode     string // "new" or "edit"
	Cocktail db.Cocktail
	Tags     string
	Products []db.Product
	IngRows  []CocktailFormRow
}

func (s *Server) CocktailNewGet(w http.ResponseWriter, r *http.Request) {
	products, _ := s.App.Store().Q.ListProducts("")
	page := CocktailFormPage{
		Mode:     "new",
		Cocktail: db.Cocktail{Difficulty: "easy", PrepTimeMinutes: 5, IsEnabled: true},
		Products: products,
		IngRows:  make([]CocktailFormRow, 10),
	}
	s.renderLayout(w, r, "New Cocktail", "cocktail_form.html", page)
}

func (s *Server) CocktailNewPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseMultipartForm(10 << 20)

	c, items, imagePath, ok := s.parseCocktailForm(w, r, 0, "")
	if !ok {
		s.redirect(w, r, "/bartender/cocktails/new")
		return
	}
	c.ImagePath = imagePath

	id, err := s.App.Store().Q.CreateCocktail(db.CreateCocktailParams{
		Name:            c.Name,
		Description:     c.Description,
		ImagePath:       c.ImagePath,
		Tags:            c.Tags,
		Difficulty:      c.Difficulty,
		PrepTimeMinutes: c.PrepTimeMinutes,
		Instructions:    c.Instructions,
		IsEnabled:       c.IsEnabled,
	})
	if err != nil {
		s.App.AddFlash(w, r, app.FlashError, "Could not create cocktail (name might already exist).")
		s.redirect(w, r, "/bartender/cocktails/new")
		return
	}

	_ = s.App.Store().Q.ReplaceCocktailIngredients(id, items)
	s.broadcastInventory()
	s.App.AddFlash(w, r, app.FlashSuccess, "Cocktail created.")
	s.redirect(w, r, "/bartender/cocktails")
}

func (s *Server) CocktailEditGet(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		http.NotFound(w, r)
		return
	}

	c, _ := s.App.Store().Q.GetCocktailByID(id)
	if c == nil {
		http.NotFound(w, r)
		return
	}

	products, _ := s.App.Store().Q.ListProducts("")
	ings, _ := s.App.Store().Q.GetCocktailIngredients(id)

	rows := make([]CocktailFormRow, 10)
	for i := 0; i < len(ings) && i < len(rows); i++ {
		q := ""
		if ings[i].Quantity != nil {
			q = fmt.Sprintf("%.0f", *ings[i].Quantity)
		}
		rows[i] = CocktailFormRow{
			ProductID:   ings[i].ProductID,
			QuantityStr: q,
			Unit:        ings[i].Unit,
			Required:    ings[i].Required,
		}
	}

	page := CocktailFormPage{
		Mode:     "edit",
		Cocktail: *c,
		Tags:     c.Tags,
		Products: products,
		IngRows:  rows,
	}
	s.renderLayout(w, r, "Edit Cocktail", "cocktail_form.html", page)
}

func (s *Server) CocktailEditPost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/bartender/cocktails")
		return
	}

	existing, _ := s.App.Store().Q.GetCocktailByID(id)
	if existing == nil {
		s.redirect(w, r, "/bartender/cocktails")
		return
	}

	_ = r.ParseMultipartForm(10 << 20)

	c, items, imagePath, ok := s.parseCocktailForm(w, r, id, existing.ImagePath)
	if !ok {
		s.redirect(w, r, "/bartender/cocktails/"+idStr+"/edit")
		return
	}

	err := s.App.Store().Q.UpdateCocktail(db.UpdateCocktailParams{
		ID:              id,
		Name:            c.Name,
		Description:     c.Description,
		ImagePath:       imagePath,
		Tags:            c.Tags,
		Difficulty:      c.Difficulty,
		PrepTimeMinutes: c.PrepTimeMinutes,
		Instructions:    c.Instructions,
		IsEnabled:       c.IsEnabled,
	})
	if err != nil {
		s.App.AddFlash(w, r, app.FlashError, "Update failed (name might already exist).")
		s.redirect(w, r, "/bartender/cocktails/"+idStr+"/edit")
		return
	}

	_ = s.App.Store().Q.ReplaceCocktailIngredients(id, items)
	s.broadcastInventory()
	s.App.AddFlash(w, r, app.FlashSuccess, "Cocktail updated.")
	s.redirect(w, r, "/bartender/cocktails")
}

func (s *Server) CocktailTogglePost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/bartender/cocktails")
		return
	}
	_ = r.ParseForm()
	enabled := formBool(r, "is_enabled")
	_ = s.App.Store().Q.ToggleCocktailEnabled(id, enabled)
	s.broadcastInventory()
	s.redirect(w, r, "/bartender/cocktails")
}

func (s *Server) CocktailDeletePost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/bartender/cocktails")
		return
	}
	_ = s.App.Store().Q.DeleteCocktail(id)
	s.broadcastInventory()
	s.App.AddFlash(w, r, app.FlashSuccess, "Cocktail deleted.")
	s.redirect(w, r, "/bartender/cocktails")
}

func (s *Server) parseCocktailForm(w http.ResponseWriter, r *http.Request, id int64, existingImage string) (db.Cocktail, []db.IngredientUpsertItem, string, bool) {
	name := strings.TrimSpace(r.FormValue("name"))
	desc := strings.TrimSpace(r.FormValue("description"))
	tags := strings.TrimSpace(r.FormValue("tags"))
	diff := strings.TrimSpace(r.FormValue("difficulty"))
	if diff == "" {
		diff = "easy"
	}
	prep := int64(5)
	if v := strings.TrimSpace(r.FormValue("prep_time_minutes")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			prep = n
		}
	}
	instr := strings.TrimSpace(r.FormValue("instructions"))
	enabled := formBool(r, "is_enabled")

	if name == "" {
		s.App.AddFlash(w, r, app.FlashError, "Name is required.")
		return db.Cocktail{}, nil, existingImage, false
	}

	// Ingredients arrays
	pids := r.Form["ingredient_product_id"]
	qtys := r.Form["ingredient_quantity"]
	units := r.Form["ingredient_unit"]
	reqs := r.Form["ingredient_required"]

	var items []db.IngredientUpsertItem
	n := len(pids)
	for i := 0; i < n; i++ {
		pid, ok := parseInt64(pids[i])
		if !ok {
			continue
		}
		var q *float64
		if i < len(qtys) {
			qs := strings.TrimSpace(qtys[i])
			if qs != "" {
				if f, err := strconv.ParseFloat(qs, 64); err == nil {
					q = &f
				}
			}
		}
		unit := ""
		if i < len(units) {
			unit = strings.TrimSpace(units[i])
		}
		required := true
		if i < len(reqs) {
			required = strings.TrimSpace(reqs[i]) != "0"
		}
		items = append(items, db.IngredientUpsertItem{
			ProductID: pid,
			Quantity:  q,
			Unit:      unit,
			Required:  required,
		})
	}

	// Image upload (optional)
	imagePath := existingImage
	if file, hdr, err := r.FormFile("image"); err == nil && file != nil {
		defer file.Close()
		if p, err := s.saveUpload(hdr.Filename, file); err == nil {
			imagePath = p
		}
	}

	c := db.Cocktail{
		ID:              id,
		Name:            name,
		Description:     desc,
		Tags:            tags,
		Difficulty:      diff,
		PrepTimeMinutes: prep,
		Instructions:    instr,
		IsEnabled:       enabled,
	}
	return c, items, imagePath, true
}

func (s *Server) saveUpload(original string, src io.Reader) (string, error) {
	ext := strings.ToLower(filepath.Ext(original))
	if ext == "" {
		ext = ".img"
	}
	name := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "cocktail", ext)
	dstPath := filepath.Join(s.App.Config().UploadDir, name)

	if err := os.MkdirAll(s.App.Config().UploadDir, 0o755); err != nil {
		return "", err
	}

	f, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, src)
	if err != nil {
		return "", err
	}
	return "/uploads/" + name, nil
}

func (s *Server) broadcastInventory() {
	s.App.SSE().BroadcastInventory(app.SSEEvent{
		Type: "inventory:updated",
		Data: map[string]any{"ts": time.Now().Unix()},
	})
}
