package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/krisrang/fancypants/Godeps/_workspace/src/github.com/codegangsta/cli"
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

	mux := http.DefaultServeMux
	mux.HandleFunc("/", handlers.Autodir(fullPath, http.FileServer(http.Dir(fullPath))))

	loggingHandler := handlers.NewApacheLoggingHandler(mux, logger)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: loggingHandler,
	}

	log.Printf("Listening on %v", port)
	server.ListenAndServe()
}
