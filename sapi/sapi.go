package sapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
    "strconv"
    "github.com/railsagainstignorance/alignment/Godeps/_workspace/src/github.com/joho/godotenv"
)

// priority for constructing queryString:
// Author (if not nil), then title (if Source=="title-only"), then exact text 
type SearchParams struct {
    Text   string
    Source string
    Author string
}

func constructQueryString(params SearchParams) string {

    var queryString string

    if params.Author != "" {
        queryString = `authors:\"` + params.Author + `\"`
    } else {
        text := params.Text
        if text == "" {
            defaultText := os.Getenv("DEFAULT_TEXT")
            if defaultText=="" {
                defaultText = "has its own"
            }

            text = defaultText
        }

        if params.Source == "title-only" {
            queryString = `title:\"` + text + `\"`
        } else {
            queryString = `\"` + text + `\"`
        }
    }

    return queryString
}

func convertStringsToQuotedCSV( sList []string ) string {
    sListQuoted := []string{}
    for _,s := range sList {
        sListQuoted = append( sListQuoted, `"` + s + `"`)
    }

    sCsv := strings.Join( sListQuoted, ", " )
    return sCsv
}

func getSapiResponseJsonBody(queryString string) ([]byte) {
	sapiKey := os.Getenv("SAPI_KEY")
	url     := "http://api.ft.com/content/search/v1?apiKey=" + sapiKey

    fmt.Println("sapi: getSapiResponseJsonBody: queryString:", queryString)
    curationsString := convertStringsToQuotedCSV( []string{ "ARTICLES", "BLOGS" } )
    aspectsString   := convertStringsToQuotedCSV( []string{ "title", "location", "summary", "lifecycle", "metadata" } )
    maxResults      := 100

	jsonStr         := []byte(
        `{` +
            `"queryString" : "` + queryString + `",` +
            `"queryContext" : {"curations" : [ ` + curationsString + ` ]},` +
            `"resultContext" : {` + 
                `"maxResults" : "` + strconv.Itoa(maxResults) + `",` + 
                `"offset" : "0",` + 
                `"aspects" : [ ` + aspectsString + `],` + 
                `"sortOrder": "DESC",` + 
                `"sortField": "lastPublishDateTime"` +
            `}` + 
        `}` )

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("sapi: getSapiResponseJsonBody: response Status:", resp.Status)
	jsonBody, _ := ioutil.ReadAll(resp.Body)

	return jsonBody
}

type ResultItem struct {
    Phrase      string
    Excerpt     string
    Title       string
    LocationUri string
    Id          string
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

func keepAZ(r rune) rune { if r>='A' && r<='Z' {return r} else {return -1} }
func KeepAZString( s string ) string {return strings.Map(keepAZ, s)}

func Search(params SearchParams) *SearchResult {
    queryString := constructQueryString( params )
	jsonBody    := getSapiResponseJsonBody(queryString)
    titleOnly   := (params.Source == "title-only")

	// locate results
	var data interface{}
	json.Unmarshal(jsonBody, &data)
	outerResults := data.(map[string]interface{})[`results`].([]interface{})[0].(map[string]interface{})

    indexCount := int(outerResults["indexCount"].(float64))
    fmt.Println("sapi: Search: indexCount:", indexCount)

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

            id := r.(map[string]interface{})["id"].(string)

    		title1  := r.(map[string]interface{})["title"].(map[string]interface{})
            title   := title1["title"].(string)
            
            location    := r.(map[string]interface{})["location"].(map[string]interface{})
    		locationUri := location["uri"].(string)

    		phrase := excerpt
    		if titleOnly {
    			phrase = title
    		}

            phraseAZ := KeepAZString(phrase)

            if _, ok := knownPhrases[phraseAZ]; !ok {
                item := ResultItem{
                    Phrase:      phrase,
                    Excerpt:     excerpt,
                    Title:       title,
                    LocationUri: locationUri,
                    Id:          id,
                }

                items = append( items, &item )
                knownPhrases[phraseAZ] = phraseAZ               
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
        Text:             params.Text, 
        Source:           params.Source, 
        FtcomUrl:         "http://www.ft.com",
        FtcomSearchUrl:   "http://search.ft.com/search?queryText=" + queryString,
        TitleOnlyChecked: titleOnlyChecked,
        AnyChecked:       anyChecked,
        IndexCount:       indexCount,
        Items:            &items,
    }

    return &sr
}

func main() {
    godotenv.Load()

    sp := SearchParams{
        Text: "fish",
        Source: "title-only",
    }

    sr := Search( sp )

    fmt.Println("main:", "sp.Text=", sp.Text, ", sp.Source=", sp.Source)
    fmt.Println("main:", "sr.IndexCount=", sr.IndexCount)

}
