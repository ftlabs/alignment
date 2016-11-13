package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"encoding/json"
	"github.com/joho/godotenv"
    "github.com/railsagainstignorance/alignment/rss"
    "github.com/railsagainstignorance/alignment/content"
)

func getEnvParam(key string, defaultValue string) string {
	godotenv.Load()
	value := os.Getenv(key)

	if key == "" {
		value = defaultValue
	}

	return value
}

var keywordsCsv = getEnvParam("KEYWORDS_CSV", "if,and,but")
var keywords    = strings.Split(keywordsCsv, ",")

func findKeywordMatches( text string ) *[]string {
	matchingKeywords := []string {}

	for _,keyword := range keywords {
		matched,_ := regexp.MatchString("(?i)\\b" + keyword + "\\b", text)
		if matched {
			matchingKeywords = append( matchingKeywords, keyword )
		}
	}
	fmt.Println("meditation: findKeywordMatches: text=", text, ", matchingKeywords=", matchingKeywords)

	return &matchingKeywords
}

func GetHaikusWithImages(maxItems int) *[]*rss.Haiku {
	itemsIncludingMissingImages := rss.GenerateItems( maxItems )
	items := []*rss.Haiku {}

	for i, item := range *itemsIncludingMissingImages {
		if item.Uuid == "" {
			fmt.Println("meditation: GetHaikusWithImages: discarding (no Uuid) item=", item)
		} else {
			if item.ImageUrl == "" {
				capiArticle := content.GetArticle(item.Uuid)
				item.ImageUrl = capiArticle.ImageUrl
			}
			if item.ImageUrl == "" {
				fmt.Println("meditation: GetHaikusWithImages: discarding (no ImageUrl) item=", item)
			} else {
				items = append( items, item )
				item.Text        = ""
				item.TextAsHtml  = ""
				item.Description = ""

				themes := *item.Themes
				keywordMatches := findKeywordMatches( item.TextRaw )
				for _,keyword := range *keywordMatches {
					themes = append( themes, keyword )
				}

				item.Themes = &themes
				fmt.Println("meditation: GetHaikusWithImages: item.Themes=", item.Themes)

				fmt.Println("meditation: GetHaikusWithImages: ", i, ") ", item.Title, ", imageUrl=", item.ImageUrl )
			}
		}
	}
	return &items
}

func main() {
	godotenv.Load()
	haikus := GetHaikusWithImages( 5 )
	haikusB, _ := json.Marshal(haikus)

	fmt.Println("main: haikusB=", string(haikusB) )
}