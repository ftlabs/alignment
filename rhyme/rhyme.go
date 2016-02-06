package rhyme

import (
    "bufio"
    "os"
    "strings"
    "fmt"
    "regexp"
    "sort"
)

const (
	SyllableFilename = "./cmudict-0.7b"
)

type Word struct {
    Name            string
    FragmentsString string
    Fragments       []string
    NumSyllables    int
    FinalSyllable   string
    FinalSyllableAZ string
    EmphasisPoints  []string
    EmphasisPointsString string
}

var (
	syllableRegexp      = regexp.MustCompile(`^[A-Z]+(\d+)$`)
	finalSyllableRegexp = regexp.MustCompile(`([A-Z]+\d+(?:[^\d]*))$`)
	finalWordRegexp     = regexp.MustCompile(`\b(\w+)\b\W*$`)
	wordsRegexp         = regexp.MustCompile(`\b(\w+)\b`)
	unknownEmphasis      = "X"
	loneSyllableEmphasis = "*"
)

func readSyllables(filename string) (*map[string]*Word, int, int) {

	words := map[string]*Word{}
	countFragments      := 0
	countSyllables      := 0
    // syllableRegexp      := regexp.MustCompile(`^[A-Z]+\d+$`)
    // FinalSyllableRegexp := regexp.MustCompile(`([A-Z]+\d+(?:[^\d]*))$`)

    // Open the file.
    f, _ := os.Open(filename)
    // Create a new Scanner for the file.
    scanner := bufio.NewScanner(f)
    // Loop over all lines in the file
    for scanner.Scan() {
		line := scanner.Text()
		if ! strings.HasPrefix(line, ";;;") {
			nameAndRemainder := strings.Split(line, "  ")
			name             := nameAndRemainder[0]
			remainder        := nameAndRemainder[1]
			fragments        := strings.Split(remainder, " ")
			emphasisPoints   := []string{}

			numSyllables := 0
			for _,f := range fragments {
				matches := syllableRegexp.FindStringSubmatch(f)
				if matches != nil {
					numSyllables = numSyllables + 1
					emphasisPoints = append(emphasisPoints, matches[1])
	    		}
	    	}

	    	emphasisPointsString := strings.Join(emphasisPoints, "")

	    	if numSyllables == 0 {
	    		fmt.Println("WARNING: no syllables found for name=", name) 
	    		emphasisPointsString = unknownEmphasis
	    	} else if numSyllables == 1 {
	    		emphasisPointsString = loneSyllableEmphasis
	    	}

	    	matches := finalSyllableRegexp.FindStringSubmatch(remainder)
	    	finalSyllable := ""
	    	if matches != nil {
	    		finalSyllable = matches[1]
	    	} else {
	    		fmt.Println("WARNING: no final syllable found for name=", name) 
	    	}

	    	countSyllables = countSyllables + numSyllables
			countFragments = countFragments + len(fragments)
			words[name] = &Word{
				Name:            name,
				FragmentsString: remainder,
				Fragments:       fragments,
				NumSyllables:    numSyllables,
				FinalSyllable:   finalSyllable,
				FinalSyllableAZ: drop09String(finalSyllable),
				EmphasisPoints:  emphasisPoints,
				EmphasisPointsString: emphasisPointsString,
			}
		}
    }

    // fmt.Println("num fragments = ", countFragments, ", num syllables = ", countSyllables) 

    return &words, countFragments, countSyllables
}

func processFinalSyllables(words *map[string]*Word) (*map[string][]*Word) {
	finalSyllables := map[string][]*Word{}

	for _,word := range *words {
		fs := word.FinalSyllable
		var rhymingWords []*Word

		rhymingWords, ok := finalSyllables[fs]
		if ! ok {
			rhymingWords = []*Word{}
		}

		finalSyllables[fs] = append( rhymingWords, word )
	}

	return &finalSyllables
}

type Stats struct {
	NumWords                int
	NumUniqueFinalSyllables int
	NumFragments            int
	NumSyllables            int 
}

type Syllabi struct {
	Stats          Stats
    SourceFilename string
    FindRhymes     func(string) []string
    CountSyllables func(string) int
    EmphasisPoints func(string) []string
    FinalSyllable  func(string) string
    FinalSyllableOfPhrase func(string) string
    SortPhrasesByFinalSyllable func( []string ) *RhymingPhrases
    RhymeAndMeterOfPhrase func(string) *RhymeAndMeter
    FindMatchingWord func(string) *Word
}

type RhymeAndMeter struct {
	Phrase                       string
	PhraseWords                  *[]string
	MatchingWords                *[]*Word
	EmphasisPointsStrings        *[]string
	EmphasisPointsCombinedString string
	FinalSyllable                string
	FinalSyllableAZ              string
	ContainsUnmatchedWord        bool
	FinalWord                    string
}

