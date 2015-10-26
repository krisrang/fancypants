package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Autodir(path string, h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			fullPath := filepath.Join(path, strings.TrimPrefix(r.URL.Path, "/"))
			indexPath := filepath.Join(fullPath, "index.html")

			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
				return
			}

			http.NotFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}
