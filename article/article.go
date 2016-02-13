package article

import (
	// "sort"
	//    "regexp"
	"fmt"
	"github.com/railsagainstignorance/alignment/Godeps/_workspace/src/github.com/joho/godotenv"
	"github.com/railsagainstignorance/alignment/Godeps/_workspace/src/github.com/kennygrant/sanitize"
	"github.com/railsagainstignorance/alignment/capi"
	"github.com/railsagainstignorance/alignment/rhyme"
	"strings"
)

func splitTextIntoSentences(text string) *[]string {
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
			trimmedSentences = append(trimmedSentences, ts)
		}
	}

	return &trimmedSentences
}

type ArticleWithSentences struct {
	*capi.Article
	Sentences *[]string
}

func getArticleWithSentences(uuid string) *ArticleWithSentences {
	article := capi.GetArticle(uuid)
	tidyBody := sanitize.HTML(article.Body)

	sentences := splitTextIntoSentences(tidyBody)
	for _, s := range *sentences {
		fmt.Println("main: s=", s)
	}

	aws := ArticleWithSentences{
		article,
		sentences,
	}

	return &aws
}

type ArticleWithSentencesAndMeter struct {
	*ArticleWithSentences
	Meter          string
	MatchedPhrases *[]*rhyme.RhymeAndMeter
}

func GetArticleWithSentencesAndMeter(uuid string, meter string, syllabi *rhyme.Syllabi) *ArticleWithSentencesAndMeter {
	aws := getArticleWithSentences(uuid)
	rams := []*rhyme.RhymeAndMeter{}

	if meter == "" {
		meter = rhyme.DefaultMeter
	}

	emphasisRegexp := rhyme.ConvertToEmphasisPointsStringRegexp(meter)

	for _, s := range *(aws.Sentences) {
		ram := syllabi.RhymeAndMeterOfPhrase(s, emphasisRegexp)

		if ram.EmphasisRegexpMatch2 != "" {
			rams = append(rams, ram)
		}
	}

	awsam := ArticleWithSentencesAndMeter{
		aws,
		meter,
		&rams,
	}

	return &awsam
}

func main() {
	godotenv.Load()
	var syllabi = rhyme.ConstructSyllabi(&[]string{"../rhyme/cmudict-0.7b", "../rhyme/cmudict-0.7b_my_additions"})

	uuid := "b57fee24-cb3c-11e5-be0b-b7ece4e953a0"
	meter := "1010101010"

	aws := GetArticleWithSentencesAndMeter(uuid, meter, syllabi)
	fmt.Println("main: article.Title=", aws.Title)
	fmt.Println("main: body=", aws.Body)

	fmt.Println("main: num Sentences=", len(*(aws.Sentences)), "num MatchedPhrases=", len(*(aws.MatchedPhrases)))

	for _, mp := range *(aws.MatchedPhrases) {
		fmt.Println("main: mp.MatchesOnMeter.During=", mp.MatchesOnMeter.During)
	}
}
