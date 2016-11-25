package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const longformPubDate = "2006-01-02T15:04:05Z" // needs to be this exact string, according to http://stackoverflow.com/questions/25845172/parsing-date-string-in-golang
const baseUriCapi = "http://api.ft.com/content/items/v1/"
const baseUriSapi = "http://api.ft.com/content/search/v1"
const newsFeedJsonUri = "http://www.ft-static.com/contentapi/live/latestNews.json"
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

func getCapiArticleJsonBody(uuid string) *[]byte {
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
		if pd, err := time.Parse(longformPubDate, pds); err == nil {
			pubDateTime = &pd
		} else {
			fmt.Println("WARNING: content.parsePubDateString: could not parse pubdate string: ", pds, ", where longformPubDate=", longformPubDate, "err=", err)
		}
	}

	return pubDateTime
}

func parseCapiArticleJsonBody(jsonBody *[]byte) *Article {

	var data interface{}
	json.Unmarshal(*jsonBody, &data)

	aSiteUrl := ""
	aUuid := ""
	aTitle := ""
	aAuthor := ""
	aBody := ""
	aPubDateString := ""
	aBrand := ""
	aGenre := ""
	var aPubDate *time.Time

	// look for article img, widest promo img, and widest non-promo img
	aArticleImgUrl     := ""
	aArticleImgWidth   := 0
	aArticleImgHeight  := 0
	aNonPromoImgUrl    := ""
	aNonPromoImgWidth  := 0
	aNonPromoImgHeight := 0
	aPromoImgUrl       := ""
	aPromoImgWidth     := 0
	aPromoImgHeight    := 0

	if item, ok := data.(map[string]interface{})[`item`].(map[string]interface{}); ok {
		if uuid, ok := item["id"].(string); ok {
			aUuid = uuid
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

			if location, ok := item["location"].(map[string]interface{}); ok {
				if uri, ok := location["uri"].(string); ok {
					aSiteUrl = uri
				}
			}

			if editorial, ok := item["editorial"].(map[string]interface{}); ok {
				if byline, ok := editorial["byline"].(string); ok {
					aAuthor = byline
				}
			}

			if metadata, ok := item["metadata"].(map[string]interface{}); ok {
				if brandItems, ok := metadata["brand"].([]interface{}); ok {
					if len(brandItems) > 0 {
						if term, ok := brandItems[0].(map[string]interface{})["term"].(map[string]interface{}); ok {
							if name, ok := term["name"].(string); ok {
								aBrand = name
							}
						}
					}
				}
				if genreItems, ok := metadata["genre"].([]interface{}); ok {
					if len(genreItems) > 0 {
						if term, ok := genreItems[0].(map[string]interface{})["term"].(map[string]interface{}); ok {
							if name, ok := term["name"].(string); ok {
								aGenre = name
							}
						}
					}
				}
			}

			if images, ok := item["images"].([]interface{}); ok {
				if len(images) > 0 {
					fmt.Println("content: parseCapiArticleJsonBody: found images, len > 0")

					for _, image := range images {
						if imageType, ok := image.(map[string]interface{})["type"].(string); ok {
							if imageUrl, ok := image.(map[string]interface{})["url"].(string); ok {
								if imageWidthFloat64, ok := image.(map[string]interface{})["width"].(float64); ok {
									imageWidth := int(imageWidthFloat64)
									if imageHeightFloat64, ok := image.(map[string]interface{})["height"].(float64); ok {
										imageHeight := int(imageHeightFloat64)
										if imageType == "article" {
											aArticleImgWidth  = imageWidth
											aArticleImgHeight = imageHeight
											aArticleImgUrl    = imageUrl										
										}
										if imageType == "promo" {
											if imageWidth > aPromoImgWidth {
												aPromoImgWidth  = imageWidth
												aPromoImgHeight = imageHeight
												aPromoImgUrl    = imageUrl
											}
										} else {
											if imageWidth > aNonPromoImgWidth {
												aNonPromoImgWidth  = imageWidth
												aNonPromoImgHeight = imageHeight
												aNonPromoImgUrl    = imageUrl
											}
										}
									}
								}
							}
						}
					}
					if aArticleImgUrl == "" {
						aArticleImgWidth  = aNonPromoImgWidth
						aArticleImgHeight = aNonPromoImgHeight
						aArticleImgUrl    = aNonPromoImgUrl										
					}
				}
			}

			if aAuthor == "" {
				if aBrand != "" {
					aAuthor = aBrand
				} else {
					aAuthor = aGenre
				}
			}
		}
	}

	aPubDate = parsePubDateString(aPubDateString)

	article := Article{
		SiteUrl:       aSiteUrl,
		Uuid:          aUuid,
		Title:         aTitle,
		Author:        aAuthor,
		Body:          aBody,
		PubDateString: aPubDateString,
		PubDate:       aPubDate,
		ImageUrl:      aArticleImgUrl,
		ImageWidth:    aArticleImgWidth,
		ImageHeight:   aArticleImgHeight,
		PromoImageUrl:    aPromoImgUrl,
		PromoImageWidth:  aPromoImgWidth,
		PromoImageHeight: aPromoImgHeight,
		NonPromoImageUrl:    aNonPromoImgUrl,
		NonPromoImageWidth:  aNonPromoImgWidth,
		NonPromoImageHeight: aNonPromoImgHeight,
	}

	fmt.Println("content: parseCapiArticleJsonBody: Uuid=", aUuid, ", ImageUrl=", aArticleImgUrl)

	return &article
}

