package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"

	"gingersnap/app"
	"gingersnap/app/utils"
)

func main() {
	if len(os.Args) < 2 {
		loginfo(helpText)
		os.Exit(0)
	}

	g := app.NewGingersnap()
	g.Debug = true
	g.ConfigPath = "gingersnap.json"
	g.PostsPath = "posts"
	g.MediaPath = "media"
	g.ExportPath = "dist"

	switch os.Args[1] {
	case "version":

		// ----------------------------------------------------------
		//
		//
		// Version - View build info
		//
		//
		// ----------------------------------------------------------

		loginfo("\nGingersnap")
		loginfo("  Built At    %s", BuildDate)
		loginfo("  Git Hash    %s\n", BuildHash)

	case "init":

		// ----------------------------------------------------------
		//
		//
		// Init - Create a new project. Scaffold assets.
		//
		//
		// ----------------------------------------------------------

		// If the config exists in the current directory,
		// then do not scaffold a new project here.
		if utils.Exists(g.ConfigPath) {
			logerr("Project already initialized. Skipping.")
		}

		// Copy embedded resources into the current directory.
		g.Unpack()

		loginfo("Gingersnap project initialized ✅")

	case "dev":

		// ----------------------------------------------------------
		//
		//
		// Dev - Start the dev server, and reload on file changes.
		//
		//
		// ----------------------------------------------------------

		// Check that the project files exist.
		ensureProject(g)

		// Configure the gingersnap engine.
		g.Configure()

		// Run the server with file watcher.
		runServerWithWatcher(g)

	case "webp":

		// ----------------------------------------------------------
		//
		//
		// Webp - Convert images to webp format.
		//
		//
		// ----------------------------------------------------------

		// Check that the project files exist.
		ensureProject(g)

		w := app.NewWebp()

		// Gather the images in the media directory.
		imgPaths, err := utils.LocalGlob(g.MediaPath, "png", "jpg", "jpeg")
		if err != nil {
			logerr("webp error: %s", err)
		}

		// Convert all the images in the media directory to webp.
		if err := w.Convert(imgPaths...); err != nil {
			logerr("webp error: %s", err)
		}

	case "export":

		// ----------------------------------------------------------
		//
		//
		// Export - Export the project as a static site.
		//
		//
		// ----------------------------------------------------------

		// Check that the project files exist.
		ensureProject(g)

		g.Debug = false

		// Configure the gingersnap engine.
		g.Configure()

		// Export the site.
		if err := g.Export(); err != nil {
			logerr("export error: %s", err)
		}

		loginfo("Site export complete ✅")

	case "deploy":

		// ----------------------------------------------------------
		//
		//
		// Deploy - Export the project, and push it to a dedicated repository.
		//
		//
		// ----------------------------------------------------------

		// Check that the project files exist.
		ensureProject(g)

		g.Debug = false

		// Configure the gingersnap engine.
		g.Configure()

		// [1/2] Preliminary checks ---------------------------------

		// The gingersnap project directory.
		projectDir := currentDir()

		// The git repository where the production static site will be stored.
		repository := g.Repository()

		if repository == "" {
			logerr("Please specify a git repository to deploy the site to.")
		}

		if projectDir == repository {
			logerr("Cannot push site to current directory. You must specify another git repository.")
		}

		if !utils.Exists(repository) {
			logerr("Dir %s does not exist.", repository)
		}

		if !utils.Exists(filepath.Join(repository, ".git")) {
			logerr("Dir %s is not a git repository.", repository)
		}

		// [2/2] Export and Deploy ----------------------------------

		// Export the site.
		//
		loginfo("[1/6] Exporting the site")
		if err := g.Export(); err != nil {
			logerr("export error: %s", err)
		}

		// Remove all content from the prod repo directory.
		//
		loginfo("[2/6] Removing static site contents")
		chdir(repository)
		command("git", "rm", "-rf", "--ignore-unmatch", "--quiet", ".")
		chdir(projectDir)

		// Copy the exported site to the prod repo directory.
		//
		loginfo("[3/6] Copying into repository")
		utils.CopyDir(os.DirFS(g.ExportPath), ".", repository)

		// Navigate to the prod repo and commit the changes.
		//
		loginfo("[4/6] Commiting changes")
		chdir(repository)
		command("git", "add", "-A")
		command("git", "commit", "-m", fmt.Sprintf("Updated site on %s", time.Now().Format(time.UnixDate)))
		loginfo("[5/6] Pushing site upstream")
		command("git", "push", "-f", "origin", "main")
		chdir(projectDir)

		// Remove the exported site from the project directory.
		//
		loginfo("[6/6] Cleaning up")
		remove(g.ExportPath)

		loginfo("Site export and deploy complete ✅")

	case "clean":

		// ----------------------------------------------------------
		//
		//
		// Clean - Remove temp files and dirs
		//
		//
		// ----------------------------------------------------------

		loginfo("[1/2] Remove temp directories")
		remove(g.ExportPath)

		loginfo("[2/2] Remove helper directories")
		w := app.NewWebp()

		// Remove the `cwebp` binary.
		if err := w.Clean(); err != nil {
			logerr("webp error: %s", err)
		}

	default:
		loginfo("Unknown command '%s'", os.Args[1])
		loginfo("Run 'gingersnap' for help with usage")
		loginfo("")
	}
}

