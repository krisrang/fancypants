package handlers

import (
	// "fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/krisrang/fancypants/Godeps/_workspace/src/github.com/dustin/go-humanize"
)

const (
	tmpl = `
{{$p := .Path}}
<html>
<head>
	<title>Files</title>
	<link rel="stylesheet" href="/autoindexassets/bootstrap.min.css">
	<link rel="stylesheet" href="/autoindexassets/bootstrap-theme.min.css">
	<link rel="stylesheet" href="/autoindexassets/autoindex.css">
</head>
<body>
	<div class="container">
		<nav class="navbar navbar-default">
		  <div class="container-fluid">
		    <div class="navbar-header">
		      <a class="navbar-brand" href="/">Files</a>
		    </div>
		  </div>
		</nav>
		{{if .Up}}
		<div class="row filerow">
			<a href="{{.Up}}" class="col-lg-12">..</a>
		</div>
		{{end}}
		{{range .Files}}
			<div class="row filerow">
				{{if .IsDir}}
					<a href="{{$p}}/{{.Name}}" class="col-lg-10">{{.Name}}/</a>
					<div class="col-lg-2">-</div>
				{{else}}
					<a href="{{$p}}/{{.Name}}" class="col-lg-10">{{.Name}}</a>
					<div class="col-lg-2">{{sizeHuman .Size}}</div>
				{{end}}
			</div>
		{{end}}
	</div>
</body>
</html>
`
)

var (
	t *template.Template
)

func init() {
	sizeHuman := func(size int64) string {
		return humanize.Bytes(uint64(size))
	}

	funcMap := template.FuncMap{
		"sizeHuman": sizeHuman,
	}

	t = template.Must(template.New("template").Funcs(funcMap).Parse(tmpl))
}

type DirInfo struct {
	Files []os.FileInfo
	Path  string
	Up    string
}

func serveIndex(w http.ResponseWriter, r *http.Request, path string) {
	urlPath := r.URL.Path
	if urlPath == "/" {
		urlPath = ""
	}

	results, err := ioutil.ReadDir(path)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	up := ""
	if urlPath != "" {
		index := strings.LastIndex(urlPath, "/")
		if index < 0 {
			index = 0
		}

		if index == 0 {
			up = "/"
		} else {
			up = urlPath[:index]
		}
	}

	info := DirInfo{make([]os.FileInfo, 0), urlPath, up}
	files := make([]os.FileInfo, 0)
	dirs := make([]os.FileInfo, 0)

	for _, file := range results {
		if file.IsDir() {
			dirs = append(dirs, file)
		} else {
			files = append(files, file)
		}
	}

	info.Files = append(dirs, files...)

	w.Header().Set("Content-Type", "text/html")

	if err := t.Execute(w, info); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func Autodir(path string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullPath := filepath.Join(path, strings.TrimPrefix(r.URL.Path, "/"))
		stat, err := os.Stat(fullPath)

		if err == nil && stat.Mode().IsDir() {
			indexPath := filepath.Join(fullPath, "index.html")

			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
				return
			}

			serveIndex(w, r, fullPath)
			return
		}

		if err == nil && stat.Size() > 10*1024*1024 {
			w.Header().Set("X-Accel-Redirect", "/moe"+r.URL.Path)
		} else {
			http.ServeFile(w, r, fullPath)
		}
	})
}
