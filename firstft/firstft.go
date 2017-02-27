package firstft

import (
	"fmt"
	"os"
	// "regexp"
	// "strings"
	"time"
	// "encoding/json"
	"github.com/joho/godotenv"
	. "github.com/gorilla/feeds"
	"regexp"
    "github.com/railsagainstignorance/alignment/content"
)

func getEnvParam(key string, defaultValue string) string {
	godotenv.Load()
	value := os.Getenv(key)

	if value == "" {
		value = defaultValue
	}

	return value
}

func getFirstFTArticles(maxArticles int) *[]*content.Article {

	sRequest := &content.SearchRequest{
		QueryType:         "brand",
		QueryText:         "FirstFT",
		MaxArticles:       maxArticles,
		MaxDurationMillis: 1000,
		SearchOnly:        false, // i.e. return full article details too from CAPI
	}

	fmt.Println("getFirstFTArticles: sRequest=", sRequest, ", maxArticles=", maxArticles)

	sapiResult := content.Search(sRequest)

	articles := []*content.Article {}

	hrefRegexp := regexp.MustCompile(`href="https:\/\/www\.ft\.com\/content\/([0-9a-f\-]+)"`)

	for i, article := range *(sapiResult.Articles) {
		fmt.Println("getFirstFTArticles: ", i, ") ", article.Title)
		articles = append( articles, article )
		// rewrite the firstFT article
		// - attributions, e.g. (FT, Miami Herald)</p> --> (by the FT and Miami Herald)
		// - item title: span class=&#34;ft-bold&#34;&gt;Indian phone fraud?&lt;/span&gt; --> append ":"

		

		// scan the firstFT article for FT article links, get the uuid
		// - e.g. &lt;a href=&#34;https://www.ft.com/content/1f9974ea-f9e7-11e6-9516-2d969e0d3b65&#34;&gt;German business leader&lt;/a&gt;s -> 1f9974ea-f9e7-11e6-9516-2d969e0d3b65
		// - lookup in CAPI, append to list

		matches := hrefRegexp.FindAllStringSubmatch(article.Body, -1)
		if matches != nil {
			fmt.Println("getFirstFTArticles: found ", len(matches), " references to FT articles) ")
			for _,m := range matches {
				uuid := m[1]
				a := content.GetArticle(uuid)
				articles = append( articles, a )
			}
		}

	}

	return &articles
}

// type (content.)Article struct {
// 	SiteUrl       string
// 	Uuid          string
// 	Title         string
// 	Author        string
// 	Excerpt       string
// 	Body          string
// 	PubDateString string
// 	PubDate       *time.Time
// 	ImageUrl    string
// 	ImageWidth  int
// 	ImageHeight int
// 	NonPromoImageUrl    string
// 	NonPromoImageWidth  int
// 	NonPromoImageHeight int
// 	PromoImageUrl    string
// 	PromoImageWidth  int
// 	PromoImageHeight int
// 	PullQuoteAssets *[]PullQuoteAsset
// }

func articlesToRss(articles *[]*content.Article) *string {
	const siteUrl = "http://www.ft.com/"
	now := time.Now()

	feed := &Feed{
		Title:       "Articles for narration from the Financial Times: FT Labs experiment",
		Link:        &Link{Href: siteUrl},
		Description: "Exploring the text to speech possibilities.",
		Author:      &Author{Name: "Chris Gathercole", Email: "chris.gathercole@ft.com"},
		Created:     now,
		Copyright:   "This work is copyright © Financial Times",
	}

	feed.Items = []*Item{}

	for _, article := range *articles {

		created, _ := time.Parse("2006-01-02", article.PubDateString)

		feed.Items = append(feed.Items, &Item{
			Title:       article.Title,
			Link:        &Link{Href: article.SiteUrl},
			Description: article.Body,
			Created:     created,
			Id:          article.Uuid,
		})

	}

	rss, _ := feed.ToRss()
	return &rss
}

func GenerateRss(maxArticles int) *string {
	articles := getFirstFTArticles( maxArticles )
	rssString := articlesToRss( articles )
	return rssString
}

func main() {
	godotenv.Load()
	// ...
}
