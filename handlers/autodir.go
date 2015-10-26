package handlers

import (
	// "fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Autodir(path string, h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullPath := filepath.Join(path, strings.TrimPrefix(r.URL.Path, "/"))

		if strings.HasSuffix(r.URL.Path, "/") {
			indexPath := filepath.Join(fullPath, "index.html")

			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
				return
			}
		}

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, fullPath)
	})
}
