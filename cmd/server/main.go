package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"gingersnap"
)

func main() {
	localConfigPath := gingersnap.Path("assets/config/gingersnap.json")
	localMediaDir := gingersnap.Path("assets/media")
	localPostsDir := gingersnap.Path("assets/posts")

	// Construct the logger.
	logger := gingersnap.NewLogger()

	// Construct the config
	configBytes := gingersnap.MustRead(localConfigPath)
	config, err := gingersnap.NewConfig(configBytes, true)
	if err != nil {
		logger.Fatal(err)
	}

	// Gather the markdown post files.
	markdownGlob := fmt.Sprintf("%s/%s", localPostsDir, "*.md")
	filePaths, err := filepath.Glob(markdownGlob)
	if err != nil {
		logger.Fatal(err)
	}

	// Parse the markdown posts.
	processor := gingersnap.NewProcessor(filePaths)
	err = processor.Process()
	if err != nil {
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
		Media:      http.Dir(localMediaDir),
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

	g.RunServer()
}
