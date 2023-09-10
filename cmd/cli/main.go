package main

import (
	"fmt"
	"os"
)

var helpText = `
This is the command line interface for Gingersnap,
a simple and opinionated static site generator.

Usage:
  gingersnap [command]

Commands:
  init        Create a new project, and scaffold the required assets
  dev         Start the dev server, and reload on file changes
  export      Export the project as a static site
`

var unknownCmdText = `
Unknown command "%s"

Run 'gingersnap' for help with usage.

`

func main() {

	if len(os.Args) < 2 {
		fmt.Println(helpText)
		os.Exit(0)
	}

	switch os.Args[1] {
	case "init":
		fmt.Println("Subcommand [init]")
	case "dev":
		fmt.Println("Subcommand [dev]")
	case "export":
		fmt.Println("Subcommand [export]")
	default:
		fmt.Printf(unknownCmdText, os.Args[1])
		os.Exit(1)
	}

	// localConfigPath := gingersnap.Path("assets/config/gingersnap.json")
	// localMediaDir := gingersnap.Path("assets/media")
	// localPostsDir := gingersnap.Path("assets/posts")

	// // Construct logger.
	// logger := gingersnap.NewLogger()

	// // Construct the templates, using the embedded FS.
	// templates, err := gingersnap.NewTemplate(gingersnap.Templates)
	// if err != nil {
	// 	logger.Fatal(err)
	// }

	// // Construct the PostManager.
	// postManager := gingersnap.NewPostManager(localPostsDir)

	// // Parse the markdown posts.
	// err = postManager.Process()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Construct the models.
	// postModel := gingersnap.NewPostModel(postManager.PostsBySlug)
	// categoryModel := gingersnap.NewCategoryModel(postManager.CategoriesBySlug)

	// // Construct the config
	// config, err := gingersnap.NewConfig(localConfigPath, true)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Construct the main Gingersnap engine.
	// g := &gingersnap.Gingersnap{
	// 	Logger:     logger,
	// 	Assets:     gingersnap.Assets,
	// 	Media:      http.Dir(localMediaDir),
	// 	Templates:  templates,
	// 	Config:     config,
	// 	Posts:      postModel,
	// 	Categories: categoryModel,
	// }

	// // Construct Http Server.
	// g.HttpServer = &http.Server{
	// 	Addr:         g.Config.ListenAddr,
	// 	Handler:      g.Routes(),
	// 	IdleTimeout:  time.Minute,
	// 	ReadTimeout:  10 * time.Second,
	// 	WriteTimeout: 10 * time.Second,
	// }

	// g.RunServer()
}
