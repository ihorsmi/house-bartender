package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"house-bartender-go/internal/app"
	"house-bartender-go/internal/db"

	"github.com/go-chi/chi/v5"
)

func (s *Server) OrderCreatePost(w http.ResponseWriter, r *http.Request) {
	u := s.App.CurrentUser(r)
	if u == nil {
		s.redirect(w, r, "/login")
		return
	}
	_ = r.ParseForm()

	cidStr := strings.TrimSpace(r.FormValue("cocktail_id"))
	cid, ok := parseInt64(cidStr)
	if !ok {
		s.App.AddFlash(w, r, app.FlashError, "Invalid cocktail.")
		s.redirect(w, r, "/")
		return
	}

	// Must be available at time of order
	c, _ := s.App.Store().Q.GetCocktailByID(cid)
	if c == nil || !c.IsEnabled {
		s.App.AddFlash(w, r, app.FlashError, "Cocktail not available.")
		s.redirect(w, r, "/")
		return
	}
	ings, _ := s.App.Store().Q.GetCocktailIngredients(cid)
	for _, it := range ings {
		if it.Required && !it.ProductAvail {
			s.App.AddFlash(w, r, app.FlashError, "Cocktail not available (missing ingredients).")
			s.redirect(w, r, "/cocktails/"+cidStr)
			return
		}
	}

	qty := int64(1)
	if q := strings.TrimSpace(r.FormValue("quantity")); q != "" {
		if n, err := strconv.ParseInt(q, 10, 64); err == nil && n > 0 && n <= 10 {
			qty = n
		}
	}

	location := strings.TrimSpace(r.FormValue("location"))
	notes := strings.TrimSpace(r.FormValue("notes"))

	if location == "" {
		s.App.AddFlash(w, r, app.FlashError, "Location is required.")
		s.redirect(w, r, "/cocktails/"+cidStr)
		return
	}

	oid, err := s.App.Store().Q.CreateOrder(db.CreateOrderParams{
		UserID:     u.ID,
		CocktailID: cid,
		Quantity:   qty,
		Notes:      notes,
		Location:   location,
	})
	if err != nil {
		s.App.AddFlash(w, r, app.FlashError, "Could not create order.")
		s.redirect(w, r, "/cocktails/"+cidStr)
		return
	}

	// SSE: order:created -> bartenders/admin
	s.App.SSE().BroadcastRole(app.RoleBartender, app.SSEEvent{Type: "order:created", Data: map[string]any{"order_id": oid}})
	s.App.SSE().BroadcastRole(app.RoleAdmin, app.SSEEvent{Type: "order:created", Data: map[string]any{"order_id": oid}})
	s.App.SSE().BroadcastOrders(app.SSEEvent{Type: "order:created", Data: map[string]any{"order_id": oid}})

	s.App.AddFlash(w, r, app.FlashSuccess, "Order placed.")
	s.redirect(w, r, "/orders")
}

func (s *Server) OrderAcceptPost(w http.ResponseWriter, r *http.Request) {
	u := s.App.CurrentUser(r)
	if u == nil {
		s.redirect(w, r, "/login")
		return
	}

	idStr := chi.URLParam(r, "id")
	oid, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/bartender/orders")
		return
	}

	o, _ := s.App.Store().Q.GetOrderByID(oid)
	if o == nil {
		s.redirect(w, r, "/bartender/orders")
		return
	}
	if o.Status != "PLACED" {
		s.redirect(w, r, "/bartender/orders")
		return
	}

	// assign to self if not assigned
	if o.AssignedBartenderID == nil {
		_ = s.App.Store().Q.AssignOrder(oid, &u.ID)
	}

	_ = s.App.Store().Q.UpdateOrderStatus(oid, "PLACED", "ACCEPTED", &u.ID)
	s.broadcastOrderUpdated(oid)
	s.redirect(w, r, "/bartender/orders")
}

