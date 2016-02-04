package sapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func getSapiResponseJsonBody(text string, titleOnly bool) ([]byte, string) {
	sapiKey := os.Getenv("SAPI_KEY")
	url := "http://api.ft.com/content/search/v1?apiKey=" + sapiKey
	queryString := `\"` + text + `\"`
	if titleOnly {
		queryString = "title:" + queryString
	}

    fmt.Println("queryString:", queryString)

	var jsonStr = []byte(`{"queryString": "` + queryString + `","queryContext" : {"curations" : [ "ARTICLES", "BLOGS" ]},  "resultContext" : {"maxResults" : "100", "offset" : "0", "aspects" : [ "title", "location", "summary", "lifecycle", "metadata"], "sortOrder": "DESC", "sortField": "lastPublishDateTime" } } }`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
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

	return jsonBody, strings.Replace(queryString, `\"`, `"`, -1)
}

type SearchParams struct {
    Text   string
    Source string
}

type ResultItem struct {
    Phrase      string
    Excerpt     string
    Title       string
    LocationUri string
}

type SearchResult struct {
    Text      string
    Source    string
    FtcomUrl       string
    FtcomSearchUrl string
    TitleOnlyChecked string
    AnyChecked       string
    IndexCount int
    Items      *[]*ResultItem
}

func Search(params SearchParams) *SearchResult {
    defaultText := os.Getenv("DEFAULT_TEXT")
    if defaultText=="" {
        defaultText = "has its own"
    }
	text := params.Text
    if text=="" {
        text = defaultText
    }
	source := params.Source
    if source=="" {
        source="any"
    }
	titleOnly := source == "title-only"

	jsonBody, queryString := getSapiResponseJsonBody(text, titleOnly)

	// locate results
	var data interface{}
	json.Unmarshal(jsonBody, &data)
	outerResults := data.(map[string]interface{})[`results`].([]interface{})[0].(map[string]interface{})

    indexCount := int(outerResults["indexCount"].(float64))
    fmt.Println("indexCount:", indexCount)

    var (
        items []*ResultItem = []*ResultItem{}
    )

    if innerResults, ok := outerResults[`results`].([]interface{}); ok {

    	// loop over results to pick out relevant fields
        knownPhrases := map[string]string{}

    	for _, r := range innerResults {
            summary := r.(map[string]interface{})["summary"].(map[string]interface{})
            var excerpt string = ""
            if summary != nil {
                if _, ok := summary["excerpt"].(string); ok {
                    excerpt = summary["excerpt"].(string)                           
                }
            }

    		title1  := r.(map[string]interface{})["title"].(map[string]interface{})
            title   := title1["title"].(string)
            
            location    := r.(map[string]interface{})["location"].(map[string]interface{})
    		locationUri := location["uri"].(string)

    		phrase := excerpt
    		if titleOnly {
    			phrase = title
    		}

            if _, ok := knownPhrases[phrase]; !ok {
                item := ResultItem{
                    Phrase:      phrase,
                    Excerpt:     excerpt,
                    Title:       title,
                    LocationUri: locationUri,
                }

                items = append( items, &item )
                knownPhrases[phrase] = phrase               
            }
    	}
    } 

    var (
        titleOnlyChecked string = ""
        anyChecked       string = ""
    )

    if titleOnly {
        titleOnlyChecked = "checked"
    } else {
        anyChecked       = "checked"
    }

	sr := SearchResult{
        Text:             text, 
        Source:           source, 
        FtcomUrl:         "http://www.ft.com",
        FtcomSearchUrl:   "http://search.ft.com/search?queryText=" + queryString,
        TitleOnlyChecked: titleOnlyChecked,
        AnyChecked:       anyChecked,
        IndexCount:       indexCount,
        Items:            &items,
    }

    return &sr
}
