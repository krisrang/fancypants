package handlers

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	tmpl = `
{{$p := .Path}}
<html>
<head>
	<title>Files</title>
	<link rel="stylesheet" href="/autoindexassets/bootstrap.min.css">
	<link rel="stylesheet" href="/autoindexassets/bootstrap-theme.min.css">
</head>
<body>
	<h1>Files</h1>
	<ul>
	{{range .Files}}
		<li><a href="{{urlquery $p}}/{{urlquery .Name}}">{{.Name}}</a></li>
	{{end}}
	</ul>
</body>
</html>
`
)

var (
	t = template.Must(template.New("template").Parse(tmpl))
)

type dirInfo struct {
	Files []os.FileInfo
	Path  string
}

func serveIndex(w http.ResponseWriter, r *http.Request, path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	urlPath := r.URL.Path
	if urlPath == "/" {
		urlPath = ""
	}

	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, dirInfo{files, urlPath}); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func Autodir(path string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullPath := filepath.Join(path, strings.TrimPrefix(r.URL.Path, "/"))

		if stat, err := os.Stat(fullPath); err == nil && stat.Mode().IsDir() {
			indexPath := filepath.Join(fullPath, "index.html")

			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
				return
			}

			serveIndex(w, r, fullPath)
			return
		}

		http.ServeFile(w, r, fullPath)
	})
}
