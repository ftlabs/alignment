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

func main() {
	godotenv.Load()
	var maxArticles, _ = strconv.Atoi( getEnvParam("MAX_ARTICLES",   "10") )
	var maxMillis,   _ = strconv.Atoi( getEnvParam("MAX_MILLIS",   "3000") )

	pqs := GetPullQuotesWithImages( "before", "2016-12-06T19:00:00Z", maxArticles, maxMillis )
	pqsB, _ := json.Marshal(pqs)

    ofile, err := os.Create("pullquotes.json")
    if err != nil {
        log.Fatal("pullquotes:main: Cannot create file", err)
    }
    defer ofile.Close()

    fmt.Fprint(ofile, string(pqsB))
}