package align

import (
	"sort"
    "regexp"
    // "fmt"
    // "github.com/railsagainstignorance/alignment/sapi"
    "github.com/railsagainstignorance/alignment/content"
)

type PhraseBits struct {
	Before      string
	Common      string
	After       string
	Excerpt     string
	Title       string
	LocationUri string
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

type ResultParams struct {
    Text      string
    Source    string
    MaxIndent int
    Phrases   []PhraseBits
    FtcomUrl       string
    FtcomSearchUrl string
    TitleOnlyChecked string
    AnyChecked       string
}

func Search(text string, source string) *ResultParams {
    // sapiResults := sapi.Search( params )

    var textForSearch string

    if source != "title-only" {
        source = "keyword"
        textForSearch = `\"` + text + `\"`
    } else {
        textForSearch = text
    }

    sRequest := &content.SearchRequest {
        QueryType: source,
        QueryText: textForSearch,
        MaxArticles: 100,
        MaxDurationMillis: 3000,
        SearchOnly: true, // i.e. don't bother looking up articles
    }

    // fmt.Println("align.Search: sRequest=", sRequest) 

    sapiResults := content.Search( sRequest )

    // fmt.Println("align.Search: sapiResults=", sapiResults) 
    // fmt.Println("align.Search: sapiResults.Articles=", sapiResults.Articles) 

    var (
        maxIndent int = 0
        phrases []PhraseBits = []PhraseBits{}
    )

    for _, resultItem := range (*(*sapiResults).Articles) {

        var phrase string

        if source == "title-only" {
            phrase = resultItem.Title
        } else {
            phrase = resultItem.Excerpt
        }

        splitPhrase := FullOnPartial(phrase, text)   

        // fmt.Println("align.Search: resultItem.Excerpt=", resultItem.Excerpt, ", text=", text, ", splitPhrase=", splitPhrase) 

		if splitPhrase.Indent >= 0 {
			bits := &PhraseBits{
				Before:      splitPhrase.Before,
				Common:      splitPhrase.Common,
				After:       splitPhrase.After,
				Excerpt:     resultItem.Excerpt,
				Title:       resultItem.Title,
				LocationUri: resultItem.SiteUrl,
			}
			phrases = append(phrases, *bits)

			if maxIndent < splitPhrase.Indent {
				maxIndent = splitPhrase.Indent
			}                    
    	}

    	// because it looks better this way
    	sort.Sort(ByBeforeBit(phrases))
    } 

    var (
        titleOnlyChecked string = ""
        anyChecked       string = ""
    )

    if source == "title-only" {
        titleOnlyChecked = "checked"
    } else {
        anyChecked       = "checked"
    }

	p := &ResultParams{
        Text:             text, 
        Source:           source, 
        MaxIndent:        maxIndent, 
        Phrases:          phrases,
        FtcomUrl:         sapiResults.SiteUrl,
        FtcomSearchUrl:   sapiResults.SiteSearchUrl,
        TitleOnlyChecked: titleOnlyChecked,
        AnyChecked:       anyChecked,
    }

    return p
}

type SplitPhrase struct {
    Before      string
    Common      string
    After       string
    Indent      int
    Full        string
    Partial     string
}

func FullOnPartial(full string,  partial string) *SplitPhrase {

    phraseRegexp := regexp.MustCompile(`(?i)^(.*)\b(` + partial + `)\b(.*)$`)
    matches      := phraseRegexp.FindStringSubmatch(full)

    var(
        before      string = ""
        matchedText string = ""
        after       string = ""
        indent      int    = -1
    )

    if matches != nil {
        before      = matches[1]
        matchedText = matches[2]
        after       = matches[3]
        indent      = len(before)
    }

    sp := SplitPhrase{
        Before:      before,
        Common:      matchedText,
        After:       after,
        Indent:      indent,
        Full:        full,
        Partial:     partial,
    }

    return &sp
}