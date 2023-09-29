package main

import (
	"gingersnap"
)

func main() {
	s := gingersnap.Settings{
		ConfigPath: "assets/config/gingersnap.json",
		PostsGlob:  "assets/posts/**/*.md",
		MediaDir:   "assets/media",
		Debug:      true,
	}

	g := gingersnap.New()
	g.Configure(s)
	g.RunServer()
}
