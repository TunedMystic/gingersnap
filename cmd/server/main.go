package main

import "gingersnap/app"

func main() {
	g := app.NewGingersnap()
	g.Configure()
	g.RunServer()
}
