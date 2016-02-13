package main

import (
	"github.com/railsagainstignorance/alignment/Godeps/_workspace/src/github.com/joho/godotenv"
	"html/template"
	"net/http"
	"os"
    "sort"
    "regexp"
    // "fmt"
    "github.com/railsagainstignorance/alignment/align"
    "github.com/railsagainstignorance/alignment/sapi"
    "github.com/railsagainstignorance/alignment/rhyme"
    "github.com/railsagainstignorance/alignment/article"
)

// compile all templates and cache them
var templates = template.Must(template.ParseGlob("templates/*"))

func templateExecuter( w http.ResponseWriter, pageName string, data interface{} ){
    err := templates.ExecuteTemplate(w, pageName, data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }    
}

func alignFormHandler(w http.ResponseWriter, r *http.Request) {
    templateExecuter( w, "alignPage", nil )
}

func alignHandler(w http.ResponseWriter, r *http.Request) {
    searchParams := sapi.SearchParams{
        Text:   r.FormValue("text"),
        Source: r.FormValue("source"),
    }

	p := align.Search( searchParams )
    templateExecuter( w, "alignedPage", p )
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
    KnownUnknowns *[]string
    PhraseWordsRegexpString string
}

type RhymedResultItems []*ResultItemWithRhymeAndMeter

func (rri RhymedResultItems) Len()          int  { return len(rri) }
func (rri RhymedResultItems) Swap(i, j int)      { rri[i], rri[j] = rri[j], rri[i] }
func (rri RhymedResultItems) Less(i, j int) bool { return rri[i].RhymeAndMeter.FinalSyllable > rri[j].RhymeAndMeter.FinalSyllable }

var syllabi = rhyme.ConstructSyllabi(&[]string{"rhyme/cmudict-0.7b", "rhyme/cmudict-0.7b_my_additions"})

func meterHandler(w http.ResponseWriter, r *http.Request) {
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
        rams := syllabi.RhymeAndMetersOfPhrase(item.Phrase, emphasisRegexp)

        if rams != nil {
            for _,ram := range *rams {
                if ram.EmphasisRegexpMatch2 != "" {
                    riwfs := ResultItemWithRhymeAndMeter{
                        ResultItem:    item,
                        RhymeAndMeter: ram,
                    }

                    riwfsList = append( riwfsList, &riwfs)            
                }
            }
        }
    }

    sort.Sort(RhymedResultItems(riwfsList))

    srwfs := SearchResultWithRhymeAndMeterList{
        SearchResult: sapiResult,
        ResultItemsWithRhymeAndMeterList:  riwfsList,
        MatchMeter:           matchMeter,
        EmphasisRegexp:       emphasisRegexp,
        EmphasisRegexpString: emphasisRegexp.String(),
        KnownUnknowns:        syllabi.KnownUnknowns(),
        PhraseWordsRegexpString: syllabi.PhraseWordsRegexpString,
    }

    templateExecuter( w, "meteredPage", &srwfs )
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
    uuid  := r.FormValue("uuid")
    meter := r.FormValue("meter")

    p := article.GetArticleWithSentencesAndMeter(uuid, meter, syllabi )
    templateExecuter( w, "articlePage", p )
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
    if port=="" {
        port = "8080"
    }

	http.HandleFunc("/", alignFormHandler)
    http.HandleFunc("/align", alignHandler)
    http.HandleFunc("/meter", meterHandler)
    http.HandleFunc("/article", articleHandler)
	http.ListenAndServe(":"+string(port), nil)
}
