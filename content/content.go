package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
    "github.com/railsagainstignorance/alignment/Godeps/_workspace/src/github.com/joho/godotenv"
    "time"
    "strconv"
)

const longformPubDate     = "2006-01-02T15:04:05Z" // needs to be this exact string, according to http://stackoverflow.com/questions/25845172/parsing-date-string-in-golang
const baseUriCapi         = "http://api.ft.com/content/items/v1/"
const baseUriSapi         = "http://api.ft.com/content/search/v1"
const sapiKeyEnvParamName = "SAPI_KEY"

func getApiKey() string {
    godotenv.Load()
    key := os.Getenv(sapiKeyEnvParamName)

    if key == "" {
        fmt.Println("content: getApiKey: no such env params: ", sapiKeyEnvParamName)
    }

    return key
}

var apiKey = getApiKey()

func getCapiArticleJsonBody(uuid string) (*[]byte) {
    url := baseUriCapi + uuid + "?apiKey=" + apiKey
    fmt.Println("content: getCapiArticleJsonBody: uuid=", uuid)

    req, err := http.NewRequest("GET", url, nil)
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    fmt.Println("content: getCapiArticleJsonBody: response Status:", resp.Status)
    jsonBody, _ := ioutil.ReadAll(resp.Body)

    return &jsonBody
}

func parsePubDateString(pds string) *time.Time {
    var pubDateTime *time.Time
    if pds != "" {
        if pd,err := time.Parse(longformPubDate, pds); err == nil {
            pubDateTime = &pd
        } else {
            fmt.Println("WARNING: content.parsePubDateString: could not parse pubdate string: ", pds, ", where longformPubDate=", longformPubDate, "err=", err)
        }
    }

    return pubDateTime
}

func parseCapiArticleJsonBody(jsonBody *[]byte) (*Article) {

    var data interface{}
    json.Unmarshal(*jsonBody, &data)

    aSiteUrl := ""
    aUuid    := ""
    aTitle   := ""
    aAuthor  := ""
    aBody    := ""
    aPubDateString := ""
    var aPubDate *time.Time

    if item, ok := data.(map[string]interface{})[`item`].(map[string]interface{}); ok {
        if uuid, ok := item["id"].(string); ok {
            aUuid    = uuid
            aSiteUrl = "http://www.ft.com/cms/s/2/" + uuid + ".html"

            if titleOuter, ok := item["title"].(map[string]interface{}); ok {
                if title, ok := titleOuter["title"].(string); ok {
                    aTitle = title
                }
            }

            if bodyOuter, ok := item["body"].(map[string]interface{}); ok {
                if body, ok := bodyOuter["body"].(string); ok {
                    aBody = body
                }
            }

            if lifecycle, ok := item["lifecycle"].(map[string]interface{}); ok {
                if lastPublishDateTime, ok := lifecycle["lastPublishDateTime"].(string); ok {
                    aPubDateString = lastPublishDateTime
                }
            }

            if editorial, ok := item["editorial"].(map[string]interface{}); ok {
                if byline, ok := editorial["byline"].(string); ok {
                    aAuthor = byline
                }
            }
        }
    }

    aPubDate = parsePubDateString( aPubDateString )

    article := Article{
        SiteUrl: aSiteUrl,
        Uuid:    aUuid,
        Title:   aTitle,
        Author:  aAuthor,
        Body:    aBody,
        PubDateString: aPubDateString,
        PubDate: aPubDate,
    }

    return &article
}

var uuidJsonBodyCache = map[string]*[]byte{}

func GetArticle(uuid string) (*Article) {
    var jsonBody *[]byte

    if _, ok := uuidJsonBodyCache[uuid]; ok {
        fmt.Println("content.GetArticle: cache hit: uuid=", uuid)
        jsonBody = uuidJsonBodyCache[uuid]
    } else {
        fmt.Println("content.GetArticle: cache miss: uuid=", uuid)
        jsonBody = getCapiArticleJsonBody(uuid)
        uuidJsonBodyCache[uuid] = jsonBody
    }

    article := parseCapiArticleJsonBody( jsonBody )

    return article
}

// now same for SAPI stuff
type SearchRequest struct {
    QueryType         string // e.g "keyword", "title", "topicXYZ", etc
    QueryText         string // e.g. "tail spin" or "\"tail spin\""
    MaxArticles       int
    MaxDurationMillis int
    SearchOnly        bool // i.e. don't bother looking up articles
    QueryStringValue  string
}

