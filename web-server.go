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
        "strings"
        "sort"
    )
    
    func alignFormHandler(w http.ResponseWriter, r *http.Request) {
        t, _ := template.ParseFiles("align.html")
        t.Execute(w, nil)
    }

    type PhraseBits struct {
        Before string
        Common string
        After  string
    }

    type ByBeforeBit []PhraseBits

    func (s ByBeforeBit) Len() int {
        return len(s)
    }
    func (s ByBeforeBit) Swap(i, j int) {
        s[i], s[j] = s[j], s[i]
    }
    func (s ByBeforeBit) Less(i, j int) bool {
        return len(s[i].Before) > len(s[j].Before)
    }


    type AlignParams struct {
        Text      string
        Source    string
        MaxIndent int
        Phrases   []PhraseBits
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

        textLength := len(text)
        phrases := []PhraseBits{}
        maxIndent := 0

        for _,r := range results {
            excerpt := r.(map[string]interface{})["summary"].(map[string]interface{})["excerpt"].(string)
            if indent := strings.Index(excerpt, text); indent > -1 {
                // phrases = append( phrases, excerpt )
                bits := &PhraseBits{ Before: excerpt[0:indent], Common: text, After: excerpt[indent+textLength:len(excerpt)-1]}
                phrases = append(phrases, *bits)

                if maxIndent < indent {
                    maxIndent = indent
                }
            }
        }

        sort.Sort( ByBeforeBit(phrases) )

        p    := &AlignParams{ Text: text, Source: source, MaxIndent: maxIndent, Phrases: phrases }
        t, _ := template.ParseFiles("aligned.html")
        t.Execute(w, p)
    }

    func main() {
        godotenv.Load()

        http.HandleFunc("/", alignFormHandler)
        http.HandleFunc("/align", alignHandler)
        http.ListenAndServe(":8080", nil)
    }