var uuidJsonBodyCache = map[string]*[]byte{}

func GetArticle(uuid string) *Article {
	var jsonBody *[]byte

	if _, ok := uuidJsonBodyCache[uuid]; ok {
		fmt.Println("content.GetArticle: cache hit: uuid=", uuid)
		jsonBody = uuidJsonBodyCache[uuid]
	} else {
		fmt.Println("content.GetArticle: cache miss: uuid=", uuid)
		jsonBody = getCapiArticleJsonBody(uuid)
		uuidJsonBodyCache[uuid] = jsonBody
	}

	article := parseCapiArticleJsonBody(jsonBody)

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

func constructQueryString(sr *SearchRequest) string {

	var queryString string

	// does not yet handle "page"
	switch sr.QueryType {
	case "keyword", "":
		queryString = sr.QueryText
	case "title-only":
		queryString = "title" + `:\"` + sr.QueryText + `\"`
	case "before":
		if sr.QueryText == "" || sr.QueryText == "now" {
			queryString = ""
		} else {
			queryString = "lastPublishDateTime:<" + sr.QueryText
		}
	default:
		queryString = sr.QueryType + `:\"` + sr.QueryText + `\"`
	}

	return queryString
}

func convertStringsToQuotedCSV(sList []string) string {
	sListQuoted := []string{}

	for _, s := range sList {
		sListQuoted = append(sListQuoted, `"`+s+`"`)
	}

	sCsv := strings.Join(sListQuoted, ", ")
	return sCsv
}

var stringJsonBodyCache = map[string]*[]byte{}

func getSapiResponseJsonBody(queryString string, maxResults int) *[]byte {
	curationsString := convertStringsToQuotedCSV([]string{"ARTICLES", "BLOGS"})
	aspectsString := convertStringsToQuotedCSV([]string{"title", "location", "summary", "lifecycle", "metadata", "editorial"})

	jsonStr := []byte(
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
			`}`)

	var jsonBody *[]byte
	jsonStrAsKey := string(jsonStr[:])
	if _, ok := stringJsonBodyCache[jsonStrAsKey]; ok {
		fmt.Println("content.getSapiResponseJsonBody: cache hit: jsonStrAsKey=", jsonStrAsKey)
		jsonBody = stringJsonBodyCache[jsonStrAsKey]
	} else {
		fmt.Println("content.getSapiResponseJsonBody: cache miss: jsonStrAsKey=", jsonStrAsKey)
		jsonBody = constructSapiResponseJsonBody(&jsonStr)
		stringJsonBodyCache[jsonStrAsKey] = jsonBody
	}

	return jsonBody
}

func constructSapiResponseJsonBody(jsonStr *[]byte) *[]byte {
	url := baseUriSapi + "?apiKey=" + apiKey

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(*jsonStr))
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

	return &jsonBody
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
	SiteUrl       string
	Uuid          string
	Title         string
	Author        string
	Excerpt       string
	Body          string
	PubDateString string
	PubDate       *time.Time
	ImageUrl    string
	ImageWidth  int
	ImageHeight int
	NonPromoImageUrl    string
	NonPromoImageWidth  int
	NonPromoImageHeight int
	PromoImageUrl    string
	PromoImageWidth  int
	PromoImageHeight int
}

