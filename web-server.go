package main

import (
	"github.com/upthebuzzard/alignment/Godeps/_workspace/src/github.com/joho/godotenv"
	"html/template"
	"net/http"
	"os"
    "sort"
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

type ResultItemWithFinalSyllable struct {
    ResultItem    *(sapi.ResultItem)
    FinalSyllable string
}

type SearchResultWithFinalSyllables struct {
    SearchResult sapi.SearchResult
    ItemsWithFS  []ResultItemWithFinalSyllable
}

type RhymedResultItems []ResultItemWithFinalSyllable

func (rri RhymedResultItems) Len()          int  { return len(rri) }
func (rri RhymedResultItems) Swap(i, j int)      { rri[i], rri[j] = rri[j], rri[i] }
func (rri RhymedResultItems) Less(i, j int) bool { return rri[i].FinalSyllable > rri[j].FinalSyllable }

var syllabi = rhyme.ConstructSyllabi("rhyme/cmudict-0.7b")

func rhymeHandler(w http.ResponseWriter, r *http.Request) {
    searchParams := sapi.SearchParams{
        Text:   r.FormValue("text"),
        Source: r.FormValue("source"),
    }
    sapiResult := sapi.Search( searchParams )

    riwfsList := []ResultItemWithFinalSyllable{}

    for _, item := range *(sapiResult.Items) {
        fs := syllabi.FinalSyllableOfPhrase(item.Phrase)
        // fmt.Println("rhymeHandler: loop:", "item.Phrase=", item.Phrase, ", fs=", fs)
        riwfs := ResultItemWithFinalSyllable{
            ResultItem:    item,
            FinalSyllable: fs,
        }

        riwfsList = append( riwfsList, riwfs)
    }

    sort.Sort(RhymedResultItems(riwfsList))

    srwfs := SearchResultWithFinalSyllables{
        SearchResult: *sapiResult,
        ItemsWithFS:  riwfsList,
    }

    // fmt.Println("rhymeHandler:", "srwfs.ItemsWithFS[0].ResultItem.Phrase=", srwfs.ItemsWithFS[0].ResultItem.Phrase)
    // fmt.Println("rhymeHandler:", "srwfs.ItemsWithFS[0].FinalSyllable=", srwfs.ItemsWithFS[0].FinalSyllable)

    t, _ := template.ParseFiles("rhymed.html")
    t.Execute(w, &srwfs)
    // t, _ := template.ParseFiles("rhymed.html")
    // t.Execute(w, nil)
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
