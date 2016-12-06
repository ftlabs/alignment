package main

import (
	"fmt"
	"os"
	"log"
	// "regexp"
	// "strings"
	"strconv"
	"encoding/json"
	"github.com/joho/godotenv"
    "github.com/railsagainstignorance/alignment/content"
    "github.com/railsagainstignorance/alignment/image"
)

func getEnvParam(key string, defaultValue string) string {
	godotenv.Load()
	value := os.Getenv(key)

	if value == "" {
		value = defaultValue
	}

	return value
}

// var keywordsCsv = getEnvParam("KEYWORDS_CSV", "if,and,but,light,you")
// var keywords    = strings.Split(keywordsCsv, ",")
// var defaultImageUrl    = `https://www.ft.com/__origami/service/image/v2/images/raw/http%3A%2F%2Fprod-upp-image-read.ft.com%2F69f10230-2272-11e6-aa98-db1e01fabc0c?source=next&fit=scale-down&compression=best&width=600`
// var defaultImageWidth  = 600
// var defaultImageHeight = 338

type PullQuote struct {
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

func GetPullQuotesWithImages(ontologyName string, ontologyValue string, maxArticles int, maxMillis int) *[]*PullQuote {

	sRequest := &content.SearchRequest{
		QueryType:         ontologyName,
		QueryText:         ontologyValue,
		MaxArticles:       maxArticles,
		MaxDurationMillis: maxMillis,
		SearchOnly:        false,
	}

	fmt.Println("GetPullQuotesWithImages: sRequest=", sRequest, ", maxArticles=", maxArticles, ", maxMillis=", maxMillis)

	sapiResult := content.Search(sRequest)

	items := []*PullQuote {}

	for i, article := range *(sapiResult.Articles) {
		fmt.Println("GetPullQuotesWithImages: ", i, ") ", article.Title)
	}

	// for i, rssItem := range *rssItemsIncludingMissingImages {
	// 	if rssItem.Uuid == "" {
	// 		fmt.Println("meditation: GetHaikusWithImages: discarding (no Uuid) rssItem=", rssItem)
	// 	} else {
	// 		item := &MeditationHaiku{
	// 			Author:       rssItem.Author,
	// 			Title:        rssItem.Title,
	// 			Url:          rssItem.Url,
	// 			DateSelected: rssItem.DateSelected,
	// 			ImageUrl:     rssItem.ImageUrl,
	// 			Themes:       rssItem.Themes,
	// 			Uuid:         rssItem.Uuid,
	// 		}

	// 		capiArticle := content.GetArticle(item.Uuid)
	// 		item.ImageUrl      = capiArticle.ImageUrl
	// 		item.ImageWidth    = capiArticle.ImageWidth
	// 		item.ImageHeight   = capiArticle.ImageHeight
	// 		item.PromoImageUrl    = capiArticle.PromoImageUrl
	// 		item.PromoImageWidth  = capiArticle.PromoImageWidth
	// 		item.PromoImageHeight = capiArticle.PromoImageHeight
	// 		item.NonPromoImageUrl    = capiArticle.NonPromoImageUrl
	// 		item.NonPromoImageWidth  = capiArticle.NonPromoImageWidth
	// 		item.NonPromoImageHeight = capiArticle.NonPromoImageHeight
	// 		item.PubDateString = capiArticle.PubDateString
	// 		item.PubDateEpoch  = capiArticle.PubDate.Unix()

	// 		item.Title = capiArticle.Title // cos is sometimes missing from the rss feed

	// 		if item.ImageUrl == "" {
	// 			fmt.Println("meditation: GetHaikusWithImages: no ImageUrl: item=", item)
	// 			item.ImageUrl    = defaultImageUrl
	// 			item.ImageWidth  = defaultImageWidth
	// 			item.ImageHeight = defaultImageHeight
	// 		} 

	// 		items = append( items, item )

	// 		item.TextWithBreaks = reLineBreaks.ReplaceAllString(rssItem.TextRaw, "<BR>")

	// 		haikuPieces := reHaikuPieces.FindAllString(rssItem.TextRaw, -1)
	// 		item.Id = string( strings.Join(haikuPieces, "") )

	// 		themes := []string {}
	// 		for _,theme := range *item.Themes {
	// 			themes = append(themes, strings.ToUpper(theme))
	// 		}
	// 		keywordMatches := findKeywordMatches( rssItem.TextRaw )
	// 		for _,keyword := range *keywordMatches {
	// 			themes = append( themes, keyword )
	// 		}

	// 		item.Themes = &themes
	// 		fmt.Println("meditation: GetHaikusWithImages: item.Themes=", item.Themes)

	// 		item.ProminentColours = image.GetProminentColours( item.ImageUrl )

	// 		fmt.Println("meditation: GetHaikusWithImages: ", i, ") ", item.Title, ", imageUrl=", item.ImageUrl )
	// 	}
	// }
	return &items
}

func main() {
	godotenv.Load()
	var maxArticles, _ = strconv.Atoi( getEnvParam("MAX_ARTICLES",   "10") )
	var maxMillis,   _ = strconv.Atoi( getEnvParam("MAX_MILLIS",   "3000") )

	pqs := GetPullQuotesWithImages( "before", "2014-03-18T19:00:00Z", maxArticles, maxMillis )
	pqsB, _ := json.Marshal(pqs)

    ofile, err := os.Create("pullquotes.json")
    if err != nil {
        log.Fatal("pullquotes:main: Cannot create file", err)
    }
    defer ofile.Close()

    fmt.Fprint(ofile, string(pqsB))
}