type RhymingPhrase struct {
	Phrase        string
	FinalSyllable string
}

type RhymingPhrases []RhymingPhrase

func (rps RhymingPhrases) Len()          int  { return len(rps) }
func (rps RhymingPhrases) Swap(i, j int)      { rps[i], rps[j] = rps[j], rps[i] }
func (rps RhymingPhrases) Less(i, j int) bool { return rps[i].FinalSyllable > rps[j].FinalSyllable }

func keepAZ(r rune) rune { if r>='A' && r<='Z' {return r} else {return -1} }
func KeepAZString( s string ) string {return strings.Map(keepAZ, s)}

func drop09(r rune) rune { if r>='0' && r<='9' {return -1} else {return r} }
func drop09String( s string ) string {return strings.Map(drop09, s)}

var (
	acceptableMeterRegex = regexp.MustCompile(`^\^*[012]+\$*$`)
	defaultMeter         = `01$`
	anchorAtStartChar    = "^"
	wordBoundaryChar     = `\b`
)

// convertToEmphasisPointsStringRegexp takes a string of the form "01010101", or "01010101$", or "^0101",
// and expands it to be able to match against an EmphasisPointsCombinedString,
// with \b prepended if not already anchored to ^.
func convertToEmphasisPointsStringRegexp(meter string) string {
	matchMeter := acceptableMeterRegex.FindString(meter)
	if matchMeter == "" {
		meter = defaultMeter
	}

	meterPieces         := strings.Split(meter, "")
	if meterPieces[0] != anchorAtStartChar {
		meterPieces = append( []string{wordBoundaryChar}, meterPieces...)
	}
	meterWithSpaces     := strings.Join(meterPieces, `\s*`)
	meterWithExpanded0s := strings.Replace(meterWithSpaces, `0`, `[0*]`, -1)
	meterWithExpanded1s := strings.Replace(meterWithExpanded0s, `1`, `[12*]`, -1)

	return meterWithExpanded1s
}

