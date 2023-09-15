package main

import (
	"fmt"
	"gingersnap"
	"log"
	"os"
)

func main() {
	s := gingersnap.Settings{
		ConfigPath: "assets/config/gingersnap.json",
		PostsGlob:  "assets/posts/*.md",
		MediaDir:   "assets/media",
	}

	g := gingersnap.New()
	g.Init(s)

	// --------------------------------------------------------------
	//
	// Construct the Exporter
	//
	// --------------------------------------------------------------

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
}
