package main

import (
	"fmt"
	"gingersnap"
	"log"
)

func main() {
	fmt.Println("🚀 Parsing Posts")

	// Construct the PostManager.
	postManager := gingersnap.NewPostManager(gingersnap.Path("assets/posts"))

	// Run.
	err := postManager.Process()
	if err != nil {
		log.Fatal(err)
	}

	// Printing results.
	fmt.Printf(
		"📝 Parsed %d Posts, %d Categories\n",
		len(postManager.PostsBySlug),
		len(postManager.CategoriesBySlug),
	)

	postModel := gingersnap.NewPostModel(postManager.PostsBySlug)
	fmt.Printf("%+v\n", postModel)

	categoryModel := gingersnap.NewCategoryModel(postManager.CategoriesBySlug)
	fmt.Printf("%+v\n", categoryModel)
}