func parseSapiResponseJsonBody(jsonBody *[]byte, sReq *SearchRequest, queryString string) *SearchResponse {

	siteUrl := "http://www.ft.com"
	siteSearchUrl := "http://search.ft.com/search?queryText=" + strings.Replace(queryString, `\"`, `"`, -1)
	numPossible := 0
	articles := []*Article{}

	// locate results
	var data interface{}
	json.Unmarshal(*jsonBody, &data)
	if outerResults, ok := data.(map[string]interface{})["results"].([]interface{}); ok {
		if results0, ok := outerResults[0].(map[string]interface{}); ok {
			if indexCount, ok := results0["indexCount"]; ok {
				numPossible = int(indexCount.(float64))
			}
			if innerResults, ok := results0["results"].([]interface{}); ok {
				for _, r := range innerResults {
					siteUrl := ""
					uuid := ""
					title := ""
					author := ""
					excerpt := ""
					pubDateString := ""
					brand := ""
					genre := ""
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

					if metadata, ok := r.(map[string]interface{})["metadata"].(map[string]interface{}); ok {
						if brandItems, ok := metadata["brand"].([]interface{}); ok {
							if len(brandItems) > 0 {
								if term, ok := brandItems[0].(map[string]interface{})["term"].(map[string]interface{}); ok {
									if name, ok := term["name"].(string); ok {
										brand = name
									}
								}
							}
						}
						if genreItems, ok := metadata["genre"].([]interface{}); ok {
							if len(genreItems) > 0 {
								if term, ok := genreItems[0].(map[string]interface{})["term"].(map[string]interface{}); ok {
									if name, ok := term["name"].(string); ok {
										genre = name
									}
								}
							}
						}
					}

					if author == "" {
						if brand != "" {
							author = brand
						} else {
							author = genre
						}
					}

					var body string
					if sReq.QueryType == "keyword" || sReq.QueryType == "" {
						body = excerpt
					} else {
						body = title
					}

					article := Article{
						SiteUrl:       siteUrl,
						Uuid:          uuid,
						Title:         title,
						Author:        author,
						Excerpt:       excerpt,
						Body:          body,
						PubDateString: pubDateString,
						PubDate:       pubDateTime,
					}

					articles = append(articles, &article)
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

func lookupCapiArticles(sRequest *SearchRequest, sResponse *SearchResponse, startTiming time.Time) *[]*Article {
	maxDurationNanoseconds := int64(sRequest.MaxDurationMillis * 1e6)
	capiArticles := []*Article{}

	if sRequest.MaxArticles > 0 {
		for i, sapiA := range *(sResponse.Articles) {
			articleLookupStartTiming := time.Now()
			capiA := GetArticle(sapiA.Uuid)
			capiArticles = append(capiArticles, capiA)
			articleLookupDuration := time.Since(articleLookupStartTiming).Nanoseconds()
			fmt.Println("content.lookupCapiArticles: articleLookupDuration=", (articleLookupDuration / 1000000))
			durationNanoseconds := time.Since(startTiming).Nanoseconds()
			if i > sRequest.MaxArticles {
				break
			}
			if durationNanoseconds > maxDurationNanoseconds {
				fmt.Println("content.lookupCapiArticles: curtailing CAPI lookups: durationMillis=", durationNanoseconds/1e6)
				break
			}
		}
	}

	return &capiArticles
}

func constructArticlesFromSearchResults(sRequest *SearchRequest, sResponse *SearchResponse) *[]*Article {
	articles := []*Article{}

	if sRequest.MaxArticles > 0 {
		for _, sapiA := range *(sResponse.Articles) {
			articles = append(articles, sapiA)
		}
	}

	return &articles
}

func Search(sRequest *SearchRequest) *SearchResponse {
	startTiming := time.Now()

	fmt.Println("content.Search: sRequest.QueryType==page")

	var sResponse *SearchResponse
	if sRequest.QueryType == "pages" {
		webUrl := sRequest.QueryText
		if webUrl == "http://www.ft.com/news-feed" {
			jsonBody := constructGetResponseJsonBody(newsFeedJsonUri)
			sResponse = parseNewsFeedContentJsonBody(jsonBody, sRequest, webUrl)
		} else {
			pageId := getPageIdByWebUrl(webUrl)
			jsonBody := constructMainContentJsonBodyFromId(pageId)
			sResponse = parseMainContentJsonBody(jsonBody, sRequest, webUrl)
		}
	} else {
		queryString := constructQueryString(sRequest)
		jsonBody := getSapiResponseJsonBody(queryString, sRequest.MaxArticles)
		sResponse = parseSapiResponseJsonBody(jsonBody, sRequest, queryString)
	}

	fmt.Println("content.Search: found sResponse.NumArticles=", sResponse.NumArticles)

	var articles *[]*Article
	if !sRequest.SearchOnly {
		articles = lookupCapiArticles(sRequest, sResponse, startTiming)
	} else {
		articles = constructArticlesFromSearchResults(sRequest, sResponse)
	}

	sResponse.SetArticles(articles)

	fmt.Println("content.Search: fleshed out len(articles)=", len(*articles))

	return sResponse
}

func constructGetResponseJsonBody(url string) *[]byte {
	urlWithKey := url + "?apiKey=" + apiKey

	req, err := http.NewRequest("GET", urlWithKey, nil)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		fmt.Println("WARNING: content: constructGetResponseJsonBody: response Status:", resp.Status, ", url=", url)
	}

	jsonBody, _ := ioutil.ReadAll(resp.Body)

	return &jsonBody
}

func constructAllPagesJsonBody() *[]byte {
	return constructGetResponseJsonBody("http://api.ft.com/site/v1/pages")
}

func parseAllPagesJsonBody(jsonBody *[]byte) *map[string]string {

	mapWebUrlToId := map[string]string{}

	// locate results
	var data interface{}
	json.Unmarshal(*jsonBody, &data)

	if pages, ok := data.(map[string]interface{})["pages"].([]interface{}); ok {
		for _, p := range pages {
			id := ""
			webUrl := ""

			if _, ok := p.(map[string]interface{})["id"]; ok {
				id = p.(map[string]interface{})["id"].(string)
			}

			if _, ok := p.(map[string]interface{})["webUrl"]; ok {
				webUrl = p.(map[string]interface{})["webUrl"].(string)
			}

			mapWebUrlToId[webUrl] = id
		}
	}
	return &mapWebUrlToId
}

var allKnownPageIdsByWebUrl *map[string]string

func getAllPages() *map[string]string {
	if allKnownPageIdsByWebUrl == nil {
		jsonBody := constructAllPagesJsonBody()
		allKnownPageIdsByWebUrl = parseAllPagesJsonBody(jsonBody)
		fmt.Println("content.getAllPages: len(allKnownPageIdsByWebUrl)=", len(*allKnownPageIdsByWebUrl))
	}

	return allKnownPageIdsByWebUrl
}

func getPageIdByWebUrl(webUrl string) string {
	return (*getAllPages())[webUrl]
}

func constructMainContentJsonBodyFromId(id string) *[]byte {
	url := "http://api.ft.com/site/v1/pages/" + id + "/main-content"
	return constructGetResponseJsonBody(url)
}

func parseMainContentJsonBody(jsonBody *[]byte, sReq *SearchRequest, webUrl string) *SearchResponse {

	siteSearchUrl := "http://search.ft.com/"
	numPossible := 0
	articles := []*Article{}

	// locate results
	var data interface{}
	json.Unmarshal(*jsonBody, &data)
	if pageItems, ok := data.(map[string]interface{})["pageItems"].([]interface{}); ok {
		for _, r := range pageItems {
			siteUrl := ""
			uuid := ""
			title := ""
			author := ""
			pubDateString := ""
			var pubDateTime *time.Time

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

			body := title
			excerpt := title

			article := Article{
				SiteUrl:       siteUrl,
				Uuid:          uuid,
				Title:         title,
				Author:        author,
				Excerpt:       excerpt,
				Body:          body,
				PubDateString: pubDateString,
				PubDate:       pubDateTime,
			}

			articles = append(articles, &article)
		}
	}

	searchResponse := SearchResponse{
		SiteUrl:       webUrl,
		SiteSearchUrl: siteSearchUrl,
		NumArticles:   len(articles),
		NumPossible:   numPossible,
		Articles:      &articles,
		QueryString:   "",
		SearchRequest: sReq,
	}

	return &searchResponse
}

func parseNewsFeedContentJsonBody(jsonBody *[]byte, sReq *SearchRequest, webUrl string) *SearchResponse {

	siteSearchUrl := "http://search.ft.com/"
	numPossible := 0
	articles := []*Article{}

	// locate results
	var data interface{}
	json.Unmarshal(*jsonBody, &data)
	if nfArticles, ok := data.(map[string]interface{})["articles"].([]interface{}); ok {
		for _, r := range nfArticles {
			siteUrl := ""
			uuid := ""
			title := ""
			author := ""
			pubDateString := ""
			var pubDateTime *time.Time

			if id, ok := r.(map[string]interface{})["id"].(string); ok {
				uuid = id
			}

			if titleString, ok := r.(map[string]interface{})["title"].(string); ok {
				title = titleString
			}

			if url, ok := r.(map[string]interface{})["url"].(string); ok {
				siteUrl = url
			}

			author = "Soz. No author in news-feed."

			if publishDate, ok := r.(map[string]interface{})["publishDate"].(string); ok {
				pubDateString = publishDate
				pubDateTime = parsePubDateString(pubDateString)
			}

			body := title
			excerpt := title

			article := Article{
				SiteUrl:       siteUrl,
				Uuid:          uuid,
				Title:         title,
				Author:        author,
				Excerpt:       excerpt,
				Body:          body,
				PubDateString: pubDateString,
				PubDate:       pubDateTime,
			}

			articles = append(articles, &article)
		}
	}

	searchResponse := SearchResponse{
		SiteUrl:       webUrl,
		SiteSearchUrl: siteSearchUrl,
		NumArticles:   len(articles),
		NumPossible:   numPossible,
		Articles:      &articles,
		QueryString:   "",
		SearchRequest: sReq,
	}

	return &searchResponse
}

func main() {
	godotenv.Load()
	uuid := "b57fee24-cb3c-11e5-be0b-b7ece4e953a0"

	article := GetArticle(uuid)
	fmt.Println("main: article.Title=", article.Title)

	sRequest := &SearchRequest{
		QueryType:         "keyword",    // e.g "keyword", "title", "topicXYZ", etc
		QueryText:         "a bit of a", // e.g. "tail spin" or "\"tail spin\""
		MaxArticles:       10,
		MaxDurationMillis: 3000,
		SearchOnly:        true, // i.e. don't bother looking up articles
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

	for i, a := range *(sResponse.Articles) {
		fmt.Println(i, " - ", a.Uuid, ", ", a.Title, ", by ", a.Author)
	}

	sRequest = &SearchRequest{
		QueryType:         "keyword",    // e.g "keyword", "title", "topicXYZ", etc
		QueryText:         "a bit of a", // e.g. "tail spin" or "\"tail spin\""
		MaxArticles:       100,
		MaxDurationMillis: 3000,
		SearchOnly:        false, // i.e.do lookup CAPI
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

	for i, a := range *(sResponse.Articles) {
		fmt.Println(i, " - ", a.Uuid, ", ", a.Title, ", by ", a.Author, ", body=", a.Body[0:20])
	}
}
