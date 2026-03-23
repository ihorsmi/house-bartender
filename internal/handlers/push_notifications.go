package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"house-bartender-go/internal/app"
	pushsvc "house-bartender-go/internal/services/push"
)

type pushSubscriptionRequest struct {
	Endpoint    string `json:"endpoint"`
	DeviceLabel string `json:"device_label"`
	Keys        struct {
		P256DH string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

type pushUnsubscribeRequest struct {
	Endpoint string `json:"endpoint"`
}

type pushAPIResponse struct {
	OK    bool   `json:"ok"`
	State string `json:"state,omitempty"`
	Error string `json:"error,omitempty"`
}

func (s *Server) PushSubscribePost(w http.ResponseWriter, r *http.Request) {
	u := s.App.CurrentUser(r)
	if u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if u.Role != app.RoleBartender {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if !s.requireTrustedOrigin(w, r) {
		return
	}

	var req pushSubscriptionRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, pushAPIResponse{OK: false, Error: "Invalid notification payload."})
		return
	}

	err := s.App.Push().SaveSubscription(u.ID, truncateString(r.UserAgent(), 512), pushsvc.SubscriptionInput{
		Endpoint:    req.Endpoint,
		P256DH:      req.Keys.P256DH,
		Auth:        req.Keys.Auth,
		DeviceLabel: req.DeviceLabel,
	})
	if err != nil {
		switch {
		case errors.Is(err, pushsvc.ErrNotConfigured):
			writeJSON(w, http.StatusServiceUnavailable, pushAPIResponse{OK: false, Error: "Push notifications are not configured on this server."})
		case errors.Is(err, pushsvc.ErrInvalidSubscription):
			writeJSON(w, http.StatusBadRequest, pushAPIResponse{OK: false, Error: "This browser subscription is missing required fields."})
		default:
			writeJSON(w, http.StatusInternalServerError, pushAPIResponse{OK: false, Error: "Could not save notification settings."})
		}
		return
	}

	writeJSON(w, http.StatusOK, pushAPIResponse{OK: true, State: "enabled"})
}

func (s *Server) PushUnsubscribePost(w http.ResponseWriter, r *http.Request) {
	u := s.App.CurrentUser(r)
	if u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if u.Role != app.RoleBartender {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if !s.requireTrustedOrigin(w, r) {
		return
	}

	var req pushUnsubscribeRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, pushAPIResponse{OK: false, Error: "Invalid notification payload."})
		return
	}

	if err := s.App.Push().DisableSubscription(u.ID, req.Endpoint); err != nil {
		if errors.Is(err, pushsvc.ErrInvalidSubscription) {
			writeJSON(w, http.StatusBadRequest, pushAPIResponse{OK: false, Error: "Subscription endpoint is required."})
			return
		}
		writeJSON(w, http.StatusInternalServerError, pushAPIResponse{OK: false, Error: "Could not disable notifications for this device."})
		return
	}

	writeJSON(w, http.StatusOK, pushAPIResponse{OK: true, State: "disabled"})
}

func decodeJSONBody(r *http.Request, dst any) error {
	defer r.Body.Close()

	dec := json.NewDecoder(io.LimitReader(r.Body, 32<<10))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(new(struct{})); err != io.EOF {
		return errors.New("unexpected extra JSON data")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func truncateString(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func (s *Server) requireTrustedOrigin(w http.ResponseWriter, r *http.Request) bool {
	allowed := map[string]struct{}{}

	if base := strings.TrimSpace(s.App.Config().BaseURL); base != "" {
		if parsed, err := url.Parse(base); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			allowed[parsed.Scheme+"://"+parsed.Host] = struct{}{}
		}
	}
	if origin := requestOrigin(r); origin != "" {
		allowed[origin] = struct{}{}
	}

	check := func(raw string) bool {
		if strings.TrimSpace(raw) == "" {
			return false
		}
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return false
		}
		_, ok := allowed[parsed.Scheme+"://"+parsed.Host]
		return ok
	}

	if check(r.Header.Get("Origin")) || check(r.Referer()) {
		return true
	}

	http.Error(w, "forbidden", http.StatusForbidden)
	return false
}

func requestOrigin(r *http.Request) string {
	scheme := "http"
	if proto := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0]); proto != "" {
		scheme = proto
	} else if r.TLS != nil {
		scheme = "https"
	}
	if r.Host == "" {
		return ""
	}
	return scheme + "://" + r.Host
}
