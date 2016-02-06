package main

import (
	"github.com/upthebuzzard/alignment/Godeps/_workspace/src/github.com/joho/godotenv"
	"html/template"
	"net/http"
	"os"
    "sort"
    "regexp"
    // "fmt"
    "github.com/upthebuzzard/alignment/align"
    "github.com/upthebuzzard/alignment/sapi"
    "github.com/upthebuzzard/alignment/rhyme"
)

func alignFormHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("align.html")
	t.Execute(w, nil)
}

func alignHandler(w http.ResponseWriter, r *http.Request) {
    searchParams := sapi.SearchParams{
        Text:   r.FormValue("text"),
        Source: r.FormValue("source"),
    }

	p := align.Search( searchParams )

	t, _ := template.ParseFiles("aligned.html")
	t.Execute(w, p)
}

type ResultItemWithRhymeAndMeter struct {
    ResultItem    *(sapi.ResultItem)
    RhymeAndMeter *(rhyme.RhymeAndMeter)
}

type SearchResultWithRhymeAndMeterList struct {
    SearchResult *(sapi.SearchResult)
    ResultItemsWithRhymeAndMeterList []*ResultItemWithRhymeAndMeter
    MatchMeter string
    EmphasisRegexp *(regexp.Regexp)
    EmphasisRegexpString string
}

type RhymedResultItems []*ResultItemWithRhymeAndMeter

func (rri RhymedResultItems) Len()          int  { return len(rri) }
func (rri RhymedResultItems) Swap(i, j int)      { rri[i], rri[j] = rri[j], rri[i] }
func (rri RhymedResultItems) Less(i, j int) bool { return rri[i].RhymeAndMeter.FinalSyllable > rri[j].RhymeAndMeter.FinalSyllable }

var syllabi = rhyme.ConstructSyllabi("rhyme/cmudict-0.7b")

func rhymeHandler(w http.ResponseWriter, r *http.Request) {
    searchParams := sapi.SearchParams{
        Text:   r.FormValue("text"),
        Source: r.FormValue("source"),
    }
    sapiResult := sapi.Search( searchParams )

    matchMeter     := r.FormValue("meter")
    if matchMeter == "" {
        matchMeter = rhyme.DefaultMeter
    }

    emphasisRegexp := rhyme.ConvertToEmphasisPointsStringRegexp(matchMeter)

    riwfsList := []*ResultItemWithRhymeAndMeter{}

    for _, item := range *(sapiResult.Items) {
        ram := syllabi.RhymeAndMeterOfPhrase(item.Phrase, emphasisRegexp)

        if ram.EmphasisRegexpMatch2 != "" {
            riwfs := ResultItemWithRhymeAndMeter{
                ResultItem:    item,
                RhymeAndMeter: ram,
            }

            riwfsList = append( riwfsList, &riwfs)            
        }
    }

    sort.Sort(RhymedResultItems(riwfsList))

    srwfs := SearchResultWithRhymeAndMeterList{
        SearchResult: sapiResult,
        ResultItemsWithRhymeAndMeterList:  riwfsList,
        MatchMeter:           matchMeter,
        EmphasisRegexp:       emphasisRegexp,
        EmphasisRegexpString: emphasisRegexp.String(),
    }

    t, _ := template.ParseFiles("rhymed.html")
    t.Execute(w, &srwfs)
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
    if port=="" {
        port = "8080"
    }

	http.HandleFunc("/", alignFormHandler)
    http.HandleFunc("/align", alignHandler)
    http.HandleFunc("/rhyme", rhymeHandler)
	http.ListenAndServe(":"+string(port), nil)
}
