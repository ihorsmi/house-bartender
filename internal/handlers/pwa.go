package handlers

import (
	"net/http"
	"path/filepath"
)

func (s *Server) ManifestGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/manifest+json")
	http.ServeFile(w, r, filepath.Join("static", "manifest.webmanifest"))
}

func (s *Server) ServiceWorkerGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Service-Worker-Allowed", "/")
	http.ServeFile(w, r, filepath.Join("static", "sw.js"))
}
