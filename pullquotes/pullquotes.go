package main

import (
	"fmt"
	"os"
	"log"
	// "regexp"
	// "strings"
	"strconv"
	"time"
	// "encoding/json"
	"github.com/joho/godotenv"
	. "github.com/gorilla/feeds"
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

var defaultImageUrl    = `https://www.ft.com/__origami/service/image/v2/images/raw/http%3A%2F%2Fprod-upp-image-read.ft.com%2F69f10230-2272-11e6-aa98-db1e01fabc0c?source=next&fit=scale-down&compression=best&width=600`
var defaultImageWidth  = 600
var defaultImageHeight = 338

type PullQuote struct {
	Author         string
	Title          string
	Url            string
	Uuid           string
	PubDateString  string
	PubDateEpoch   int64
	TextWithBreaks string
	ImageUrl       string
	ImageWidth     int
	ImageHeight    int
	ProminentColours *[]image.ProminentColour
	PullQuoteAssets *[]content.PullQuoteAsset
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

		if len(*article.PullQuoteAssets) == 0 {
			fmt.Println("GetPullQuotesWithImages: found no PullQuoteAssets")
		} else {
			fmt.Println("GetPullQuotesWithImages: found some PullQuoteAssets")
			
			item := &PullQuote{
					Author:          article.Author,
					Title:           article.Title,
					Url:             article.SiteUrl,
					Uuid:            article.Uuid,
					ImageUrl:        article.ImageUrl,
					ImageWidth:      article.ImageWidth,
					ImageHeight:     article.ImageHeight,
					PubDateString:   article.PubDateString,
					PubDateEpoch:    article.PubDate.Unix(),
					PullQuoteAssets: article.PullQuoteAssets,
			}

			if item.ImageUrl == "" {
				item.ImageUrl    = defaultImageUrl
				item.ImageWidth  = defaultImageWidth
				item.ImageHeight = defaultImageHeight
			} 

			item.ProminentColours = image.GetProminentColours( item.ImageUrl )

			items = append( items, item )
		}
	}

	return &items
}

func pullQuotesToRss(pullQuotes *[]*PullQuote) *string {
	const siteUrl = "http://www.ft.com/"
	now := time.Now()

	feed := &Feed{
		Title:       "Pull Quotes from the latest Financial Times articles",
		Link:        &Link{Href: siteUrl},
		Description: "Exploring the visceral impact of Pull Quotes in Financial Times articles.",
		Author:      &Author{Name: "Chris Gathercole", Email: "chris.gathercole@ft.com"},
		Created:     now,
		Copyright:   "This work is copyright © Financial Times",
	}

	feed.Items = []*Item{}

	for _, pq := range *pullQuotes {

		created, _ := time.Parse("2006-01-02", pq.PubDateString)

		for pqAssetI, pqAsset := range *pq.PullQuoteAssets {
			guid       := pq.Url + "#" + strconv.Itoa(pqAssetI)
			var description string
			if pqAsset.Attribution == "" {
				description = pqAsset.Body
			} else {
				description = pqAsset.Body + "<p>" + pqAsset.Attribution + "</p>"
			}

			feed.Items = append(feed.Items, &Item{
				Title:       pq.Title,
				Link:        &Link{Href: pq.Url},
				Description: description,
				Created:     created,
				Id:          guid,
			})
		}
	}

	rss, _ := feed.ToRss()
	return &rss
}

func GenerateRss(ontologyName string, ontologyValue string, maxArticles int, maxMillis int) *string {
	pqs := GetPullQuotesWithImages( ontologyName, ontologyValue, maxArticles, maxMillis )
	rssString := pullQuotesToRss( pqs )
	return rssString
}

func main() {
	godotenv.Load()
	var maxArticles, _ = strconv.Atoi( getEnvParam("MAX_ARTICLES",   "10") )
	var maxMillis,   _ = strconv.Atoi( getEnvParam("MAX_MILLIS",   "3000") )

	// pqs := GetPullQuotesWithImages( "before", "2016-12-06T19:00:00Z", maxArticles, maxMillis )
	// pqsB, _ := json.Marshal(pqs)

 //    ofile, err := os.Create("pullquotes.json")
 //    if err != nil {
 //        log.Fatal("pullquotes:main: Cannot create file", err)
 //    }
 //    defer ofile.Close()

	rssString := GenerateRss( "before", "2016-12-06T19:00:00Z", maxArticles, maxMillis )

    ofile, err := os.Create("pullquotes.rss")
    if err != nil {
        log.Fatal("pullquotes:main: Cannot create file", err)
    }
    defer ofile.Close()


    fmt.Fprint(ofile, *rssString)
}