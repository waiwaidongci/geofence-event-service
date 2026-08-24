package httpapi

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/*
var consoleAssets embed.FS

func (s *Server) registerConsoleRoutes() {
	assets, err := fs.Sub(consoleAssets, "web")
	if err != nil {
		panic(err)
	}
	files := http.FileServer(http.FS(assets))
	s.mux.Handle("GET /console/", http.StripPrefix("/console/", files))
	s.mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/console/", http.StatusTemporaryRedirect)
	})
}
