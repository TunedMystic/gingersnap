package main

import (
	"gingersnap"
)

func main() {
	s := gingersnap.Settings{
		ConfigPath: "assets/config/gingersnap.json",
		PostsGlob:  "assets/posts/*.md",
		MediaDir:   "assets/media",
	}

	g := gingersnap.New()
	g.Init(s)

	go g.RunServerWithWatcher(s)

	// Block main goroutine forever.
	<-make(chan struct{})
}
