package main

import (
	"fmt"
	"os"
	"log"
	"regexp"
	"strings"
	"strconv"
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

var keywordsCsv = getEnvParam("KEYWORDS_CSV", "if,and,but,light,you")
var keywords    = strings.Split(keywordsCsv, ",")
var defaultImageUrl    = `https://www.ft.com/__origami/service/image/v2/images/raw/http%3A%2F%2Fprod-upp-image-read.ft.com%2F69f10230-2272-11e6-aa98-db1e01fabc0c?source=next&fit=scale-down&compression=best&width=600`
var defaultImageWidth  = 600
var defaultImageHeight = 338

func findKeywordMatches( text string ) *[]string {
	matchingKeywords := []string {}

	for _,keyword := range keywords {
		matched,_ := regexp.MatchString("(?i)\\b" + keyword + "\\b", text)
		if matched {
			matchingKeywords = append( matchingKeywords, keyword )
		}
	}

	return &matchingKeywords
}

type MeditationHaiku struct {
	Author       string
	Title        string
	Url          string
	DateSelected string
	ImageUrl      string
	ImageWidth    int
	ImageHeight   int
	PromoImageUrl    string
	PromoImageWidth  int
	PromoImageHeight int
	NonPromoImageUrl    string
	NonPromoImageWidth  int
	NonPromoImageHeight int
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

			capiArticle := content.GetArticle(item.Uuid)
			item.ImageUrl      = capiArticle.ImageUrl
			item.ImageWidth    = capiArticle.ImageWidth
			item.ImageHeight   = capiArticle.ImageHeight
			item.PromoImageUrl    = capiArticle.PromoImageUrl
			item.PromoImageWidth  = capiArticle.PromoImageWidth
			item.PromoImageHeight = capiArticle.PromoImageHeight
			item.NonPromoImageUrl    = capiArticle.NonPromoImageUrl
			item.NonPromoImageWidth  = capiArticle.NonPromoImageWidth
			item.NonPromoImageHeight = capiArticle.NonPromoImageHeight
			item.PubDateString = capiArticle.PubDateString
			item.PubDateEpoch  = capiArticle.PubDate.Unix()

			item.Title = capiArticle.Title // cos is sometimes missing from the rss feed

			if item.ImageUrl == "" {
				fmt.Println("meditation: GetHaikusWithImages: no ImageUrl: item=", item)
				item.ImageUrl    = defaultImageUrl
				item.ImageWidth  = defaultImageWidth
				item.ImageHeight = defaultImageHeight
			} 

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
	return &items
}

func main() {
	godotenv.Load()
	var maxItems, _ = strconv.Atoi( getEnvParam("MAX_ITEMS", "1000") )

	haikus := GetHaikusWithImages( maxItems )
	haikusB, _ := json.Marshal(haikus)

    ofile, err := os.Create("meditation_haiku.json")
    if err != nil {
        log.Fatal("meditation:main: Cannot create file", err)
    }
    defer ofile.Close()

    fmt.Fprint(ofile, string(haikusB))
}