func constructQueryString(sr *SearchRequest ) string {

    var queryString string

    // does not yet handle "page"
    switch sr.QueryType {
    case "keyword", "":
        queryString = sr.QueryText
    case "title-only":
        queryString = "title" + `:\"` + sr.QueryText + `\"`
    default:
        queryString = sr.QueryType + `:\"` + sr.QueryText + `\"`
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

func getSapiResponseJsonBody(queryString string, maxResults int) ([]byte) {
    url := "http://api.ft.com/content/search/v1?apiKey=" + apiKey

    // fmt.Println("sapi: getSapiResponseJsonBody: queryString:", queryString)
    curationsString := convertStringsToQuotedCSV( []string{ "ARTICLES", "BLOGS" } )
    aspectsString   := convertStringsToQuotedCSV( []string{ "title", "location", "summary", "lifecycle", "metadata", "editorial" } )

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

    if resp.Status != "200 OK" {
        fmt.Println("WARNING: content: getSapiResponseJsonBody: response Status:", resp.Status)
    }

    jsonBody, _ := ioutil.ReadAll(resp.Body)

    return jsonBody
}

type SearchResponse struct {
    SiteUrl       string
    SiteSearchUrl string
    NumArticles   int
    NumPossible   int
    Articles      *[]*Article
    QueryString   string
    SearchRequest *SearchRequest
}

func (r *SearchResponse) SetArticles(articles *[]*Article) {
    r.Articles = articles
    r.NumArticles = len(*articles)
}

type Article struct {
    SiteUrl string
    Uuid    string
    Title   string
    Author  string
    Excerpt string
    Body    string
    PubDateString string
    PubDate *time.Time
}

func parseSapiResponseJsonBody(jsonBody []byte, sReq *SearchRequest, queryString string) *SearchResponse {

    siteUrl       := "http://www.ft.com"
    siteSearchUrl := "http://search.ft.com/search?queryText=" + strings.Replace(queryString, `\"`, `"`, -1)
    numPossible   := 0
    articles      := []*Article{}

    // locate results
    var data interface{}
    json.Unmarshal(jsonBody, &data)
    if outerResults, ok := data.(map[string]interface{})["results"].([]interface{}); ok {
        if results0, ok := outerResults[0].(map[string]interface{}); ok {
            if indexCount, ok := results0["indexCount"]; ok {
                numPossible = int(indexCount.(float64))
            }
            if innerResults, ok := results0["results"].([]interface{}); ok {
                for _, r := range innerResults {
                    siteUrl := ""
                    uuid    := ""
                    title   := ""
                    author  := ""
                    excerpt := ""
                    pubDateString := ""
                    var pubDateTime *time.Time

                    if summary, ok := r.(map[string]interface{})["summary"].(map[string]interface{}); ok {
                        if excerptString, ok := summary["excerpt"].(string); ok {
                            excerpt = excerptString                      
                        }
                    }

                    if id, ok := r.(map[string]interface{})["id"].(string); ok {
                        uuid = id
                    }

                    if titleOuter, ok := r.(map[string]interface{})["title"].(map[string]interface{}); ok {
                        if titleString, ok := titleOuter["title"].(string); ok {
                            title = titleString
                        }
                    }

                    if location, ok := r.(map[string]interface{})["location"].(map[string]interface{}); ok {
                        if locationUri, ok := location["uri"].(string); ok {
                            siteUrl = locationUri
                        }
                    }

                    if editorial, ok := r.(map[string]interface{})["editorial"].(map[string]interface{}); ok {
                        if byline, ok := editorial["byline"].(string); ok {
                            author = byline
                        }
                    }

                    if lifecycle, ok := r.(map[string]interface{})["lifecycle"].(map[string]interface{}); ok {
                        if lastPublishDateTimeString, ok := lifecycle["lastPublishDateTime"].(string); ok {
                            pubDateString = lastPublishDateTimeString
                            pubDateTime = parsePubDateString(pubDateString)
                        }
                    }

                    var body string
                    if sReq.QueryType == "keyword" || sReq.QueryType == "" {
                        body = excerpt
                    } else {
                        body = title
                    }

                    article := Article{
                        SiteUrl: siteUrl,
                        Uuid:    uuid,
                        Title:   title,
                        Author:  author,
                        Excerpt: excerpt,
                        Body:    body,
                        PubDateString: pubDateString,
                        PubDate:       pubDateTime,
                    }

                    articles = append( articles, &article )
                }
            }
        }
    }

    searchResponse := SearchResponse{
        SiteUrl:       siteUrl,
        SiteSearchUrl: siteSearchUrl,
        NumArticles:   len(articles),
        NumPossible:   numPossible,
        Articles:      &articles,
        QueryString:   queryString,
        SearchRequest: sReq,
    }

    return &searchResponse
}

func lookupCapiArticles( sRequest *SearchRequest, sResponse *SearchResponse, startTiming time.Time ) *[]*Article {
    maxDurationNanoseconds := int64(sRequest.MaxDurationMillis * 1e6)
    capiArticles := []*Article{}

    if sRequest.MaxArticles > 0 {
        for i, sapiA := range *(sResponse.Articles) {
            capiA := GetArticle(sapiA.Uuid) 
            capiArticles = append( capiArticles, capiA )
            durationNanoseconds := time.Since(startTiming).Nanoseconds()
            if i > sRequest.MaxArticles {
                break
            }
            if durationNanoseconds > maxDurationNanoseconds {
                fmt.Println("content.lookupCapiArticles: curtailing CAPI lookups: durationMillis=", durationNanoseconds / 1e6)
                break
            }
        }
    }

    return &capiArticles
}

func constructArticlesFromSearchResults( sRequest *SearchRequest, sResponse *SearchResponse ) *[]*Article {
    articles := []*Article{}

    if sRequest.MaxArticles > 0 {
        for _, sapiA := range *(sResponse.Articles) {
            articles = append( articles, sapiA )
        }
    }

    return &articles
}

func Search(sRequest *SearchRequest) *SearchResponse {
    startTiming := time.Now()

    queryString       := constructQueryString( sRequest )
    jsonBody          := getSapiResponseJsonBody(queryString, sRequest.MaxArticles)
    sResponse         := parseSapiResponseJsonBody(jsonBody, sRequest, queryString)

    var articles *[]*Article
    if !sRequest.SearchOnly {
        articles = lookupCapiArticles(sRequest, sResponse, startTiming)
    } else {
        articles = constructArticlesFromSearchResults( sRequest, sResponse )
    }

    sResponse.SetArticles(articles)

    return sResponse
}

func main() {
    godotenv.Load()
    uuid := "b57fee24-cb3c-11e5-be0b-b7ece4e953a0"

    article := GetArticle( uuid )
    fmt.Println("main: article.Title=", article.Title)

    sRequest := &SearchRequest {
        QueryType: "keyword", // e.g "keyword", "title", "topicXYZ", etc
        QueryText: "a bit of a", // e.g. "tail spin" or "\"tail spin\""
        MaxArticles: 10,
        MaxDurationMillis: 3000,
        SearchOnly: true, // i.e. don't bother looking up articles
    }

    fmt.Println("sRequest=", sRequest)
    sResponse := Search(sRequest)

    fmt.Println("sResponse:", 
        "\nSiteSearchUrl=", sResponse.SiteSearchUrl,
        "\nNumArticles=", sResponse.NumArticles,
        "\nNumPossible=", sResponse.NumPossible,
        "\nQueryString=", sResponse.QueryString,
        "\nSearchRequest=", sResponse.SearchRequest,
        "\nArticles:\n",
        )

    for i,a := range *(sResponse.Articles) {
        fmt.Println(i, " - ", a.Uuid, ", ", a.Title, ", by ", a.Author)
    }

    sRequest = &SearchRequest {
        QueryType: "keyword", // e.g "keyword", "title", "topicXYZ", etc
        QueryText: "a bit of a", // e.g. "tail spin" or "\"tail spin\""
        MaxArticles: 100,
        MaxDurationMillis: 3000,
        SearchOnly: false, // i.e.do lookup CAPI
    }

    fmt.Println("sRequest=", sRequest)
    sResponse = Search(sRequest)

    fmt.Println("sResponse:", 
        "\nSiteSearchUrl=", sResponse.SiteSearchUrl,
        "\nNumArticles=", sResponse.NumArticles,
        "\nNumPossible=", sResponse.NumPossible,
        "\nQueryString=", sResponse.QueryString,
        "\nSearchRequest=", sResponse.SearchRequest,
        "\nArticles:\n",
        )

    for i,a := range *(sResponse.Articles) {
        fmt.Println(i, " - ", a.Uuid, ", ", a.Title, ", by ", a.Author, ", body=", a.Body[0:20])
    }
}
