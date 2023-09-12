package main

import (
	"fmt"
	"gingersnap"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	configPath := "assets/config/gingersnap.json"
	postsGlob := "assets/posts/*.md"
	mediaDir := "assets/media"

	// Construct the logger.
	logger := gingersnap.NewLogger()

	// Construct the config
	configBytes, err := gingersnap.ReadFile(configPath)
	if err != nil {
		logger.Fatal(err)
	}

	config, err := gingersnap.NewDebugConfig(configBytes)
	if err != nil {
		logger.Fatal(err)
	}

	// Gather the markdown post files.
	filePaths, err := filepath.Glob(postsGlob)
	if err != nil {
		logger.Fatal(err)
	}

	// Parse the markdown posts.
	processor := gingersnap.NewProcessor(filePaths)
	if err := processor.Process(); err != nil {
		logger.Fatal(err)
	}

	// Construct the models from the processed markdown posts.
	postModel := gingersnap.NewPostModel(processor.PostsBySlug)
	categoryModel := gingersnap.NewCategoryModel(processor.CategoriesBySlug)

	// Construct the templates, using the embedded FS.
	templates, err := gingersnap.NewTemplate(gingersnap.Templates)
	if err != nil {
		logger.Fatal(err)
	}

	// Construct the main Gingersnap engine.
	g := &gingersnap.Gingersnap{
		Logger:     logger,
		Assets:     gingersnap.Assets,
		Media:      http.Dir(mediaDir),
		Templates:  templates,
		Config:     config,
		Posts:      postModel,
		Categories: categoryModel,
	}

	// Construct Http Server.
	g.HttpServer = &http.Server{
		Addr:         g.Config.ListenAddr,
		Handler:      g.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	urls := []string{"/", "/404/", "/sitemap.xml", "/robots.txt", "/CNAME", "/styles.css"}

	for _, post := range g.Posts.All() {
		urls = append(urls, post.Route())
	}

	for _, cat := range g.Categories.All() {
		urls = append(urls, cat.Route())
	}

	// --------------------------------------------------------------
	//
	// Construct the Exporter
	//
	// --------------------------------------------------------------

	exporter := &gingersnap.Exporter{
		Handler:    g.Routes(),
		Urls:       g.AllUrls(),
		MediaDir:   os.DirFS("assets/media"),
		OutputPath: "dist",
	}

	// Perform site export.
	fmt.Println("🤖 Exporting Site")
	err = exporter.Export()
	if err != nil {
		log.Fatal(err)
	}
}