func (s *Server) OrderAssignPost(w http.ResponseWriter, r *http.Request) {
	u := s.App.CurrentUser(r)
	if u == nil {
		s.redirect(w, r, "/login")
		return
	}

	idStr := chi.URLParam(r, "id")
	oid, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/bartender/orders")
		return
	}

	_ = r.ParseForm()
	bidStr := strings.TrimSpace(r.FormValue("bartender_id"))
	var bid *int64
	if bidStr == "" {
		bid = &u.ID
	} else {
		b, ok := parseInt64(bidStr)
		if ok {
			bid = &b
		}
	}

	_ = s.App.Store().Q.AssignOrder(oid, bid)
	s.broadcastOrderUpdated(oid)
	s.redirect(w, r, "/bartender/orders")
}

func (s *Server) OrderStatusPost(w http.ResponseWriter, r *http.Request) {
	u := s.App.CurrentUser(r)
	if u == nil {
		s.redirect(w, r, "/login")
		return
	}
	idStr := chi.URLParam(r, "id")
	oid, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/bartender/orders")
		return
	}

	_ = r.ParseForm()
	to := strings.TrimSpace(r.FormValue("to_status"))
	if to == "" {
		s.redirect(w, r, "/bartender/orders")
		return
	}

	o, _ := s.App.Store().Q.GetOrderByID(oid)
	if o == nil {
		s.redirect(w, r, "/bartender/orders")
		return
	}

	from := o.Status
	if !allowedTransition(from, to) {
		s.App.AddFlash(w, r, app.FlashError, "Invalid status transition.")
		s.redirect(w, r, "/bartender/orders")
		return
	}

	// auto-assign to self if none
	if o.AssignedBartenderID == nil {
		_ = s.App.Store().Q.AssignOrder(oid, &u.ID)
	}

	_ = s.App.Store().Q.UpdateOrderStatus(oid, from, to, &u.ID)
	s.broadcastOrderUpdated(oid)
	s.redirect(w, r, "/bartender/orders")
}

func (s *Server) OrderCancelPost(w http.ResponseWriter, r *http.Request) {
	u := s.App.CurrentUser(r)
	if u == nil {
		s.redirect(w, r, "/login")
		return
	}
	idStr := chi.URLParam(r, "id")
	oid, ok := parseInt64(idStr)
	if !ok {
		s.redirect(w, r, "/bartender/orders")
		return
	}

	o, _ := s.App.Store().Q.GetOrderByID(oid)
	if o == nil {
		s.redirect(w, r, "/bartender/orders")
		return
	}
	if o.Status == "DELIVERED" || o.Status == "CANCELLED" {
		s.redirect(w, r, "/bartender/orders")
		return
	}

	_ = s.App.Store().Q.UpdateOrderStatus(oid, o.Status, "CANCELLED", &u.ID)
	s.broadcastOrderUpdated(oid)
	s.redirect(w, r, "/bartender/orders")
}

func (s *Server) broadcastOrderUpdated(orderID int64) {
	o, _ := s.App.Store().Q.GetOrderByID(orderID)
	if o == nil {
		return
	}
	ev := app.SSEEvent{Type: "order:updated", Data: map[string]any{"order_id": orderID, "status": o.Status}}

	// to owner + bartenders/admin
	s.App.SSE().BroadcastUser(o.UserID, ev)
	s.App.SSE().BroadcastRole(app.RoleBartender, ev)
	s.App.SSE().BroadcastRole(app.RoleAdmin, ev)
	s.App.SSE().BroadcastOrders(ev)
}

func allowedTransition(from, to string) bool {
	switch from {
	case "PLACED":
		return to == "ACCEPTED" || to == "CANCELLED"
	case "ACCEPTED":
		return to == "IN_PROGRESS" || to == "CANCELLED"
	case "IN_PROGRESS":
		return to == "READY" || to == "CANCELLED"
	case "READY":
		return to == "DELIVERED" || to == "CANCELLED"
	default:
		return false
	}
}
