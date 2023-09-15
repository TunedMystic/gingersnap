package main

import (
	"fmt"
	"gingersnap"
	"os"
	"os/exec"
	"path/filepath"
	"time"
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
	// s := gingersnap.Settings{
	// 	ConfigPath: "gingersnap.json",
	// 	PostsGlob:  "posts/*.md",
	// 	MediaDir:   "media",
	// 	Debug:      true,
	// }
	s := gingersnap.Settings{
		ConfigPath: "assets/config/gingersnap.json",
		PostsGlob:  "assets/posts/*.md",
		MediaDir:   "assets/media",
		Debug:      true,
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
		if Exists(s.ConfigPath) {
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
		if !Exists(s.ConfigPath) {
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

		s.Debug = false

		// Construct the gingersnap engine.
		g := gingersnap.New()
		g.Configure(s)

		// Export the site.
		if err := g.Export(); err != nil {
			fmt.Printf("\nexport error: %s\n\n", err)
			os.Exit(1)
		}

	case "deploy":

		// ----------------------------------------------------------
		//
		//
		// Deploy - Export the project, and push it to a dedicated repository.
		//
		//
		// ----------------------------------------------------------

		s.Debug = false

		// Construct the gingersnap engine.
		g := gingersnap.New()
		g.Configure(s)

		// Export the site.
		if err := g.Export(); err != nil {
			fmt.Printf("\nexport error: %s\n\n", err)
			os.Exit(1)
		}

		// The directory where the static site will be exported to.
		// This is only a temporary directory. To fully deploy the site,
		// the exported site must be moved to the production site repository.
		ExportDir := gingersnap.ExportDir

		// The git repository where the production static site will be stored.
		ProdRepo := g.Config.ProdRepo

		// Deploy the exported site

		if ProdRepo == "" {
			fmt.Println("Please specify a git repository to deploy the site to.")
			os.Exit(1)
		}

		if CurrentDir() == ProdRepo {
			fmt.Println("Cannot push site to current directory. You must specify another git repository.")
			os.Exit(1)
		}

		fmt.Println("Checking that the static site dir exists and is a repository")
		if !Exists(ProdRepo) {
			fmt.Printf("Dir %s does not exist.\n", ProdRepo)
			os.Exit(1)
		}

		if !Exists(filepath.Join(ProdRepo, ".git")) {
			fmt.Printf("Dir %s is not a git repository.\n", ProdRepo)
			os.Exit(1)
		}

		fmt.Println("Removing current static site contents")
		Command(fmt.Sprintf("cd %s && git rm -rf --ignore-unmatch --quiet .", ProdRepo))

		fmt.Println("Copying into the static site directory")
		Command("cp", "-R", ExportDir, ProdRepo)

		fmt.Println("Deploying the site")
		Command(fmt.Sprintf("cd %s && git add -A && git commit -m 'Updated site on %s' && git push -f origin main", ProdRepo, time.Now().Format(time.Stamp)))

		fmt.Println("Cleaning up")
		Remove(ExportDir)

		fmt.Println("Done!")

	default:
		fmt.Printf(unknownCmdText, os.Args[1])
		os.Exit(1)
	}
}

func CurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return dir
}

func Command(cmds ...string) {
	_, err := exec.Command(cmds[0], cmds[1:]...).Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Exists(source string) bool {
	_, err := os.Stat(source)

	if os.IsNotExist(err) {
		return false
	}
	return true
}

func Remove(source string) {
	if err := os.RemoveAll(source); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
