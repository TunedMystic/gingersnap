package main

import (
	"gingersnap"
)

func main() {
	s := gingersnap.Settings{
		ConfigPath: "assets/config/gingersnap.json",
		PostsDir:   "assets/posts",
		MediaDir:   "assets/media",
		Debug:      true,
	}

	g := gingersnap.New()
	g.Configure(s)
	g.RunServer()
}
