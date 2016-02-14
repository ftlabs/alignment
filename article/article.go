package article

import (
	"sort"
	//    "regexp"
	"fmt"
	"github.com/railsagainstignorance/alignment/Godeps/_workspace/src/github.com/joho/godotenv"
	"github.com/railsagainstignorance/alignment/Godeps/_workspace/src/github.com/kennygrant/sanitize"
	"github.com/railsagainstignorance/alignment/capi"
	"github.com/railsagainstignorance/alignment/sapi"
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
    KnownUnknowns *[]string
}

func FindRhymeAndMetersInSentences(sentences *[]string, meter string, syllabi *rhyme.Syllabi) *[]*rhyme.RhymeAndMeter {
	rams := []*rhyme.RhymeAndMeter{}

	if meter == "" {
		meter = rhyme.DefaultMeter
	}

	emphasisRegexp := rhyme.ConvertToEmphasisPointsStringRegexp(meter)

	for _, s := range *(sentences) {
		syllabiRams := syllabi.RhymeAndMetersOfPhrase(s, emphasisRegexp)

		fmt.Println("article: FindRhymeAndMetersInSentences: s=", s, ", len(syllabiRams)=", len(*syllabiRams))

		for _,ram := range *syllabiRams {
			if ram.EmphasisRegexpMatch2 != "" {
				rams = append(rams, ram)
			}			
		}
	}

	return &rams
}

func GetArticleWithSentencesAndMeter(uuid string, meter string, syllabi *rhyme.Syllabi) *ArticleWithSentencesAndMeter {
	aws := getArticleWithSentences(uuid)
	rams := FindRhymeAndMetersInSentences( aws.Sentences, meter, syllabi )

    sort.Sort(rhyme.RhymeAndMeters(*rams))

	awsam := ArticleWithSentencesAndMeter{
		aws,
		meter,
		rams,
        syllabi.KnownUnknowns(),
	}

	return &awsam
}

type MatchedPhraseWithUrl struct {
	*rhyme.RhymeAndMeter
	Url *string
}

type MatchedPhrasesWithUrl []*MatchedPhraseWithUrl

func (mpwus MatchedPhrasesWithUrl) Len()          int  { return len(mpwus) }
func (mpwus MatchedPhrasesWithUrl) Swap(i, j int)      { mpwus[i], mpwus[j] = mpwus[j], mpwus[i] }
func (mpwus MatchedPhrasesWithUrl) Less(i, j int) bool { return mpwus[i].MatchesOnMeter.FinalDuringSyllableAZ > mpwus[j].MatchesOnMeter.FinalDuringSyllableAZ }

func GetArticlesByAuthorWithSentencesAndMeter(author string, meter string, syllabi *rhyme.Syllabi, maxArticles int) (*[]*ArticleWithSentencesAndMeter, *[]*MatchedPhraseWithUrl) {
	sapiResult := sapi.Search( sapi.SearchParams{ Author: author } )

	articles := []*ArticleWithSentencesAndMeter{}

	for _,item := range (*(sapiResult.Items))[0:maxArticles] {
		aws := GetArticleWithSentencesAndMeter(item.Id, meter, syllabi)
		articles = append( articles, aws )
	}

	mpwus := []*MatchedPhraseWithUrl{}

	for _,article := range articles {
		for _, mp := range *(article.MatchedPhrases) {
			mpwu := &MatchedPhraseWithUrl{
				mp,
				&article.Url,
			}

			mpwus = append( mpwus, mpwu )
		}
	}

	return &articles, &mpwus
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
