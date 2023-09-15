package main

import (
	"gingersnap"
	"log"

	"github.com/fsnotify/fsnotify"
)

func main() {
	s := gingersnap.Settings{
		ConfigPath: "assets/config/gingersnap.json",
		PostsGlob:  "assets/posts/*.md",
		MediaDir:   "assets/media",
		Debug:      true,
	}

	g := gingersnap.New()
	g.Configure(s)

	go RunServerWithWatcher(g, s)

	// Block main goroutine forever.
	<-make(chan struct{})
}

// RunServerWithWatcher runs the server and and watches for file changes.
// On file change, it resets the gingersnap engine and restarts the server.
// .
func RunServerWithWatcher(g *gingersnap.Gingersnap, s gingersnap.Settings) {
	// Create new watcher.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	paths := []string{
		"gingersnap.go",
		"assets/config",
		"assets/css",
		"assets/media",
		"assets/posts",
		"assets/templates",
	}

	for _, path := range paths {
		if err = w.Add(path); err != nil {
			log.Fatal(err)
		}
	}

	g.Logger.Printf("Watching for file changes")

	go g.RunServer()

	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) {
				g.Logger.Println("Files changed. Restarting server")

				g.Configure(s)
				go g.RunServer()
			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			g.Logger.Println("error:", err)
		}
	}
}
