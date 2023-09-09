package main

import (
	"fmt"
	"gingersnap/app"
	"gingersnap/ui"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	fmt.Println("Test cli for gingersnap")

	// Construct logger.
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

	// Construct the templates, using the embedded FS.
	templates, err := app.NewTemplate(ui.Templates)
	if err != nil {
		logger.Fatal(err)
	}

	// Construct the site info
	site := &app.Site{
		Name:        "Gingersnap",
		Host:        "gingersnap.dev",
		Tagline:     "The snappy starter project with Go and SQLite!",
		Description: "The snappy starter project, built with Go and SQLite. Get up and running with only one command. Dockerize and deploy when you're ready to ship!",
	}
	site.Title = fmt.Sprintf("%s - %s", site.Name, site.Tagline)
	site.Host = fmt.Sprintf("localhost%s", ":4000")
	site.Url = fmt.Sprintf("http://%s", site.Host)
	site.Email = fmt.Sprintf("admin@%s", site.Host)
	site.Image = app.Image{
		Url:    "/static/meta-img.webp",
		Alt:    "some img alt here",
		Type:   "webp",
		Width:  "800",
		Height: "450",
	}

	// Construct the main Gingersnap engine.
	g := &app.Gingersnap{
		Debug:      true,
		Logger:     logger,
		Static:     ui.Static,
		Templates:  templates,
		ListenAddr: ":4000",
		Site:       site,
	}

	// Construct Http Server.
	g.HttpServer = &http.Server{
		Addr:         g.ListenAddr,
		Handler:      g.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	g.RunServer()
}
