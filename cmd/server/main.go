package main

import app "gingersnap"

func main() {
	s := app.Settings{
		ConfigPath: "assets/config/gingersnap.json",
		PostsDir:   "assets/posts",
		MediaDir:   "assets/media",
		ExportDir:  "dist",
		Debug:      true,
	}

	g := app.NewGingersnap()
	g.Configure(s)
	g.RunServer()
}
