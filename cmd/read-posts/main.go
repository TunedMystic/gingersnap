package main

import (
	"fmt"
	"gingersnap"
	"log"
)

func main() {
	fmt.Println("🚀 Parsing Posts")

	// Construct the Processor.
	processor := gingersnap.NewProcessor(gingersnap.Path("assets/posts"))

	// Parse the markdown posts.
	err := processor.Process()
	if err != nil {
		log.Fatal(err)
	}

	// Printing results.
	fmt.Printf(
		"📝 Parsed %d Posts, %d Categories\n",
		len(processor.PostsBySlug),
		len(processor.CategoriesBySlug),
	)

	postModel := gingersnap.NewPostModel(processor.PostsBySlug)
	fmt.Printf("%+v\n", postModel)

	categoryModel := gingersnap.NewCategoryModel(processor.CategoriesBySlug)
	fmt.Printf("%+v\n", categoryModel)
}
