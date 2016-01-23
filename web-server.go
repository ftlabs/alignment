package main
    
    import (
    //    "fmt"
        "html/template"
        "net/http"
    )
    
    func alignFormHandler(w http.ResponseWriter, r *http.Request) {
        t, _ := template.ParseFiles("align.html")
        t.Execute(w, nil)
    }

    type AlignParams struct {
        Text string
        Source string
    }

    func alignHandler(w http.ResponseWriter, r *http.Request) {
        t, _ := template.ParseFiles("aligned.html")
        p := &AlignParams{Text: r.FormValue("text"), Source: r.FormValue("source")}
        t.Execute(w, p)
    }

    func main() {
        http.HandleFunc("/", alignFormHandler)
        http.HandleFunc("/align", alignHandler)
        http.ListenAndServe(":8080", nil)
    }