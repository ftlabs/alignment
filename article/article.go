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

type ArticleWithSentences struct {
    *capi.Article
    Sentences *[]string
}

func getArticleWithSentences( uuid string ) (*ArticleWithSentences){
    article  := capi.GetArticle( uuid )
    tidyBody := sanitize.HTML(article.Body)

    sentences := splitTextIntoSentences( tidyBody )
    for _,s := range *sentences {
        fmt.Println("main: s=", s)
    }

    aws := ArticleWithSentences{
        article,
        sentences,
    }

    return &aws
}

func main() {
    godotenv.Load()

    uuid := "b57fee24-cb3c-11e5-be0b-b7ece4e953a0"

    aws := getArticleWithSentences( uuid )
    fmt.Println("main: article.Title=", aws.Title)
    fmt.Println("main: body=", aws.Body)

    for _,s := range *(aws.Sentences) {
        fmt.Println("main: s=", s)
    }
}