func ConstructSyllabi(sourceFilename string) (*Syllabi){
	if sourceFilename == "" {
		sourceFilename = SyllableFilename
	}

	words, numFragments, numSyllables := readSyllables(sourceFilename)
	finalSyllables := processFinalSyllables(words)

	knownUnknowns := map[string]int{}

	stats := Stats{
		NumWords:                len(*words),
		NumUniqueFinalSyllables: len( *finalSyllables),
		NumFragments:            numFragments,
		NumSyllables:            numSyllables,
	}

	findMatchingWord := func(s string) *Word {
		var word *Word
		upperS := strings.ToUpper(s)
		if w,ok := (*words)[upperS]; ok {
			word = w
		} else if _,ok := knownUnknowns[upperS]; ok {
			knownUnknowns[upperS]++
		} else {
			knownUnknowns[upperS] = 1
			fmt.Println("rhyme: findMatchingWord: new knownUnknown:", upperS)
		}
		return word
	}

	findRhymes := func(s string) []string {
		matchingStrings := []string{}
		matchingWord := findMatchingWord(s)

		if matchingWord != nil {
			finalSyllable := matchingWord.FinalSyllable
		 	if rhymingWords, ok := (*finalSyllables)[finalSyllable]; ok {
		 		for _,w := range rhymingWords {
		 			matchingStrings = append(matchingStrings, (*w).Name)
				}
			}
		}

		return matchingStrings
	}

	countSyllables := func(s string) int {
		count  := 0
		w := findMatchingWord(s)
		if w != nil {
			count = (*w).NumSyllables
		}

		return count
	}

	emphasisPoints := func(s string) []string {
		ep := []string{}
		w := findMatchingWord(s)
		if w != nil {
			ep = (*w).EmphasisPoints
		}
		return ep
	}

	finalSyllableFunc := func(s string) string {
		fs := ""
		w := findMatchingWord(s)

		if w != nil {
			fs = (*w).FinalSyllable
		}
		return fs
	}

	finalSyllableOfPhraseFunc := func(s string) string {
		finalWord := ""
		matches := finalWordRegexp.FindStringSubmatch(s)
		if matches != nil {
			finalWord = matches[1]
		}

		fs := finalSyllableFunc(finalWord)
		return fs
	}

	rhymeAndMeterOfPhrase := func(phrase string) *RhymeAndMeter {
		finalWord := ""
		phraseWords := []string{}
		matchingWords := []*Word{}
		emphasisPointsStrings := []string{}
		emphasisPointsCombinedString := ""
		containsUnmatchedWord := false
		finalSyllable   := ""
		finalSyllableAZ := ""

		matches := wordsRegexp.FindAllStringSubmatch(phrase, -1)
		if matches != nil {
			for _, match := range matches{
				phraseWord := match[1]
				phraseWords = append( phraseWords, phraseWord)
				matchingWord := findMatchingWord(phraseWord)
				emphasisPointsString := "X"
				if matchingWord == nil {
					containsUnmatchedWord = true
				} else {
					emphasisPointsString = matchingWord.EmphasisPointsString
				}

				matchingWords = append(matchingWords, matchingWord)
				emphasisPointsStrings = append( emphasisPointsStrings, emphasisPointsString)
			}

			finalMatchingWord := matchingWords[len(matchingWords)-1]; 
			if finalMatchingWord != nil {
				finalSyllable   = finalMatchingWord.FinalSyllable
				finalSyllableAZ = finalMatchingWord.FinalSyllableAZ
			}

			emphasisPointsCombinedString = strings.Join(emphasisPointsStrings, " ")
		}

		ram := RhymeAndMeter{
			Phrase:                       phrase,
			PhraseWords:                  &phraseWords,
			MatchingWords:                &matchingWords,
			EmphasisPointsStrings:        &emphasisPointsStrings,
			EmphasisPointsCombinedString: emphasisPointsCombinedString,
			FinalSyllable:                finalSyllable,
			FinalSyllableAZ:              finalSyllableAZ,
			ContainsUnmatchedWord:        containsUnmatchedWord,
			FinalWord:                    finalWord,
		}

		return &ram
	}


	sortPhrasesByFinalSyllable := func(phrases []string) *RhymingPhrases {
		rhymingPhrases := RhymingPhrases{}
		for _,p := range phrases {
			fs := finalSyllableOfPhraseFunc(p)
			fsAZ := KeepAZString(fs)

			rp := RhymingPhrase{
				Phrase:        p,
				FinalSyllable: fsAZ,
			}
			rhymingPhrases = append(rhymingPhrases, rp)
		}

    	sort.Sort(RhymingPhrases(rhymingPhrases))

		return &rhymingPhrases
	}

	syllabi := Syllabi{
		Stats:          stats,
		SourceFilename: sourceFilename,
		FindRhymes:     findRhymes,
		CountSyllables: countSyllables,
		EmphasisPoints: emphasisPoints,
		FinalSyllable:  finalSyllableFunc,
		FinalSyllableOfPhrase: finalSyllableOfPhraseFunc,
		SortPhrasesByFinalSyllable: sortPhrasesByFinalSyllable,
		RhymeAndMeterOfPhrase:      rhymeAndMeterOfPhrase,
		FindMatchingWord:           findMatchingWord,
	}

	return &syllabi
}

func main() {
	syllabi := ConstructSyllabi(SyllableFilename)

    fmt.Println("main: num words =", (*syllabi).Stats.NumWords ) 
	fmt.Println("main: num unique final syllables =", (*syllabi).Stats.NumUniqueFinalSyllables)
	fmt.Println("main: num fragments =", (*syllabi).Stats.NumFragments)
	fmt.Println("main: num syllables =", (*syllabi).Stats.NumSyllables)

	s := "hyperactivity"
	rhymesWith := (*syllabi).FindRhymes(s)
	sort.Strings(rhymesWith)
	fmt.Println("main:", s, "rhymes with", len(rhymesWith), ": first=", rhymesWith[0], ", last=", rhymesWith[len(rhymesWith)-1])

	numSyllables := (*syllabi).CountSyllables(s)
	fmt.Println("main:", s, "has", numSyllables, "syllables")

	ep := (*syllabi).EmphasisPoints(s)
	fmt.Println("main:", s, "emphasisPoints=", strings.Join(ep, ","))

	fs := (*syllabi).FinalSyllable(s)
	fmt.Println("main:", s, "finalSyllable=", fs)

	p := "bananas are the scourge of " + s
	fsop := (*syllabi).FinalSyllableOfPhrase( p )
	fmt.Println( "main:", p, ": final syllable of phrase=", fsop)

	phrases := []string{
		"I am a savanna",
		"please fetch me my banjo.",
		"give me my bandana",
		"I am an armadillo",
		"catch me if you can.",
		"so so so.",
	}
	fmt.Println("main: phrases:")
	fmt.Println(strings.Join(phrases, "\n"))

	rps := (*syllabi).SortPhrasesByFinalSyllable( phrases)
	fmt.Println("main: sorted rhyming phrases:")
	for _,rp := range *rps {
		fmt.Println("fs:", rp.FinalSyllable, ", phrase:", rp.Phrase)
	}
}
