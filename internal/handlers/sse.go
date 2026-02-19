package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"house-bartender-go/internal/app"
)

func (s *Server) SSEGet(w http.ResponseWriter, r *http.Request) {
	u := s.App.CurrentUser(r)
	if u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Topics
	topics := []string{
		app.TopicUser(u.ID),
		app.TopicInventory(),
	}

	// Bartenders/Admin get global order events + role events
	if u.Role == app.RoleBartender || u.Role == app.RoleAdmin {
		topics = append(topics, app.TopicOrdersGlobal(), app.TopicRole(u.Role))
	}

	ch, cancel := s.App.SSE().Subscribe(topics, 32)
	defer cancel()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// initial hello
	hello, _ := json.Marshal(map[string]any{"ok": true, "ts": time.Now().Unix()})
	fmt.Fprintf(w, "event: hello\ndata: %s\n\n", hello)
	flusher.Flush()

	keep := time.NewTicker(25 * time.Second)
	defer keep.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-keep.C:
			// comment ping
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		case ev, ok := <-ch:
			if !ok {
				return
			}
			b, _ := json.Marshal(ev.Data)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, b)
			flusher.Flush()
		}
	}
}
