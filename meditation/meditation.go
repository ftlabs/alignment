package main

import (
	"fmt"
	"github.com/joho/godotenv"
    "github.com/railsagainstignorance/alignment/rss"
    "github.com/railsagainstignorance/alignment/content"
)

func GetHaikusWithImages(maxItems int) *[]*rss.Haiku {
	items := rss.GenerateItems( maxItems )
	for i, item := range *items {
		if item.ImageUrl == "" && item.Uuid != "" {
			capiArticle := content.GetArticle(item.Uuid)
			item.ImageUrl = capiArticle.ImageUrl
		}
		fmt.Println("meditation: GetHaikusWithImages: ", i, ") ", item.Title, ", imageUrl=", item.ImageUrl )
	}
	return items
}

func main() {
	godotenv.Load()
	haikus := GetHaikusWithImages( 10 )

	fmt.Println("main: haikus=", *haikus)
}