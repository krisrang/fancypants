package handlers

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/krisrang/fancypants/Godeps/_workspace/src/github.com/dustin/go-humanize"
)

const (
	tmpl = `
{{$p := .Path}}
<html>
<head>
	<meta http-equiv="content-type" content="text/html;charset=utf-8">
	<meta name="viewport" content="width=device-width" />
    <meta http-equiv="X-UA-Compatible" content="IE=Edge">
	<title>Files</title>
	<link rel="stylesheet" href="/autoindexassets/bootstrap.min.css">
	<link rel="stylesheet" href="/autoindexassets/bootstrap-theme.min.css">
	<link rel="stylesheet" href="/autoindexassets/autoindex.css?{{.Time}}">
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
		<ol class="breadcrumb">
			{{range .Breadcrumbs}}
				{{if .Path}}
					<li><a href="{{.Path}}">{{.Name}}</a></li>
				{{else}}
					<li class="active">{{.Name}}</li>
				{{end}}
			{{end}}
		</ol>
		{{range .Files}}
			<div class="row filerow">
				{{if .IsDir}}
					<a href="{{$p}}/{{.Name}}" class="col-xs-11">{{.Name}}/</a>
					<div class="col-xs-1 text-right">-</div>
				{{else}}
					<div class="col-md-9 col-sm-8 col-xs-9 filename">
						<a href="{{$p}}/{{.Name}}">{{.Name}}</a>
					</div>
					<div class="col-md-1 col-sm-1 col-xs-3 text-right">{{sizeHuman .Size}}</div>
					<div class="col-md-2 col-sm-3 hidden-xs text-right">{{timeHuman .ModTime}}</div>
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

	// timeHuman := func(time time.Time) string {
	// 	return time.Format("02 Jan 2006 15:04")
	// }

	funcMap := template.FuncMap{
		"sizeHuman": sizeHuman,
		"timeHuman": humanize.Time,
	}

	t = template.Must(template.New("template").Funcs(funcMap).Parse(tmpl))
}

type BreadCrumb struct {
	Name string
	Path string
}

type DirInfo struct {
	Files       []os.FileInfo
	Breadcrumbs []BreadCrumb
	Path        string
	Time        int64
}

type ByNameCaseInsensitive []os.FileInfo

func (a ByNameCaseInsensitive) Len() int      { return len(a) }
func (a ByNameCaseInsensitive) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByNameCaseInsensitive) Less(i, j int) bool {
	return strings.ToLower(a[i].Name()) < strings.ToLower(a[j].Name())
}

func serveIndex(w http.ResponseWriter, r *http.Request, path string) {
	urlPath := r.URL.Path
	if urlPath == "/" {
		urlPath = ""
	} else if string(urlPath[len(urlPath)-1]) == "/" {
		urlPath = urlPath[0 : len(urlPath)-1]
	}

	results, err := ioutil.ReadDir(path)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	// Sort files list by names non case sensitive
	sort.Sort(ByNameCaseInsensitive(results))

	// Build breadcrumbs
	crumbs := make([]BreadCrumb, 0)
	previousPath := ""
	bits := strings.Split(r.URL.Path, "/")

	for _, crumb := range bits[0 : len(bits)-1] {
		path := previousPath + "/" + crumb
		name := crumb

		previousPath = path

		if crumb == "" {
			name = "Home"
			previousPath = ""
		}

		crumbs = append(crumbs, BreadCrumb{
			Name: name,
			Path: path,
		})
	}

	crumbs = append(crumbs, BreadCrumb{
		Name: bits[len(bits)-1],
	})

	// List directories before files
	files := make([]os.FileInfo, 0)
	dirs := make([]os.FileInfo, 0)

	for _, file := range results {
		// Ignore dotfiles
		if string(file.Name()[0]) == "." {
			continue
		}

		if file.IsDir() {
			dirs = append(dirs, file)
		} else {
			files = append(files, file)
		}
	}

	info := DirInfo{
		Files:       append(dirs, files...),
		Breadcrumbs: crumbs,
		Path:        urlPath,
		Time:        time.Now().Unix(),
	}

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

			// If index.html exists, serve that
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
				return
			}

			serveIndex(w, r, fullPath)
			return
		}

		// If file is above 100MB use x-accel-redirect to have nginx serve it directly
		if err == nil && stat.Size() > 100*1024*1024 {
			w.Header().Set("X-Accel-Redirect", "/kris.moe"+r.URL.Path)
			w.Header().Set("Content-Disposition", "attachment")
		} else {
			http.ServeFile(w, r, fullPath)
		}
	})
}
