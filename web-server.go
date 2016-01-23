package main
    
    import (
    //    "fmt"
        "html/template"
        "net/http"
        "os"
        "github.com/joho/godotenv"
 )
    
    func alignFormHandler(w http.ResponseWriter, r *http.Request) {
        t, _ := template.ParseFiles("align.html")
        t.Execute(w, nil)
    }

    type AlignParams struct {
        Text    string
        Source  string
        SapiKey string
    }

    func alignHandler(w http.ResponseWriter, r *http.Request) {
        t, _ := template.ParseFiles("aligned.html")
        p    := &AlignParams{ Text: r.FormValue("text"), Source: r.FormValue("source"), SapiKey: os.Getenv("SAPI_KEY") }
        t.Execute(w, p)
    }

    func main() {
        godotenv.Load()

        http.HandleFunc("/", alignFormHandler)
        http.HandleFunc("/align", alignHandler)
        http.ListenAndServe(":8080", nil)
    }