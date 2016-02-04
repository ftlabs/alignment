package main

import (
	"github.com/upthebuzzard/alignment/Godeps/_workspace/src/github.com/joho/godotenv"
	"html/template"
	"net/http"
	"os"
    "github.com/upthebuzzard/alignment/align"
    "github.com/upthebuzzard/alignment/sapi"
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

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
    if port=="" {
        port = "8080"
    }

	http.HandleFunc("/", alignFormHandler)
	http.HandleFunc("/align", alignHandler)
	http.ListenAndServe(":"+string(port), nil)
}
