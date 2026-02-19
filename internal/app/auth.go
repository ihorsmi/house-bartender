package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sessionCookieName = "hb_session"
const flashCookieName = "hb_flash"

type sessionPayload struct {
	UID   int64  `json:"uid"`
	Exp   int64  `json:"exp"`
	Nonce string `json:"nonce"`
}

func NormalizeEmail(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	return s
}

func HashPassword(pw string) (string, error) {
	pw = strings.TrimSpace(pw)
	if len(pw) < 8 {
		return "", errors.New("password too short")
	}
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CheckPassword(hash string, pw string) bool {
	if hash == "" || pw == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// SetSessionUser sets a signed cookie with user id.
func (a *App) SetSessionUser(w http.ResponseWriter, r *http.Request, userID int64) error {
	now := time.Now()
	pl := sessionPayload{
		UID:   userID,
		Exp:   now.Add(14 * 24 * time.Hour).Unix(),
		Nonce: randomNonce(),
	}
	val, err := a.signJSON(pl)
	if err != nil {
		return err
	}

	c := &http.Cookie{
		Name:     sessionCookieName,
		Value:    val,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.HasPrefix(strings.ToLower(a.cfg.BaseURL), "https://"),
		Expires:  time.Unix(pl.Exp, 0),
	}
	http.SetCookie(w, c)
	return nil
}

func (a *App) ClearSession(w http.ResponseWriter, r *http.Request) error {
	c := &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.HasPrefix(strings.ToLower(a.cfg.BaseURL), "https://"),
	}
	http.SetCookie(w, c)
	return nil
}

func (a *App) GetSessionUserID(r *http.Request) (int64, bool) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil || c.Value == "" {
		return 0, false
	}
	var pl sessionPayload
	if err := a.verifyJSON(c.Value, &pl); err != nil {
		return 0, false
	}
	if pl.UID <= 0 || pl.Exp <= 0 || time.Now().Unix() > pl.Exp {
		return 0, false
	}
	return pl.UID, true
}

/* ---------- signed cookie helpers (used by flash.go too) ---------- */

func (a *App) signJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(b)
	sig := a.sign(payload)
	return payload + "." + sig, nil
}

func (a *App) verifyJSON(s string, out any) error {
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return errors.New("bad format")
	}
	payload, sig := parts[0], parts[1]
	if !a.verify(payload, sig) {
		return errors.New("bad signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}

func (a *App) sign(payload string) string {
	m := hmac.New(sha256.New, a.cfg.SessionHashKey)
	_, _ = m.Write([]byte(payload))
	return hex.EncodeToString(m.Sum(nil))
}

func (a *App) verify(payload, sigHex string) bool {
	got, err := hex.DecodeString(sigHex)
	if err != nil {
		return false
	}
	m := hmac.New(sha256.New, a.cfg.SessionHashKey)
	_, _ = m.Write([]byte(payload))
	want := m.Sum(nil)
	return hmac.Equal(got, want)
}

func randomNonce() string {
	var b [16]byte
	_, _ = time.Now().UTC().MarshalBinary() // ignore
	// lightweight nonce without importing crypto/rand here (ephemeral ok); flash/session integrity comes from HMAC
	t := time.Now().UnixNano()
	for i := 0; i < len(b); i++ {
		b[i] = byte(t >> (uint(i) * 3))
	}
	return base64.RawURLEncoding.EncodeToString(b[:])
}
