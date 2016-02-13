package main

import (
	// "sort"
 //    "regexp"
    "strings"
    "fmt"
    "github.com/railsagainstignorance/alignment/capi"
    "github.com/railsagainstignorance/alignment/Godeps/_workspace/src/github.com/joho/godotenv"
    "github.com/kennygrant/sanitize"
)

func splitTextIntoSentences(text string) (*[]string) {
    sentences := strings.FieldsFunc(text, func(r rune) bool {
        switch r {
        case '.', ':':
            return true
        }
        return false
    })

    trimmedSentences := []string{}

    for _, s := range sentences {
        ts := strings.TrimSpace(s)
        if ts != "" {
            trimmedSentences = append( trimmedSentences, ts )
        }
    }

    return &trimmedSentences
}

func main() {
    godotenv.Load()

    uuid := "b57fee24-cb3c-11e5-be0b-b7ece4e953a0"

    article := capi.GetArticle( uuid )
    fmt.Println("main: article.Title=", article.Title)

    tidyBody := sanitize.HTML(article.Body)
    fmt.Println("main: len(article.Body)=", len(article.Body), ", len(tidyBody)=", len(tidyBody) )
    fmt.Println("main: tidyBody=", tidyBody)

    sentences := splitTextIntoSentences( tidyBody )
    for _,s := range *sentences {
        fmt.Println("main: s=", s)
    }
}
