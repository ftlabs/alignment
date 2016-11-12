package main

import (
	"fmt"
	"github.com/joho/godotenv"
    "github.com/railsagainstignorance/alignment/rss"
)

func GetHaikusWithImages(maxItems int) *[]*rss.Haiku {
	items := rss.GenerateItems( maxItems )
	for i, item := range *items {
		fmt.Println("meditation: GetHaikusWithImages: ", i, ") ", item.Title )
	}
	return items
}

func main() {
	godotenv.Load()
	haikus := GetHaikusWithImages( 10 )

	fmt.Println("main: haikus=", *haikus)
}