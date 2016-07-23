package rss

import (
	"fmt"
	"github.com/railsagainstignorance/alignment/Godeps/_workspace/src/github.com/joho/godotenv"
	. "github.com/gorilla/feeds"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"encoding/json"
)

func getJsonUrl() string {
	const jsonUrlEnvParamName = "HAIKU_JSON_URL"

    godotenv.Load()
    url := os.Getenv(jsonUrlEnvParamName)

    if url == "" {
        fmt.Println("rss: getJsonUrl: no such env param: ", jsonUrlEnvParamName)
    }

    return url
}

var jsonUrl = getJsonUrl()

func getJsonBody(url string) (*[]byte) {
    fmt.Println("rss: getJsonBody: url=", url)

    req, err := http.NewRequest("GET", url, nil)
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    fmt.Println("rss: getJsonBody: response Status:", resp.Status)
    jsonBody, _ := ioutil.ReadAll(resp.Body)

    return &jsonBody
}

func parseJsonToGenerateRss( jsonBody *[]byte, maxItems int ) (*string) {
	const hiddenHaikuUrl = "http://www.ft.com/hidden-haiku"
    var data interface{}
    json.Unmarshal(*jsonBody, &data)

	now := time.Now()

	feed := &Feed{
	    Title:       "Haiku found verbatim in Financial Times articles",
	    Link:        &Link{Href: hiddenHaikuUrl},
	    Description: "There are plenty more such haiku: identified by computer algorithm, selected by a human, and brought to light in ft.com/hidden-haiku.",
	    Author:      &Author{Name: "Chris Gathercole", Email: "chris.gathercole@ft.com"},
	    Created:     now,
	    Copyright:   "This work is copyright © Financial Times",
	}

	feed.Items = []*Item{}

    for i, item := range data.([]interface{}) {
    	if i >= maxItems { break }

	    mItem := item.(map[string]interface{})
	    author       := "unknown author"
	    title        := "unknown title"
	    url          := hiddenHaikuUrl
	    haiku        := "unknown haiku"
	    dateSelected := "2016-01-02"

	    if mItem["by"          ] != nil { author       = mItem["by"          ].(string) }
    	if mItem["title"       ] != nil { title        = mItem["title"       ].(string) }
    	if mItem["articleurl"  ] != nil { url          = mItem["articleurl"  ].(string) }
    	if mItem["haikuhtml"   ] != nil { haiku        = mItem["haikuhtml"   ].(string) }
    	if mItem["dateselected"] != nil { dateSelected = mItem["dateselected"].(string) }

    	created, _  := time.Parse("2006-01-02", dateSelected )

    	feed.Items = append( feed.Items, &Item{
    		Title:       title,
        	Link:        &Link{Href: url},
        	Description: "<it>" + haiku + "</it>",
       		Author:      &Author{Name: author},
       		Created:     created,
       		} )
    }

    rss, _ := feed.ToRss()
    return &rss
}

func Generate(maxItems int) *string {
	jsonBody  := getJsonBody(jsonUrl)
	rssString := parseJsonToGenerateRss( jsonBody, maxItems )
	return rssString
}

func main() {
	godotenv.Load()
	rssString := Generate(10)

	fmt.Println("main: rssString=", *rssString)
}
