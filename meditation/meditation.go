package main

import (
	"fmt"
	"os"
	"log"
	"regexp"
	"strings"
	"encoding/json"
	"github.com/joho/godotenv"
    "github.com/railsagainstignorance/alignment/rss"
    "github.com/railsagainstignorance/alignment/content"
    "github.com/railsagainstignorance/alignment/image"
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

type MeditationHaiku struct {
	Author       string
	Title        string
	Url          string
	DateSelected string
	ImageUrl     string
	ImageWidth   int
	PromoImageUrl string
	PromoImageWidth int
	Themes       *[]string
	Uuid         string
	PubDateString    string
	PubDateEpoch     int64
	TextWithBreaks   string
	ProminentColours *[]image.ProminentColour
	Id           string
}


func GetHaikusWithImages(maxItems int) *[]*MeditationHaiku {
	rssItemsIncludingMissingImages := rss.GenerateItems( maxItems )
	items := []*MeditationHaiku {}

	reHaikuPieces := regexp.MustCompile("(?i)[a-z]")
	reLineBreaks  := regexp.MustCompile("\r?\n")

	for i, rssItem := range *rssItemsIncludingMissingImages {
		if rssItem.Uuid == "" {
			fmt.Println("meditation: GetHaikusWithImages: discarding (no Uuid) rssItem=", rssItem)
		} else {
			item := &MeditationHaiku{
				Author:       rssItem.Author,
				Title:        rssItem.Title,
				Url:          rssItem.Url,
				DateSelected: rssItem.DateSelected,
				ImageUrl:     rssItem.ImageUrl,
				Themes:       rssItem.Themes,
				Uuid:         rssItem.Uuid,
			}

			// if item.ImageUrl == "" {
				capiArticle := content.GetArticle(item.Uuid)
				item.ImageUrl        = capiArticle.ImageUrl
				item.ImageWidth      = capiArticle.ImageWidth
				item.PromoImageUrl   = capiArticle.PromoImageUrl
				item.PromoImageWidth = capiArticle.PromoImageWidth
				item.PubDateString = capiArticle.PubDateString
				item.PubDateEpoch  = capiArticle.PubDate.Unix()
			// }
			if item.ImageUrl == "" {
				fmt.Println("meditation: GetHaikusWithImages: discarding (no ImageUrl) item=", item)
			} else {
				items = append( items, item )

				item.TextWithBreaks = reLineBreaks.ReplaceAllString(rssItem.TextRaw, "<BR>")

				haikuPieces := reHaikuPieces.FindAllString(rssItem.TextRaw, -1)
				item.Id = string( strings.Join(haikuPieces, "") )

				themes := []string {}
				for _,theme := range *item.Themes {
					themes = append(themes, strings.ToUpper(theme))
				}
				keywordMatches := findKeywordMatches( rssItem.TextRaw )
				for _,keyword := range *keywordMatches {
					themes = append( themes, keyword )
				}

				item.Themes = &themes
				fmt.Println("meditation: GetHaikusWithImages: item.Themes=", item.Themes)

				item.ProminentColours = image.GetProminentColours( item.ImageUrl )

				fmt.Println("meditation: GetHaikusWithImages: ", i, ") ", item.Title, ", imageUrl=", item.ImageUrl )
			}
		}
	}
	return &items
}

func main() {
	godotenv.Load()
	haikus := GetHaikusWithImages( 1000 )
	haikusB, _ := json.Marshal(haikus)

    ofile, err := os.Create("meditation_haiku.json")
    if err != nil {
        log.Fatal("meditation:main: Cannot create file", err)
    }
    defer ofile.Close()

    fmt.Fprintf(ofile, string(haikusB))
}