// ------------------------------------------------------------------
//
//
// Utility functions
//
//
// ------------------------------------------------------------------

func currentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return dir
}

func command(cmds ...string) {
	_, err := exec.Command(cmds[0], cmds[1:]...).Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func chdir(source string) {
	if err := os.Chdir(source); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func remove(source string) {
	if err := os.RemoveAll(source); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func loginfo(msg string, args ...any) {
	formattedMsg := fmt.Sprintf(msg, args...)
	fmt.Printf("%s\n", formattedMsg)
}

func logerr(msg string, args ...any) {
	fmt.Printf(msg, args...)
	fmt.Println()
	os.Exit(1)
}

// ------------------------------------------------------------------
//
//
// Helper functions
//
//
// ------------------------------------------------------------------

// ensureProject checks that the required configs/dirs
// exist for the gingersnap engine.
// .
func ensureProject(g *app.Gingersnap) {
	// If the config does not exist in the current directory,
	// then do not start the server.
	if !utils.Exists(g.ConfigPath) {
		logerr("No config detected. Skipping.")
	}

	// If the media directory does not exist in the current directory,
	// then do not start the server.
	if !utils.Exists(g.MediaPath) {
		logerr("No media directory detected. Skipping.")
	}

	// If the posts directory does not exist in the current directory,
	// then do not start the server.
	if !utils.Exists(g.PostsPath) {
		logerr("No posts directory detected. Skipping.")
	}
}

// runServerWithWatcher runs the server and and watches for file changes.
// On file change, it resets the gingersnap engine and restarts the server.
// .
func runServerWithWatcher(g *app.Gingersnap) error {
	// Create new watcher.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err = w.Add(g.ConfigPath); err != nil {
		return err
	}

	if err = w.Add(utils.SafeDir(g.PostsPath)); err != nil {
		return err
	}

	if err = w.Add(utils.SafeDir(g.MediaPath)); err != nil {
		return err
	}

	fmt.Println("Watching for file changes")

	go g.RunServer()

	for {
		select {
		case event := <-w.Events:
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) {
				fmt.Println("Files changed. Restarting server")

				g.Configure()
				go g.RunServer()
			}
		case err := <-w.Errors:
			return err
		}
	}
}

// ------------------------------------------------------------------
//
//
// Constants
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
  webp        Convert images to webp format
  export      Export the project as a static site
  deploy      Export the project, and push it to a dedicated repository
  clean       Remove temp files and dirs
  version     View build info
`

var (
	BuildDate = ""
	BuildHash = ""
)
