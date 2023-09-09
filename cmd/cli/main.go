package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"gingersnap/app"
)

func main() {
	fmt.Println("Test cli for gingersnap")

	// Construct logger.
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

	// Construct the templates, using the embedded FS.
	templates, err := app.NewTemplate(app.Templates)
	if err != nil {
		logger.Fatal(err)
	}

	// Construct the models.
	posts := &app.PostModel{}
	categories := &app.CategoryModel{}

	// Construct the site info
	config := &app.Config{
		SiteName:        "Gingersnap",
		SiteHost:        "gingersnap.dev",
		SiteTagline:     "The snappy starter project with Go and SQLite!",
		SiteDescription: "The snappy starter project, built with Go and SQLite. Get up and running with only one command. Dockerize and deploy when you're ready to ship!",
	}
	config.SiteTitle = fmt.Sprintf("%s - %s", config.SiteName, config.SiteTagline)
	config.SiteHost = fmt.Sprintf("localhost%s", ":4000")
	config.SiteUrl = fmt.Sprintf("http://%s", config.SiteHost)
	config.SiteEmail = fmt.Sprintf("admin@%s", config.SiteHost)
	config.SiteImage = app.Image{
		Url:    "/static/meta-img.webp",
		Alt:    "some img alt here",
		Type:   "webp",
		Width:  "800",
		Height: "450",
	}
	config.NavbarLinks = []app.Link{
		{Text: "Golang", Route: "/category/golang/"},
		{Text: "Python", Route: "/category/python/"},
		{Text: "SQL", Route: "/category/sql/"},
		{Text: "About Us", Route: "/about/"},
	}
	config.FooterLinks = []app.Link{
		{Text: "Home", Route: "/"},
		{Text: "About Us", Route: "/about/"},
		{Text: "Contact", Route: "/contact/"},
		{Text: "Privacy Policy", Route: "/privacy-policy/"},
		{Text: "Terms of Service", Route: "/terms-of-service/"},
	}

	// Construct the main Gingersnap engine.
	g := &app.Gingersnap{
		Debug:      true,
		ListenAddr: ":4000",
		Logger:     logger,
		Assets:     app.Assets,
		Media:      http.Dir("./app/assets/media"),
		Templates:  templates,
		Config:     config,
		Posts:      posts,
		Categories: categories,
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

func getAllFilenames(efs *embed.FS) (files []string, err error) {
	if err := fs.WalkDir(efs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		files = append(files, path)

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}
