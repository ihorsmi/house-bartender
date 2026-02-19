package app

import (
	"net/http"
	"strings"
	"time"
)

type FlashLevel string

const (
	FlashInfo    FlashLevel = "info"
	FlashSuccess FlashLevel = "success"
	FlashError   FlashLevel = "error"
)

type Flash struct {
	Level   FlashLevel `json:"level"`
	Message string     `json:"message"`
}

type flashPayload struct {
	Exp   int64   `json:"exp"`
	Items []Flash `json:"items"`
}

func (a *App) AddFlash(w http.ResponseWriter, r *http.Request, lvl FlashLevel, msg string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return
	}

	fp := flashPayload{Exp: time.Now().Add(10 * time.Minute).Unix()}
	_ = a.readFlash(r, &fp)
	fp.Items = append(fp.Items, Flash{Level: lvl, Message: msg})

	val, err := a.signJSON(fp)
	if err != nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     flashCookieName,
		Value:    val,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.HasPrefix(strings.ToLower(a.cfg.BaseURL), "https://"),
		Expires:  time.Unix(fp.Exp, 0),
	})
}

func (a *App) PopFlashes(w http.ResponseWriter, r *http.Request) []Flash {
	var fp flashPayload
	if err := a.readFlash(r, &fp); err != nil {
		return nil
	}
	if fp.Exp > 0 && time.Now().Unix() > fp.Exp {
		a.clearFlash(w)
		return nil
	}
	out := fp.Items
	a.clearFlash(w)
	return out
}

func (a *App) readFlash(r *http.Request, out *flashPayload) error {
	c, err := r.Cookie(flashCookieName)
	if err != nil || c.Value == "" {
		return err
	}
	return a.verifyJSON(c.Value, out)
}

func (a *App) clearFlash(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.HasPrefix(strings.ToLower(a.cfg.BaseURL), "https://"),
	})
}
