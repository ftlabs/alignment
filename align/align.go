package align

import (
	"sort"
    "regexp"
    "github.com/railsagainstignorance/alignment/sapi"
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

func Search(params sapi.SearchParams) *ResultParams {
    sapiResults := sapi.Search( params )

    var (
        maxIndent int = 0
        phrases []PhraseBits = []PhraseBits{}
    )

    for _, resultItem := range (*(*sapiResults).Items) {

        splitPhrase := FullOnPartial(resultItem.Phrase, params.Text)   

		if splitPhrase.Indent >= 0 {
			bits := &PhraseBits{
				Before:      splitPhrase.Before,
				Common:      splitPhrase.Common,
				After:       splitPhrase.After,
				Excerpt:     resultItem.Excerpt,
				Title:       resultItem.Title,
				LocationUri: resultItem.LocationUri,
			}
			phrases = append(phrases, *bits)

			if maxIndent < splitPhrase.Indent {
				maxIndent = splitPhrase.Indent
			}                    
    	}

    	// because it looks better this way
    	sort.Sort(ByBeforeBit(phrases))
    } 

	p := &ResultParams{
        Text:             params.Text, 
        Source:           sapiResults.Source, 
        MaxIndent:        maxIndent, 
        Phrases:          phrases,
        FtcomUrl:         sapiResults.FtcomUrl,
        FtcomSearchUrl:   sapiResults.FtcomSearchUrl,
        TitleOnlyChecked: sapiResults.TitleOnlyChecked,
        AnyChecked:       sapiResults.AnyChecked,
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