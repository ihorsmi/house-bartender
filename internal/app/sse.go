package app

import (
	"log/slog"
	"sync"
)

type SSEEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type SSEHub struct {
	log *slog.Logger

	mu    sync.RWMutex
	subs  map[string]map[chan SSEEvent]struct{} // topic -> set(ch)
	alive bool
}

func NewSSEHub(logger *slog.Logger) *SSEHub {
	if logger == nil {
		logger = slog.Default()
	}
	return &SSEHub{
		log:   logger,
		subs: map[string]map[chan SSEEvent]struct{}{},
		alive: true,
	}
}

func (h *SSEHub) Subscribe(topics []string, buf int) (<-chan SSEEvent, func()) {
	if buf <= 0 {
		buf = 16
	}
	ch := make(chan SSEEvent, buf)

	h.mu.Lock()
	for _, t := range topics {
		if h.subs[t] == nil {
			h.subs[t] = map[chan SSEEvent]struct{}{}
		}
		h.subs[t][ch] = struct{}{}
	}
	h.mu.Unlock()

	cancel := func() {
		h.mu.Lock()
		for _, t := range topics {
			if set, ok := h.subs[t]; ok {
				delete(set, ch)
				if len(set) == 0 {
					delete(h.subs, t)
				}
			}
		}
		h.mu.Unlock()
		close(ch)
	}
	return ch, cancel
}

func (h *SSEHub) Broadcast(topic string, ev SSEEvent) {
	h.mu.RLock()
	set := h.subs[topic]
	h.mu.RUnlock()

	for ch := range set {
		select {
		case ch <- ev:
		default:
			// drop if slow consumer
		}
	}
}

/* ---- topic helpers ---- */

func TopicUser(userID int64) string { return "user:" + itoa64(userID) }
func TopicRole(role string) string  { return "role:" + role }
func TopicOrdersGlobal() string     { return "orders:global" }
func TopicInventory() string        { return "inventory:global" }

func (h *SSEHub) BroadcastUser(userID int64, ev SSEEvent)      { h.Broadcast(TopicUser(userID), ev) }
func (h *SSEHub) BroadcastRole(role string, ev SSEEvent)       { h.Broadcast(TopicRole(role), ev) }
func (h *SSEHub) BroadcastOrders(ev SSEEvent)                  { h.Broadcast(TopicOrdersGlobal(), ev) }
func (h *SSEHub) BroadcastInventory(ev SSEEvent)               { h.Broadcast(TopicInventory(), ev) }

// small helper to avoid importing handlers for itoa64
func itoa64(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [32]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + (v % 10))
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
