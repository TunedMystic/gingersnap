package main

import (
	"gingersnap"
	"log"
	"net/http"
	"time"
)

func main() {
	localConfigPath := gingersnap.Path("assets/config/gingersnap.json")
	localMediaDir := gingersnap.Path("assets/media")
	localPostsDir := gingersnap.Path("assets/posts")

	// Construct logger.
	logger := gingersnap.NewLogger()

	// Construct the templates, using the embedded FS.
	templates, err := gingersnap.NewTemplate(gingersnap.Templates)
	if err != nil {
		logger.Fatal(err)
	}

	// Construct the Processor.
	processor := gingersnap.NewProcessor(localPostsDir)

	// Parse the markdown posts.
	err = processor.Process()
	if err != nil {
		log.Fatal(err)
	}

	// Construct the models.
	postModel := gingersnap.NewPostModel(processor.PostsBySlug)
	categoryModel := gingersnap.NewCategoryModel(processor.CategoriesBySlug)

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

	// --------------------------------------------------------------
	//
	// Scaffold the new project
	//
	// --------------------------------------------------------------

	gingersnap.CopyEmbeddedDir(gingersnap.Assets, "assets/media", "./media")
	gingersnap.CopyEmbeddedDir(gingersnap.Assets, "assets/posts", "./posts")
	gingersnap.CopyEmbeddedFile(gingersnap.Assets, "assets/config/gingersnap.json", "./gingersnap.json")
}
