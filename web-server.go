package main
    
    import (
        "fmt"
        "html/template"
        "net/http"
        "os"
        "github.com/joho/godotenv"
        "bytes"
        "io/ioutil"
        "encoding/json"
        "reflect"
)
    
    func alignFormHandler(w http.ResponseWriter, r *http.Request) {
        t, _ := template.ParseFiles("align.html")
        t.Execute(w, nil)
    }

    type AlignParams struct {
        Text    string
        Source  string
        SapiKey string
        Body    string
    }

    func alignHandler(w http.ResponseWriter, r *http.Request) {

        text    := r.FormValue("text")
        source  := r.FormValue("source")
        sapiKey := os.Getenv("SAPI_KEY")

        url := "http://api.ft.com/content/search/v1?apiKey=" + sapiKey
        fmt.Println("url:", url)
        var jsonStr = []byte(`{"queryString": "` + text + `","queryContext" : {"curations" : [ "ARTICLES", "BLOGS" ]},  "resultContext" : {"maxResults" : "100", "offset" : "0", "aspects" : [ "title", "location", "summary", "lifecycle", "metadata"], "sortOrder": "DESC", "sortField": "lastPublishDateTime" } } }`)
        // var jsonStr = []byte(`{"queryString": "`)

        req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
        req.Header.Set("Content-Type", "application/json")
        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            panic(err)
        }
        defer resp.Body.Close()

        fmt.Println("response Status:", resp.Status)
        fmt.Println("response Headers:", resp.Header)
        jsonBody, _ := ioutil.ReadAll(resp.Body)

        var data interface{}
        json.Unmarshal(jsonBody, &data)
        fmt.Println("TypeOf data:", reflect.TypeOf(data))
        results := data.(map[string]interface{})[`results`].([]interface{})[0].(map[string]interface{})[`results`].([]interface{})
        fmt.Println("TypeOf results:", reflect.TypeOf(results))

        // firstExcerpt := results[0].(map[string]interface{})["summary"]

        firstExcerpt := results[0].(map[string]interface{})["summary"].(map[string]interface{})["excerpt"].(string)
        fmt.Println("TypeOf firstExcerpt:", reflect.TypeOf(firstExcerpt))
        body := firstExcerpt

        // body := "testing 123"

        p    := &AlignParams{ Text: text, Source: source, SapiKey: sapiKey, Body: string(body) }
        t, _ := template.ParseFiles("aligned.html")
        t.Execute(w, p)
    }

    func main() {
        godotenv.Load()

        http.HandleFunc("/", alignFormHandler)
        http.HandleFunc("/align", alignHandler)
        http.ListenAndServe(":8080", nil)
    }