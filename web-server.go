package main

import (
	"fmt"
	"github.com/Financial-Times/ft-s3o-go/s3o"
	"github.com/joho/godotenv"
	"github.com/railsagainstignorance/alignment/align"
	"github.com/railsagainstignorance/alignment/article"
	"github.com/railsagainstignorance/alignment/ontology"
	"github.com/railsagainstignorance/alignment/rhyme"
	"github.com/railsagainstignorance/alignment/rss"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

// compile all templates and cache them
var templates = template.Must(template.ParseGlob("templates/*"))

// construct the syllable monster
var syllabi = rhyme.ConstructSyllabi(&[]string{"rhyme/cmudict-0.7b", "rhyme/cmudict-0.7b_my_additions"})

func templateExecuter(w http.ResponseWriter, pageName string, data interface{}) {
	err := templates.ExecuteTemplate(w, pageName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func alignFormHandler(w http.ResponseWriter, r *http.Request) {
	templateExecuter(w, "alignPage", nil)
}

func alignHandler(w http.ResponseWriter, r *http.Request) {
	p := align.Search(r.FormValue("text"), r.FormValue("source"))
	templateExecuter(w, "alignedPage", p)
}

func detailHandler(w http.ResponseWriter, r *http.Request) {
	phrase := r.FormValue("phrase")
	sentences := []string{phrase}
	meter := r.FormValue("meter")
	rams := article.FindRhymeAndMetersInSentences(&sentences, meter, syllabi)
	meterRegexp, _ := rhyme.ConvertToEmphasisPointsStringRegexp(meter)

	type PhraseDetails struct {
		Phrase                string
		Sentences             *[]string
		Meter                 string
		MeterRegexp           *regexp.Regexp
		RhymeAndMeters        *[]*rhyme.RhymeAndMeter
		KnownUnknowns         *[]string
		EmphasisPointsDetails *rhyme.EmphasisPointsDetails
	}

	pd := PhraseDetails{
		Phrase:                phrase,
		Sentences:             &sentences,
		Meter:                 meter,
		MeterRegexp:           meterRegexp,
		RhymeAndMeters:        rams,
		KnownUnknowns:         syllabi.KnownUnknowns(),
		EmphasisPointsDetails: syllabi.FindAllEmphasisPointsDetails(phrase),
	}

	templateExecuter(w, "detailPage", pd)
}

func ontologyHandler(w http.ResponseWriter, r *http.Request) {
	ontologyName := r.FormValue("ontology")
	ontologyValue := r.FormValue("value")
	meter := r.FormValue("meter")

	maxArticles := 10
	if r.FormValue("max") != "" {
		i, err := strconv.Atoi(r.FormValue("max"))
		if err == nil {
			maxArticles = i
		}
	}

	maxMillis := 30000

	details, containsHaikus := ontology.GetDetails(syllabi, ontologyName, ontologyValue, meter, maxArticles, maxMillis)

	if containsHaikus {
		templateExecuter(w, "ontologyHaikuPage", details)
	} else {
		templateExecuter(w, "ontologyPage", details)
	}
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
	maxItems := 20
	rssText := rss.Generate(maxItems)
	w.Header().Set("Content-Type", "application/rss+xml")
	fmt.Fprintf(w, *rssText)
}

func carouselHandler(w http.ResponseWriter, r *http.Request) {
	maxItems := 20
	items := rss.GenerateItems(maxItems)

	type CarouselDetails struct {
		MaxItems int
		Items	 *[]*rss.Haiku
	}

	cd := CarouselDetails{
		MaxItems: maxItems,
		Items:    items,
	}

	templateExecuter(w, "carouselPage", cd)
}

func meditationHandler(w http.ResponseWriter, r *http.Request) {
	templateExecuter(w, "meditationPage", nil)
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
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", log(alignFormHandler))
	http.HandleFunc("/align", log(alignHandler))
	http.HandleFunc("/detail", log(detailHandler))
	http.HandleFunc("/rss", log(rssHandler))
	http.HandleFunc("/carousel", log(carouselHandler))
	http.Handle("/ontology", s3o.Handler(http.HandlerFunc(log(ontologyHandler))))
	http.HandleFunc("/meditation", log(meditationHandler))
    
    http.Handle("/javascript/", http.StripPrefix("/javascript/", http.FileServer(http.Dir("./public/javascript"))))
    http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("./public/data"))))


	http.ListenAndServe(":"+string(port), nil)
}
