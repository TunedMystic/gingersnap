package main

import (
	"log"
	"net/http"
	"time"

	"gingersnap"
)

func main() {
	const localConfigPath = "./assets/config/gingersnap.json"
	const localMediaDir = "./assets/media"

	// Construct logger.
	logger := gingersnap.NewLogger()

	// Construct the templates, using the embedded FS.
	templates, err := gingersnap.NewTemplate(gingersnap.Templates)
	if err != nil {
		logger.Fatal(err)
	}

	// Construct the models.
	posts := &gingersnap.PostModel{}
	categories := &gingersnap.CategoryModel{}

	// Construct the config
	config, err := gingersnap.NewConfig(localConfigPath, true)
	if err != nil {
		log.Fatal(err)
	}

	// Construct the main Gingersnap engine.
	g := &gingersnap.Gingersnap{
		Logger:     logger,
		Assets:     gingersnap.Assets,
		Media:      http.Dir(localMediaDir),
		Templates:  templates,
		Config:     config,
		Posts:      posts,
		Categories: categories,
	}

	// Construct Http Server.
	g.HttpServer = &http.Server{
		Addr:         g.Config.ListenAddr,
		Handler:      g.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	g.RunServer()
}
