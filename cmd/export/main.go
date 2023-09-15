package main

import (
	"fmt"
	"gingersnap"
	"os"
)

func main() {
	s := gingersnap.Settings{
		ConfigPath: "assets/config/gingersnap.json",
		PostsGlob:  "assets/posts/*.md",
		MediaDir:   "assets/media",
	}

	g := gingersnap.New()
	g.Configure(s)

	// Export the site.
	if err := g.Export(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
