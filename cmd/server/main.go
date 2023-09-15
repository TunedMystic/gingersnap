package main

import (
	"fmt"
	"os"
)

func main() {
	f, err := os.Stat("assets/config/gingersnaps.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("all good", f)
	return

	// s := gingersnap.Settings{
	// 	ConfigPath: "assets/config/gingersnap.json",
	// 	PostsGlob:  "assets/posts/*.md",
	// 	MediaDir:   "assets/media",
	// }

	// g := gingersnap.New()
	// g.Init(s)

	// go g.RunServerWithWatcher(s)

	// // Block main goroutine forever.
	// <-make(chan struct{})
}
