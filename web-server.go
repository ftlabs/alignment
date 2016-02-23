package main

import (
	"github.com/railsagainstignorance/alignment/Godeps/_workspace/src/github.com/joho/godotenv"
	"html/template"
	"net/http"
	"os"
    "sort"
    "regexp"
    "fmt"
    // "strings"
    "strconv"
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

    emphasisRegexp, _ := rhyme.ConvertToEmphasisPointsStringRegexp(matchMeter)

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

func detailHandler(w http.ResponseWriter, r *http.Request) {
    phrase         := r.FormValue("phrase")
    sentences      := []string{ phrase }
    meter          := r.FormValue("meter")
    rams           := article.FindRhymeAndMetersInSentences( &sentences, meter, syllabi )
    meterRegexp,_  := rhyme.ConvertToEmphasisPointsStringRegexp(meter)

    type PhraseDetails struct {
        Phrase string
        Sentences *[]string
        Meter string
        MeterRegexp *regexp.Regexp
        RhymeAndMeters *[]*rhyme.RhymeAndMeter
        KnownUnknowns *[]string
        EmphasisPointsDetails *rhyme.EmphasisPointsDetails
    }

    pd := PhraseDetails{
        Phrase:         phrase,
        Sentences:      &sentences,
        Meter:          meter,
        MeterRegexp:    meterRegexp,
        RhymeAndMeters: rams,
        KnownUnknowns:  syllabi.KnownUnknowns(),
        EmphasisPointsDetails: syllabi.FindAllEmphasisPointsDetails(phrase),
    }

    templateExecuter( w, "detailPage", pd )
}

func authorHandler(w http.ResponseWriter, r *http.Request) {
    author := r.FormValue("author")
    meter  := r.FormValue("meter")

    maxArticles := 10
    if r.FormValue("max") != "" {
        i, err := strconv.Atoi(r.FormValue("max"))
        if err == nil {
            maxArticles = i
        }
    }

    if maxArticles < 1 {
        maxArticles = 1
    } else if maxArticles > 100 {
        maxArticles = 100 
    }
    
    maxMillis   := 3000

    type MatchedPhraseWithUrlWithFirst struct {
        *article.MatchedPhraseWithUrl
        FirstOfNewRhyme bool
    }

    articles, matchedPhrasesWithUrl := article.GetArticlesByAuthorWithSentencesAndMeter(author, meter, syllabi, maxArticles, maxMillis )
    type AuthorDetails struct {
        Author                string
        Meter                 string
        Articles              *[]*article.ArticleWithSentencesAndMeter
        MatchedPhrasesWithUrl    *[]*MatchedPhraseWithUrlWithFirst
        BadMatchedPhrasesWithUrl *[]*MatchedPhraseWithUrlWithFirst
        KnownUnknowns         *[]string
        MaxArticles           int
        NumArticles           int
        SecondaryMatchedPhrasesWithUrl *[]*(article.MatchedPhraseWithUrl)
        BadSecondaryMatchedPhrasesWithUrl *[]*(article.MatchedPhraseWithUrl)
   }

    finalSyllablesMap    := &map[string][]*(article.MatchedPhraseWithUrl){}
    badFinalSyllablesMap := &map[string][]*(article.MatchedPhraseWithUrl){}
    secondaryMatchedPhrasesWithUrl    := []*(article.MatchedPhraseWithUrl){}
    badSecondaryMatchedPhrasesWithUrl := []*(article.MatchedPhraseWithUrl){}

    for _,mpwu := range *matchedPhrasesWithUrl {

        if mpwu.MatchesOnMeter.SecondaryMatch != nil {
            // fwwiem := mpwu.MatchesOnMeter.SecondaryMatch.FinalWordWordInEachMatch

            isBadEnd := false
            for _,w := range *(mpwu.MatchesOnMeter.SecondaryMatch.FinalWordWordInEachMatch) {
                if w.IsBadEnd {
                    isBadEnd = true
                    break
                }
            }
            // isBadEnd := (*fwwiem)[len(*fwwiem)-1].IsBadEnd

            if isBadEnd {
                badSecondaryMatchedPhrasesWithUrl = append( badSecondaryMatchedPhrasesWithUrl, mpwu)
            } else {
                secondaryMatchedPhrasesWithUrl = append( secondaryMatchedPhrasesWithUrl, mpwu)
            }
        }

        fsAZ     := mpwu.MatchesOnMeter.FinalDuringSyllableAZ
        isBadEnd := (mpwu.MatchesOnMeter.FinalDuringWordWord == nil) || mpwu.MatchesOnMeter.FinalDuringWordWord.IsBadEnd

        fsMap := finalSyllablesMap
        if isBadEnd {
            fsMap = badFinalSyllablesMap
        }

        if _, ok := (*fsMap)[fsAZ]; !ok {
            (*fsMap)[fsAZ] = []*(article.MatchedPhraseWithUrl){}
        }

        (*fsMap)[fsAZ] = append((*fsMap)[fsAZ], mpwu)
    }

    processFSMapIntoSortedMPWUs := func( fsMap *map[string][]*(article.MatchedPhraseWithUrl) ) *[]*MatchedPhraseWithUrlWithFirst {
        fsCounts := []*FSandCount{}

        for fs, list := range (*fsMap) {
            fsCounts = append(fsCounts, &FSandCount{fs, len(list)} )
        }

        sort.Sort(FSandCounts(fsCounts))

        sortedMpwus := []*MatchedPhraseWithUrlWithFirst{}

        for _, fsc := range fsCounts {
            fsList := (*fsMap)[fsc.FinalSyllable]
            for i,mpwu := range fsList {
                isFirst := (i==0)
                mpwuf := &MatchedPhraseWithUrlWithFirst{
                    mpwu,
                    isFirst,
                }
                sortedMpwus = append( sortedMpwus, mpwuf)
            }
        }
        return &sortedMpwus
    }

    sortedMpwus    := processFSMapIntoSortedMPWUs(finalSyllablesMap)
    sortedBadMpwus := processFSMapIntoSortedMPWUs(badFinalSyllablesMap)

    ad := AuthorDetails{
        Author:                author,
        Meter:                 meter,
        Articles:              articles,
        MatchedPhrasesWithUrl:    sortedMpwus,
        BadMatchedPhrasesWithUrl: sortedBadMpwus,
        KnownUnknowns:         syllabi.KnownUnknowns(),
        MaxArticles:           maxArticles,
        NumArticles:           len(*articles),
        SecondaryMatchedPhrasesWithUrl: &secondaryMatchedPhrasesWithUrl,
        BadSecondaryMatchedPhrasesWithUrl: &badSecondaryMatchedPhrasesWithUrl,
    }

    if len(secondaryMatchedPhrasesWithUrl) > 0 {
        templateExecuter( w, "authorHaikuPage", ad )
    } else {
        templateExecuter( w, "authorPage", ad )
    }
}

type ResultWithFinalSyllable struct {
    *sapi.ResultItem
    FinalSyllableAZ string
    FirstOfNewRhyme bool
}

func (f *ResultWithFinalSyllable) SetFirstOfNewRhyme(val bool) {
    f.FirstOfNewRhyme = val
}

type ResultsWithFinalSyllable []*ResultWithFinalSyllable

func (rwfs ResultsWithFinalSyllable) Len()          int  { return len(rwfs) }
func (rwfs ResultsWithFinalSyllable) Swap(i, j int)      { rwfs[i], rwfs[j] = rwfs[j], rwfs[i] }
func (rwfs ResultsWithFinalSyllable) Less(i, j int) bool { return rwfs[i].FinalSyllableAZ > rwfs[j].FinalSyllableAZ }

type FSandCount struct {
    FinalSyllable string
    Count int
}
type FSandCounts []*FSandCount

func (fsc FSandCounts) Len()          int  { return len(fsc) }
func (fsc FSandCounts) Swap(i, j int)      { fsc[i], fsc[j] = fsc[j], fsc[i] }
func (fsc FSandCounts) Less(i, j int) bool { return (fsc[j].FinalSyllable == "") || ((fsc[i].FinalSyllable != "") && (fsc[i].Count > fsc[j].Count)) }

func rhymeHandler(w http.ResponseWriter, r *http.Request) {
    text := r.FormValue("text")
    searchParams := sapi.SearchParams{
        Text:   text,
        Source: "any",
    }

    sapiResult := sapi.Search( searchParams )

    finalSyllablesMap := map[string][]*ResultWithFinalSyllable{}

    for _, item := range *(sapiResult.Items) {
        phrase := item.Title
        fs     := syllabi.FinalSyllableOfPhrase(phrase)
        fsAZ   := rhyme.KeepAZString( fs )
        rwfs   := &ResultWithFinalSyllable{
            item,
            fsAZ,
            false,
        }
        // rwfsList = append( rwfsList, rwfs )
        if _, ok := finalSyllablesMap[fsAZ]; !ok {
            finalSyllablesMap[fsAZ] = []*ResultWithFinalSyllable{}
        }

        finalSyllablesMap[fsAZ] = append(finalSyllablesMap[fsAZ], rwfs)
    }

    fsCounts := []*FSandCount{}

    for fs, list := range finalSyllablesMap {
        fsCounts = append(fsCounts, &FSandCount{fs, len(list)} )
    }

    sort.Sort(FSandCounts(fsCounts))
    rwfsList := []*ResultWithFinalSyllable{}

    for _, fsc := range fsCounts {
        fsList := finalSyllablesMap[fsc.FinalSyllable]
        for i,rwfs := range fsList {
            isFirst := (i == 0)
            rwfs.SetFirstOfNewRhyme( isFirst )
            rwfsList = append( rwfsList, rwfs)
        }
    }

    type Results struct {
        Text string
        ResultsWithFinalSyllable *[]*ResultWithFinalSyllable
    }

    p := Results{
        Text: text,
        ResultsWithFinalSyllable: &rwfsList,
    }

    templateExecuter( w, "rhymedPage", p )
}

func log(fn http.HandlerFunc) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    fmt.Println("REQUEST URL: ", r.URL)
    fn(w, r)
  }
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
    if port=="" {
        port = "8080"
    }

	http.HandleFunc("/",        log(alignFormHandler))
    http.HandleFunc("/align",   log(alignHandler))
    http.HandleFunc("/meter",   log(meterHandler))
    http.HandleFunc("/article", log(articleHandler))
    http.HandleFunc("/detail",  log(detailHandler))
    http.HandleFunc("/author",  log(authorHandler))
    http.HandleFunc("/rhyme",   log(rhymeHandler))

	http.ListenAndServe(":"+string(port), nil)
}
