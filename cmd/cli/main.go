package main

import (
	"fmt"
	"gingersnap"
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
		ConfigPath: "gingersnap.json",
		PostsGlob:  "posts/*.md",
		MediaDir:   "media",
		Debug:      false,
	}

	switch os.Args[1] {

	case "init":

		// ----------------------------------------------------------
		//
		//
		// Init - Create a new project. Scaffold assets.
		//
		//
		// ----------------------------------------------------------

		// If gingersnap.json exists in the current directory,
		// then do not scaffold a new project here.
		if gingersnap.Exists("./gingersnap.json") {
			fmt.Printf("\nConfig gingersnap.json detected. Skipping.\n\n")
			os.Exit(1)
		}

		// Copy embedded resources into the current directory.
		gingersnap.CopyDir(gingersnap.Assets, "assets/media", ".")
		gingersnap.CopyDir(gingersnap.Assets, "assets/posts", ".")
		gingersnap.CopyFile(gingersnap.Assets, "assets/config/gingersnap.json", "./gingersnap.json")

	case "dev":

		// ----------------------------------------------------------
		//
		//
		// Dev - Start the dev server, and reload on file changes.
		//
		//
		// ----------------------------------------------------------

		// If gingersnap.json does not exist in the current directory,
		// then do not start the server
		if !gingersnap.Exists("./gingersnap.json") {
			fmt.Printf("\nNo gingersnap.json config detected. Skipping.\n\n")
			os.Exit(1)
		}

		// Construct the gingersnap engine.
		g := gingersnap.New()
		g.Configure(s)

		// Run the server.
		go g.RunServerWithWatcher(s)

		// Block main goroutine forever.
		<-make(chan struct{})

	case "export":

		// ----------------------------------------------------------
		//
		//
		// Export - Export the project as a static site.
		//
		//
		// ----------------------------------------------------------

		// Construct the gingersnap engine.
		g := gingersnap.New()
		g.Configure(s)

		// Export the site.
		if err := g.Export(); err != nil {
			fmt.Printf("\nexport error: %s\n\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf(unknownCmdText, os.Args[1])
		os.Exit(1)
	}
}
