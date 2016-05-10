package ontology

import (
    // "fmt"
    "sort"
    "github.com/railsagainstignorance/alignment/article"
    "github.com/railsagainstignorance/alignment/rhyme"
)

type MatchedPhraseWithUrlWithFirst struct {
    *article.MatchedPhraseWithUrl
    FirstOfNewRhyme bool
}

type Details struct {
    OntologyName      string
    OntologyValue     string
    Meter             string
    Articles                 *[]*article.ArticleWithSentencesAndMeter
    MatchedPhrasesWithUrl    *[]*MatchedPhraseWithUrlWithFirst
    BadMatchedPhrasesWithUrl *[]*MatchedPhraseWithUrlWithFirst
    KnownUnknowns         *[]string
    MaxArticles           int
    NumArticles           int
    SecondaryMatchedPhrasesWithUrl    *[]*(article.MatchedPhraseWithUrl)
    BadSecondaryMatchedPhrasesWithUrl *[]*(article.MatchedPhraseWithUrl)
    MaxMillis             int
    SecondaryMatchedPhrasesWithUrlArticlesAndMPWUs *[]*(ArticleAndMPWUs)
}

type FSandCount struct {
    FinalSyllable string
    Count int
}
type FSandCounts []*FSandCount

func (fsc FSandCounts) Len()          int  { return len(fsc) }
func (fsc FSandCounts) Swap(i, j int)      { fsc[i], fsc[j] = fsc[j], fsc[i] }
func (fsc FSandCounts) Less(i, j int) bool { return (fsc[j].FinalSyllable == "") || ((fsc[i].FinalSyllable != "") && (fsc[i].Count > fsc[j].Count)) }

type ArticleAndMPWUs struct {
    Article *article.ArticleWithSentencesAndMeter
    MPWUs   *[]*article.MatchedPhraseWithUrl
}

const maxMaxArticles = 250

func GetDetails(syllabi *rhyme.Syllabi, ontologyName string, ontologyValue string, meter string, maxArticles int, maxMillis int) (*Details, bool) {

    if maxArticles < 1 {
        maxArticles = 1
    } else if maxArticles > maxMaxArticles {
        maxArticles = maxMaxArticles 
    }

    articles, matchedPhrasesWithUrl := article.GetArticlesByOntologyWithSentencesAndMeter(ontologyName, ontologyValue, meter, syllabi, maxArticles, maxMillis )

    finalSyllablesMap    := &map[string][]*(article.MatchedPhraseWithUrl){}
    badFinalSyllablesMap := &map[string][]*(article.MatchedPhraseWithUrl){}
    secondaryMatchedPhrasesWithUrl    := []*(article.MatchedPhraseWithUrl){}
    badSecondaryMatchedPhrasesWithUrl := []*(article.MatchedPhraseWithUrl){}

    secondaryMatchedPhrasesWithUrlByUrl := map[string]*[]*(article.MatchedPhraseWithUrl){}

    for _,mpwu := range *matchedPhrasesWithUrl {

        if mpwu.MatchesOnMeter.SecondaryMatch != nil {
            // fwwiem := mpwu.MatchesOnMeter.SecondaryMatch.FinalWordWordInEachMatch

            isBadEnd := false
            for _,w := range *(mpwu.MatchesOnMeter.SecondaryMatch.FinalWordWordInEachMatch) {
                if w.IsBadEnd {
                    isBadEnd = true
                    break
                }
            }
            // isBadEnd := (*fwwiem)[len(*fwwiem)-1].IsBadEnd

            if isBadEnd {
                badSecondaryMatchedPhrasesWithUrl = append( badSecondaryMatchedPhrasesWithUrl, mpwu)
            } else {
                secondaryMatchedPhrasesWithUrl = append( secondaryMatchedPhrasesWithUrl, mpwu)
                url := *mpwu.Url
                if _,ok := secondaryMatchedPhrasesWithUrlByUrl[url]; ! ok {
                    secondaryMatchedPhrasesWithUrlByUrl[url] = &[]*(article.MatchedPhraseWithUrl){}
                }

                listOfSMPWU := *(secondaryMatchedPhrasesWithUrlByUrl[url])
                listOfSMPWU = append(listOfSMPWU, mpwu) 
                secondaryMatchedPhrasesWithUrlByUrl[url] = &listOfSMPWU
            }
        }

        fsAZ     := mpwu.MatchesOnMeter.FinalDuringSyllableAZ
        isBadEnd := (mpwu.MatchesOnMeter.FinalDuringWordWord == nil) || mpwu.MatchesOnMeter.FinalDuringWordWord.IsBadEnd

        fsMap := finalSyllablesMap
        if isBadEnd {
            fsMap = badFinalSyllablesMap
        }

        if _, ok := (*fsMap)[fsAZ]; !ok {
            (*fsMap)[fsAZ] = []*(article.MatchedPhraseWithUrl){}
        }

        (*fsMap)[fsAZ] = append((*fsMap)[fsAZ], mpwu)
    }

    processFSMapIntoSortedMPWUs := func( fsMap *map[string][]*(article.MatchedPhraseWithUrl) ) *[]*MatchedPhraseWithUrlWithFirst {
        fsCounts := []*FSandCount{}

        for fs, list := range (*fsMap) {
            fsCounts = append(fsCounts, &FSandCount{fs, len(list)} )
        }

        sort.Sort(FSandCounts(fsCounts))

        sortedMpwus := []*MatchedPhraseWithUrlWithFirst{}

        for _, fsc := range fsCounts {
            fsList := (*fsMap)[fsc.FinalSyllable]
            for i,mpwu := range fsList {
                isFirst := (i==0)
                mpwuf := &MatchedPhraseWithUrlWithFirst{
                    mpwu,
                    isFirst,
                }
                sortedMpwus = append( sortedMpwus, mpwuf)
            }
        }
        return &sortedMpwus
    }

    sortedMpwus    := processFSMapIntoSortedMPWUs(finalSyllablesMap)
    sortedBadMpwus := processFSMapIntoSortedMPWUs(badFinalSyllablesMap)

    listOfArticleAndMPWUs := []*(ArticleAndMPWUs){}

    if len(secondaryMatchedPhrasesWithUrlByUrl) > 0 {
        for _, article := range *articles {
            url := article.SiteUrl
            if mpwus,ok := secondaryMatchedPhrasesWithUrlByUrl[url]; ok {
                articleAndMPWUs := ArticleAndMPWUs{
                    Article: article,
                    MPWUs: mpwus,
                }
                listOfArticleAndMPWUs = append(listOfArticleAndMPWUs, &articleAndMPWUs)
            }
        }
    }

    details := Details{
        OntologyName:          ontologyName,
        OntologyValue:         ontologyValue,
        Meter:                 meter,
        Articles:              articles,
        MatchedPhrasesWithUrl:    sortedMpwus,
        BadMatchedPhrasesWithUrl: sortedBadMpwus,
        KnownUnknowns:         syllabi.KnownUnknowns(),
        MaxArticles:           maxArticles,
        NumArticles:           len(*articles),
        SecondaryMatchedPhrasesWithUrl: &secondaryMatchedPhrasesWithUrl,
        BadSecondaryMatchedPhrasesWithUrl: &badSecondaryMatchedPhrasesWithUrl,
        MaxMillis:             maxMillis,
        SecondaryMatchedPhrasesWithUrlArticlesAndMPWUs: &listOfArticleAndMPWUs,
    }

    containsHaikus := (len(secondaryMatchedPhrasesWithUrl) > 0)

    return &details, containsHaikus
}
