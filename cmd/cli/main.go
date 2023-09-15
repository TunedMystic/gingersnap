package main

import (
	"fmt"
	"gingersnap"
	"log"
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

	// Settings for gingersnap resources.
	s := gingersnap.Settings{
		ConfigPath: "assets/config/gingersnap.json",
		PostsGlob:  "assets/posts/*.md",
		MediaDir:   "assets/media",
		Debug:      true,
	}

	switch os.Args[1] {

	case "init":

		fmt.Println("Subcommand [init]")

		// If gingersnap.json exists in the current directory,
		// then do not scaffold a new project here.
		if _, err := os.Stat("./gingersnap.json"); !os.IsNotExist(err) {
			log.Fatal("Config gingersnap.json detected. Skipping")
		}

		// Copy embedded resources into the current directory.
		gingersnap.CopyDir(gingersnap.Assets, "assets/media", ".")
		gingersnap.CopyDir(gingersnap.Assets, "assets/posts", ".")
		gingersnap.CopyFile(gingersnap.Assets, "assets/config/gingersnap.json", "./gingersnap.json")

	case "dev":

		fmt.Println("Subcommand [dev]")

		// Construct the gingersnap engine.
		g := gingersnap.New()
		g.Init(s)

		go g.RunServerWithWatcher(s)

		// Block main goroutine forever.
		<-make(chan struct{})

	case "export":

		fmt.Println("Subcommand [export]")

		// Construct the gingersnap engine.
		g := gingersnap.New()
		g.Init(s)

		// Construct the exporter.
		exporter := &gingersnap.Exporter{
			Handler:    g.Routes(),
			Urls:       g.AllUrls(),
			MediaDir:   os.DirFS(s.MediaDir),
			OutputPath: "dist",
		}

		// Perform site export.
		fmt.Println("🤖 Exporting Site")
		if err := exporter.Export(); err != nil {
			log.Fatal(err)
		}

	default:
		fmt.Printf(unknownCmdText, os.Args[1])
		os.Exit(1)
	}
}
