package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/krisrang/fancypants/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/krisrang/fancypants/Godeps/_workspace/src/github.com/elazarl/go-bindata-assetfs"
	"github.com/krisrang/fancypants/Godeps/_workspace/src/github.com/thoas/stats"
	"github.com/krisrang/fancypants/handlers"
)

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	app := cli.NewApp()
	app.Name = "fancypants"
	app.Usage = "http file server with fancy autoindex"
	app.Version = "0.0.1"
	app.Action = serve
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "path",
			Value: "./",
			Usage: "path to serve",
		},
		cli.StringFlag{
			Name:  "port",
			Value: "8080",
			Usage: "port to serve on",
		},
	}
	app.Run(os.Args)
}

func serve(c *cli.Context) {
	port := c.GlobalString("port")

	fullPath, err := filepath.Abs(c.GlobalString("path"))
	if err != nil {
		log.Fatalf("Error setting path: %v", err)
	}

	logger := log.New(os.Stderr, "", log.Flags())
	stat := stats.New()
	mux := http.DefaultServeMux

	mux.HandleFunc("/stats", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		bytes, err := json.Marshal(stat.Data())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write(bytes)
	}))

	http.Handle("/autoindexassets/",
		stat.Handler(
			http.StripPrefix("/autoindexassets/",
				http.FileServer(
					&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "autoindexassets"}))))

	http.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stat.ServeHTTP(w, r, handlers.Autodir(fullPath))
	}))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handlers.NewApacheLoggingHandler(mux, logger),
	}

	log.Printf("Listening on %v", port)
	server.ListenAndServe()
}
