package capi

import (
	// "bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	// "strings"
    "github.com/railsagainstignorance/alignment/Godeps/_workspace/src/github.com/joho/godotenv"
)

func getCapiArticleJsonBody(uuid string) ([]byte) {
    apiKey := os.Getenv("SAPI_KEY")
    url := "http://api.ft.com/content/items/v1/" + uuid + "?apiKey=" + apiKey
    fmt.Println("capi: getCapiArticleJsonBody: url=", url)

    req, err := http.NewRequest("GET", url, nil)
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    fmt.Println("response Status:", resp.Status)
    // fmt.Println("response Headers:", resp.Header)
    jsonBody, _ := ioutil.ReadAll(resp.Body)

    return jsonBody
}

type Article struct {
    Uuid  string
    Url   string
    Title string
    Body  string
}

func parseCapiArticleJsonBody(jsonBody []byte) (*Article) {

    var data interface{}
    json.Unmarshal(jsonBody, &data)

    article := Article {
        Uuid:  "",
        Url:   "",
        Title: "",
        Body:  "",
    }

    if item, ok := data.(map[string]interface{})[`item`].(map[string]interface{}); ok {
        if uuid, ok := item["id"].(string); ok {
            article.Uuid = uuid
            article.Url  = "http://www.ft.com/cms/s/2/" + uuid + ".html"

            if titleOuter, ok := item["title"].(map[string]interface{}); ok {
                if title, ok := titleOuter["title"].(string); ok {
                    article.Title = title
                }
            }

            if bodyOuter, ok := item["body"].(map[string]interface{}); ok {
                if body, ok := bodyOuter["body"].(string); ok {
                    article.Body = body
                }
            }
        }
    }

    return &article
}

func GetArticle(uuid string) (*Article) {
    jsonBody := getCapiArticleJsonBody(uuid)
    article := parseCapiArticleJsonBody( jsonBody )

    return article
}

func main() {
    godotenv.Load()
    uuid := "b57fee24-cb3c-11e5-be0b-b7ece4e953a0"

    article := GetArticle( uuid )
    fmt.Println("main: article.Title=", article.Title)
}
