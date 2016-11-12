package rss

import (
	"encoding/json"
	"fmt"
	. "github.com/gorilla/feeds"
	"github.com/joho/godotenv"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"crypto/md5"
    "encoding/hex"
    "html/template"
)

func GetMD5Hash(text string) string {
    hasher := md5.New()
    hasher.Write([]byte(text))
    return hex.EncodeToString(hasher.Sum(nil))
}

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

func getJsonBody(url string) *[]byte {
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
	fmt.Println("rss: getJsonBody: response:", resp)
	jsonBody, _ := ioutil.ReadAll(resp.Body)

	return &jsonBody
}

type Haiku struct {
	Author       string
	Title        string
	Url          string
	Text         string
	TextAsHtml   template.HTML
	DateSelected string
	Description  string
	ImageUrl     string
	Themes       *[]string
}

func parseJsonToGenerateItems(jsonBody *[]byte, maxItems int) *[]*Haiku {
	const hiddenHaikuUrl = "http://www.ft.com/hidden-haiku"
	var data interface{}
	json.Unmarshal(*jsonBody, &data)

	items := []*Haiku {}

	for i, item := range data.([]interface{}) {
		if i >= maxItems {
			break
		}

		mItem := item.(map[string]interface{})
		author := "unknown author"
		title := "unknown title"
		url := hiddenHaikuUrl
		haiku := "unknown haiku"
		dateSelected := "2016-01-02"
		imageUrl := ""
		themes := []string {}

		for key, value := range mItem {
			if key == "by" && value != nil {
				author = value.(string)
			} else if key == "title" && value != nil {
				title = value.(string)
			} else if key == "articleurl" && value != nil {
				url = value.(string)
			} else if key == "haikuhtml" && value != nil {
				haiku = value.(string)
			} else if key == "dateselected" && value != nil {
				dateSelected = value.(string)
			} else if key == "imageurl" && value != nil {
				imageUrl = value.(string)
			} else {
				if value != nil && value == true {
					themes = append( themes, key )
				}
			}
		}

		description := "<strong>" + haiku + "</strong>" + "<BR>" + "-" + author

		haikuStruct := Haiku{
			Author:       author,
			Title:        title,
			Url:          url,
			Text:         haiku,
			TextAsHtml:   template.HTML(haiku),
			DateSelected: dateSelected,
			Description:  description,
			ImageUrl:     imageUrl,
			Themes:       &themes,
			}

		fmt.Println( "rss: parseJsonToGenerateItems: haikuStruct.Themes=", haikuStruct.Themes)

		items = append( items, &haikuStruct )
	}
	return &items
}

func itemsToRss(items *[]*Haiku) *string {
	const hiddenHaikuUrl = "http://www.ft.com/hidden-haiku"
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

	for _, item := range *items {
		guid       := item.Url + "#" + GetMD5Hash(item.Text)
		created, _ := time.Parse("2006-01-02", item.DateSelected)

		feed.Items = append(feed.Items, &Item{
			Title:       item.Title,
			Link:        &Link{Href: item.Url},
			Description: item.Description,
			Created:     created,
			Id:          guid,
		})
	}

	rss, _ := feed.ToRss()
	return &rss
}

func Generate(maxItems int) *string {
	jsonBody := getJsonBody(jsonUrl)
	items := parseJsonToGenerateItems( jsonBody, maxItems )
	rssString := itemsToRss(items)
	return rssString
}

func GenerateItems(maxItems int) *[]*Haiku {
	jsonBody := getJsonBody(jsonUrl)
	items := parseJsonToGenerateItems( jsonBody, maxItems )
	return items
}

func main() {
	godotenv.Load()
	rssString := Generate(10)

	fmt.Println("main: rssString=", *rssString)
}