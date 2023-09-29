package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gingersnap"
)

func main() {
	if len(os.Args) < 2 {
		Loginfo(helpText)
		os.Exit(0)
	}

	// Settings for gingersnap resources.
	s := gingersnap.Settings{
		ConfigPath: "gingersnap.json",
		PostsDir:   "posts",
		MediaDir:   "media",
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
			Logerr("Config gingersnap.json detected. Skipping.")
		}

		// Copy embedded resources into the current directory.
		gingersnap.CopyDir(gingersnap.Assets, "assets/media", ".")
		gingersnap.CopyDir(gingersnap.Assets, "assets/posts", ".")
		gingersnap.CopyFile(gingersnap.Assets, "assets/config/gingersnap.json", "./gingersnap.json")

		Loginfo("Gingersnap project initialized ✅")

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
			Logerr("No gingersnap.json config detected. Skipping.")
		}

		// Construct the gingersnap engine.
		g := gingersnap.New()
		g.Configure(s)

		// Run the server with file watcher.
		g.RunServerWithWatcher(s)

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
			Logerr("export error: %s", err)
		}

		Loginfo("Site export complete ✅")

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

		// ----------------------------------------------------------
		// Preliminary checks
		// ----------------------------------------------------------

		// The gingersnap project directory.
		ProjectDir := CurrentDir()

		// The directory where the static site will be exported to.
		ExportDir := gingersnap.ExportDir

		// The git repository where the production static site will be stored.
		ProdRepo := g.Config.ProdRepo

		if ProdRepo == "" {
			Logerr("Please specify a git repository to deploy the site to.")
		}

		if ProjectDir == ProdRepo {
			Logerr("Cannot push site to current directory. You must specify another git repository.")
		}

		if !Exists(ProdRepo) {
			Logerr("Dir %s does not exist.", ProdRepo)
		}

		if !Exists(filepath.Join(ProdRepo, ".git")) {
			Logerr("Dir %s is not a git repository.", ProdRepo)
		}

		// ----------------------------------------------------------
		// Export and deploy
		// ----------------------------------------------------------

		//
		// Export the site.
		//
		Loginfo("[1/5] Exporting the site")
		if err := g.Export(); err != nil {
			Logerr("export error: %s", err)
		}

		//
		// Remove all content from the prod repo directory.
		//
		Loginfo("[2/5] Removing static site contents")
		Chdir(ProdRepo)
		Command("git", "rm", "-rf", "--ignore-unmatch", "--quiet", ".")
		Chdir(ProjectDir)

		//
		// Copy the exported site to the prod repo directory.
		//
		Loginfo("[3/5] Copying into the static site directory")
		gingersnap.CopyDir(os.DirFS(ExportDir), ".", ProdRepo)

		//
		// Navigate to the prod repo, commit the changes and push upstream.
		//
		Loginfo("[4/5] Deploying the site")
		Chdir(ProdRepo)
		Command("git", "add", "-A")
		Command("git", "commit", "-m", fmt.Sprintf("Updated site on %s", time.Now().Format(time.UnixDate)))
		Command("git", "push", "-f", "origin", "main")
		Chdir(ProjectDir)

		//
		// Remove the exported site from the project directory.
		//
		Loginfo("[5/5] Cleaning up")
		Remove(ExportDir)

		Loginfo("Site export and deploy complete ✅")

	case "config":

		// ----------------------------------------------------------
		//
		//
		// Config - Explain the config settings
		//
		//
		// ----------------------------------------------------------

		names := []string{}

		for name := range gingersnap.Themes {
			names = append(names, name)
		}

		Loginfo(configCmdText, strings.Join(names, ", "))

	default:
		Logerr(unknownCmdText, os.Args[1])
	}
}

// ------------------------------------------------------------------
//
//
// Utility functions
//
//
// ------------------------------------------------------------------

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

func Chdir(source string) {
	if err := os.Chdir(source); err != nil {
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

func Loginfo(msg string, args ...any) {
	formattedMsg := fmt.Sprintf(msg, args...)
	fmt.Printf("%s\n", formattedMsg)
}

func Logerr(msg string, args ...any) {
	fmt.Printf(msg, args...)
	fmt.Println()
	os.Exit(1)
}

// ------------------------------------------------------------------
//
//
// Help text
//
//
// ------------------------------------------------------------------

var helpText = `
This is the command line interface for Gingersnap,
a simple and opinionated static site generator.

Usage:
  gingersnap [command]

Commands:
  init        Create a new project, and scaffold the required assets
  dev         Start the dev server, and reload on file changes
  export      Export the project as a static site
  deploy      Export the project, and push it to a dedicated repository
  config      Explain the config settings
`

var configCmdText = `
Gingersnap config settings


-------------------------------------------------
site - defines site-specific settings
-------------------------------------------------

  name         The name of the site (ex: MySite)
  host         The host of the site (ex: mysite.com)
  tagline      A short description of the site (50-70 characters)
  description  A long description of the site  (70-155 characters)
  theme        The color theme of the site (%s)

  ex: {
    "name": "MySite",
    "host": "mysite.com",
    "tagline": "short descr ...",
    "description": "longer descr ...",
    "theme": "pink"
  }


-------------------------------------------------
homepage - defines sections for the homepage
-------------------------------------------------

  Contains a list of categories

  The "$latest" represents a pseudo-category
  that contains the latest posts on the site.

  ex: ["category-slug", "$latest"]


-------------------------------------------------
navbarLinks - defines anchor links for the navbar
-------------------------------------------------

  Contains a list of objects

  text         The anchor link text
  href         The anchor link href

  ex: [{"text": "About Us", "href": "/about/"}]


-------------------------------------------------
footerLinks - defines anchor links for the footer
-------------------------------------------------

  Contains a list of objects

  text         The anchor link text
  href         The anchor link href

  ex: [{"text": "About Us", "href": "/about/"}]


-------------------------------------------------
staticRepository - defines the export destination
-------------------------------------------------

  Contains the git repository path
  where the site will be exported to

  ex: "/path/to/static/repo"
`

var unknownCmdText = `
Unknown command "%s"

Run 'gingersnap' for help with usage.
`
