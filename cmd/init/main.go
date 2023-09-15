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

	// --------------------------------------------------------------
	//
	// Scaffold the new project
	//
	// --------------------------------------------------------------

	gingersnap.CopyDir(gingersnap.Assets, "assets/media", ".")
	gingersnap.CopyDir(gingersnap.Assets, "assets/posts", ".")
	gingersnap.CopyFile(gingersnap.Assets, "assets/config/gingersnap.json", "./gingersnap.json")
}
