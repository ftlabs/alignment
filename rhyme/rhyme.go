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
    Unknown bool
    IsBadEnd bool
}

func (f *Word) SetIsBadEnd(val bool) {
    f.IsBadEnd = val
}

type TransformPair struct {
	Regexp      *(regexp.Regexp)
	Replacement string
}

var (
	syllableRegexp      = regexp.MustCompile(`^[A-Z]+(\d+)$`)
	finalSyllableRegexp = regexp.MustCompile(`([A-Z]+\d+(?:[^\d]*))$`)
	unknownEmphasis      = "X"
	loneSyllableEmphasis = "*"
	stringsAsKeys        = map[string]string{}
	wordRegexps          = []string{}
	nameTransformPairs   = []TransformPair{} 
)

func readSyllables(filenames *[]string) (*map[string]*Word, int, int) {

	words := map[string]*Word{}
	countFragments      := 0
	countSyllables      := 0
	badEnds := []string{}            

	for _,filename := range *filenames {
	    fmt.Println("rhyme: readSyllables: readingfrom: filename=", filename) 

	    // Open the file.
	    f, _ := os.Open(filename)
	    // Create a new Scanner for the file.
	    scanner := bufio.NewScanner(f)
	    // Loop over all lines in the file
	    for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				fmt.Println("WARNING: rhyme: readSyllables: empty line: ignoring")
			} else if strings.HasPrefix(line, ";;;") {
				// ignore it
			} else {
				nameAndRemainder := strings.Split(line, "  ")
				if len(nameAndRemainder) != 2 {
					fmt.Println("WARNING: rhyme: readSyllables: line doesn't split on double space:", line)
				} else {
					name      := nameAndRemainder[0]
					remainder := nameAndRemainder[1]

					if strings.HasPrefix(name, "MAP:") {
						namePieces := strings.Split(name, ":")
						stringsAsKeys[namePieces[1]] = remainder
					} else if strings.HasPrefix(name, "WORD:") {
						wordRegexps = append(wordRegexps, remainder)
					} else if strings.HasPrefix(name, "TRANSFORM:") {
						namePieces := strings.Split(name, ":")
						transformPair := TransformPair{
							Regexp:      regexp.MustCompile(namePieces[1]),
							Replacement: remainder,
						}
						nameTransformPairs = append( nameTransformPairs, transformPair)
					} else if name == "BAD:END" {
						badEnds = append( badEnds, remainder )
					} else {
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
							Unknown:         false,
							IsBadEnd:        false,
						}
					}
				}
			}
		}
    }

    for _,be := range badEnds {
    	if w,ok := words[be]; ok {
    		w.SetIsBadEnd( true )
    	}
    }

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

type EmphasisPointsDetails struct {
	Phrase                       string
	PhraseMatches                *[][]string
	PhraseWords                  []string
	MatchingWords                []*Word
	ContainsUnmatchedWord        bool
	EmphasisPointsStrings        []string
	FinalSyllable                string
	FinalSyllableAZ              string
	EmphasisPointsCombinedString string
	FinalMatchingWord            *Word
}



type Syllabi struct {
	Stats          Stats
    SourceFilenames *[]string
    FindRhymes     func(string) []string
    CountSyllables func(string) int
    EmphasisPoints func(string) []string
    FinalSyllable  func(string) string
    FinalSyllableOfPhrase func(string) string
    SortPhrasesByFinalSyllable func( []string ) *RhymingPhrases
    RhymeAndMetersOfPhrase func(string, ...*regexp.Regexp) *[]*RhymeAndMeter
    FindMatchingWord func(string) *Word
    KnownUnknowns func() *[]string
	PhraseWordsRegexp            *regexp.Regexp
	PhraseWordsRegexpString      string
	FindAllEmphasisPointsDetails func(string) *EmphasisPointsDetails
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
	EmphasisRegexp               *regexp.Regexp
	EmphasisRegexpString         string
	EmphasisRegexpMatches        []string
	EmphasisRegexpMatch2         string
	MatchesOnMeter               *MatchesOnMeter
}

type RhymeAndMeters []*RhymeAndMeter

func (rams RhymeAndMeters) Len()          int  { return len(rams) }
func (rams RhymeAndMeters) Swap(i, j int)      { rams[i], rams[j] = rams[j], rams[i] }
func (rams RhymeAndMeters) Less(i, j int) bool { return rams[i].MatchesOnMeter.FinalDuringSyllableAZ > rams[j].MatchesOnMeter.FinalDuringSyllableAZ }

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
	acceptableMeterRegex = regexp.MustCompile(`^(\^*)([012 \.]*)(\$*)$`)
	multipleSpacesRegex  = regexp.MustCompile(`\s\s+`)
	DefaultMeter         = `01$`
	anchorAtStartChar    = "^"
	anchorAtEndChar      = "$"
	wordBoundaryChar     = `\b`
)

// ConvertToEmphasisPointsStringRegexp takes a string of the form "01010101", or "01010101$", or "^0101",
// and expands it to be able to match against an EmphasisPointsCombinedString,
// with \b prepended if not already anchored to ^.
func ConvertToEmphasisPointsStringRegexp(meter string) (*regexp.Regexp, *regexp.Regexp) {
	matchMeter := acceptableMeterRegex.FindStringSubmatch(meter)

	if matchMeter == nil {
		meter      = DefaultMeter
		matchMeter = acceptableMeterRegex.FindStringSubmatch(meter)
	}

	meterCore             := matchMeter[2]
	containsAnchorAtStart := (matchMeter[1] != "")
	containsAnchorAtEnd   := (matchMeter[3] != "")

	meterCore = strings.TrimSpace(meterCore)
	meterCore = multipleSpacesRegex.ReplaceAllString(meterCore, " ")

	meterPieces := strings.Split(meterCore, "")
	mungedMeter := strings.Join(meterPieces, `\s*`)

	mungedMeter = strings.Replace(mungedMeter, `0`, `[0\*]`,   -1)
	mungedMeter = strings.Replace(mungedMeter, `1`, `[12\*]`,  -1)
	mungedMeter = strings.Replace(mungedMeter, ` `, `\s+`,     -1)
	mungedMeter = strings.Replace(mungedMeter, `.`, `[012\*]`, -1)

	var secondaryR *regexp.Regexp
	if strings.Contains(meterCore, " ") {
		mungedMeterPieces := strings.Split(mungedMeter, `\s+`)
		mungedMeterPiecesCaptured := []string{}
		for _,mmp := range mungedMeterPieces {
			mungedMeterPiecesCaptured = append( mungedMeterPiecesCaptured, "(" + mmp + ")" )
		}
		mungedMeterPiecesCombined := strings.Join(mungedMeterPiecesCaptured, `\s+`)
		secondaryR = regexp.MustCompile(mungedMeterPiecesCombined)
	}


	before := "" 
	if containsAnchorAtStart  {
		before = "^" 
	} 

	after := ""
	if containsAnchorAtEnd {
		after = "$"
	}

	capturedMeter := before + `\s(` + mungedMeter + `)\s` + after

	r := regexp.MustCompile(capturedMeter)
	return r, secondaryR
}

const cropBeforeAfterToMaxWords = 5
const cropBeforeAfterDotDotDot  = "..."

type SecondaryMatch struct {
	SecondaryRegexp       *regexp.Regexp
	FullPhrase            string
	EmphasisRegexpMatches *[]string
	NumWordsInEachMatch   *[]int
	WordsInEachMatch      *[]*[]string
	PhraseInEachMatch     *[]string
	FinalWordWordInEachMatch *[]*Word
}

type MatchesOnMeter struct {
	Before   string
	During   string
	After    string
	NumWordsBefore int
	NumWordsDuring int
	NumWordsAfter  int
	NumWordsTotal  int
	BeforeCropped string
	AfterCropped  string
	FinalDuringWord       string
	FinalDuringSyllable   string
	FinalDuringSyllableAZ string
	FinalDuringWordWord *Word
	SecondaryMatch *SecondaryMatch
}

func ConstructSyllabi(sourceFilenames *[]string) (*Syllabi){
	if sourceFilenames == nil {
		sourceFilenames = &[]string{SyllableFilename}
	}

	words, numFragments, numSyllables := readSyllables(sourceFilenames)
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
		var stringAsKey string

		if k,ok := stringsAsKeys[s]; ok {
			stringAsKey = k
		} else {
			for _, pair := range nameTransformPairs {
				s = pair.Regexp.ReplaceAllString(s, pair.Replacement)
			}
			stringAsKey = strings.ToUpper(s)
		}

		if w,ok := (*words)[stringAsKey]; ok {
			word = w
		} else {
			if _,ok := knownUnknowns[stringAsKey]; ok {
				knownUnknowns[stringAsKey]++
			} else {
				knownUnknowns[stringAsKey] = 1
				fmt.Println("rhyme: findMatchingWord: new knownUnknown:", stringAsKey)
			} 

			word = &Word{
				Name:            s,
				FragmentsString: "X",
				Fragments:       []string{"X"},
				NumSyllables:    0,
				FinalSyllable:   "?",
				FinalSyllableAZ: "?",
				EmphasisPoints:  []string{"X"},
				EmphasisPointsString: "X",
				Unknown:         true,
			}
		}
		return word
	}

	findRhymes := func(s string) []string {
		matchingStrings := []string{}
		matchingWord := findMatchingWord(s)

		if ! matchingWord.Unknown {
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
		if ! w.Unknown {
			count = (*w).NumSyllables
		}

		return count
	}

	emphasisPoints := func(s string) []string {
		ep := []string{}
		w := findMatchingWord(s)
		if !w.Unknown {
			ep = (*w).EmphasisPoints
		}
		return ep
	}

	finalSyllableFunc := func(s string) string {
		fs := ""
		w := findMatchingWord(s)

		if !w.Unknown {
			fs = (*w).FinalSyllable
		}
		return fs
	}

	wordRegexps       = append(wordRegexps, `\w+`)
	wordRegexpsAsOrs := strings.Join(wordRegexps, "|")
	finalWordRegexp  := regexp.MustCompile(`(` + wordRegexpsAsOrs + `)\W*$`)
	wordsRegexp      := regexp.MustCompile(`(` + wordRegexpsAsOrs + `)`)

	finalSyllableOfPhraseFunc := func(s string) string {
		finalWord := ""
		matches := finalWordRegexp.FindStringSubmatch(s)
		if matches != nil {
			finalWord = matches[1]
		}

		fs := finalSyllableFunc(finalWord)
		return fs
	}

	findAllPhraseMatches := func(phrase string) *[][]string {
		matches := wordsRegexp.FindAllStringSubmatch(phrase, -1)
		return &matches
	}

	findAllEmphasisPointsDetails := func(phrase string) (*EmphasisPointsDetails) {
		phraseMatches                := findAllPhraseMatches(phrase)
		phraseWords                  := []string{}
		matchingWords                := []*Word{}
		emphasisPointsStrings        := []string{}
		emphasisPointsCombinedString := ""
		containsUnmatchedWord        := false
		finalSyllable                := ""
		finalSyllableAZ              := ""
		var finalMatchingWord *Word = nil

		if phraseMatches != nil && len(*phraseMatches) > 0 {
			for _, match := range *phraseMatches{
				phraseWord := match[1]
				phraseWords = append( phraseWords, phraseWord)
				matchingWord := findMatchingWord(phraseWord)
				emphasisPointsString := "X"
				if matchingWord.Unknown {
					containsUnmatchedWord = true
				} else {
					emphasisPointsString = matchingWord.EmphasisPointsString
				}

				matchingWords = append(matchingWords, matchingWord)
				emphasisPointsStrings = append( emphasisPointsStrings, emphasisPointsString)
			}

			finalMatchingWord = matchingWords[len(matchingWords)-1]; 
			if finalMatchingWord != nil {
				finalSyllable   = finalMatchingWord.FinalSyllable
				finalSyllableAZ = finalMatchingWord.FinalSyllableAZ
			}

			emphasisPointsCombinedString = " " + strings.Join(emphasisPointsStrings, " ") + " "
		}

		epd := EmphasisPointsDetails{
			Phrase: phrase,
			PhraseMatches: phraseMatches,
			PhraseWords: phraseWords,
			MatchingWords: matchingWords,
			ContainsUnmatchedWord: containsUnmatchedWord,
			EmphasisPointsStrings: emphasisPointsStrings,
			FinalSyllable: finalSyllable,
			FinalSyllableAZ: finalSyllableAZ,
			EmphasisPointsCombinedString: emphasisPointsCombinedString,
			FinalMatchingWord: finalMatchingWord,
		}

		return &epd
	}

	// reproduces functionality of func (*Regexp) FindAllIndex, but returns *all* fixed-length matches, including overlapping
	findAllIndexIncludingOverlapping := func( r *regexp.Regexp, s string ) ([][]int) {
		allMatches := [][]int{}
		startFrom  := 0

		matchOnPartial := func (r *regexp.Regexp, s string, i int) ([]int){
			partialS := s[i:len(s)]
			m := r.FindStringSubmatchIndex(partialS)
			return m
		}

		matches := matchOnPartial(r, s, startFrom)

		for matches != nil {
			adjustedMatch := []int{
				startFrom + matches[0],
				startFrom + matches[1],
			}

			allMatches = append(allMatches, adjustedMatch)
			startFrom = startFrom + matches[0] + 1
			if startFrom < len(s) {
				matches = matchOnPartial(r, s, startFrom)
			}
		}

		return allMatches
	}

	rhymeAndMetersOfPhrase := func(phrase string, emphasisRegexps ...*regexp.Regexp) (*[]*RhymeAndMeter) {

		emphasisRegexp               := emphasisRegexps[0]
		var emphasisRegexpSecondary *regexp.Regexp
		if len(emphasisRegexps)>1 {
			emphasisRegexpSecondary = emphasisRegexps[1]
		}

		// fmt.Println("rhyme.rhymeAndMetersOfPhrase: emphasisRegexpSecondary=", emphasisRegexpSecondary)

		emphasisPointsDetails        := findAllEmphasisPointsDetails( phrase )
		emphasisPointsCombinedString := emphasisPointsDetails.EmphasisPointsCombinedString
		phraseWords                  := emphasisPointsDetails.PhraseWords
		matchingWords                := emphasisPointsDetails.MatchingWords
		finalWord := "" // ????

		rams := []*RhymeAndMeter{}
		// allEmphasisRegexpIndexes := emphasisRegexp.FindAllStringIndex(emphasisPointsCombinedString, -1)
		allEmphasisRegexpIndexes := findAllIndexIncludingOverlapping(emphasisRegexp, emphasisPointsCombinedString)

		if allEmphasisRegexpIndexes != nil {
			for _,emphasisRegexpIndexes := range allEmphasisRegexpIndexes {

				emphasisRegexpMatches := []string{
					emphasisPointsCombinedString,
					emphasisPointsCombinedString[                       0 : emphasisRegexpIndexes[0]],
					emphasisPointsCombinedString[emphasisRegexpIndexes[0] : emphasisRegexpIndexes[1]],
					emphasisPointsCombinedString[emphasisRegexpIndexes[1] : len(emphasisPointsCombinedString)],
				}

				var emphasisRegexpMatch2 string
				if emphasisRegexpMatches == nil {
					emphasisRegexpMatch2 = ""
				} else {
					emphasisRegexpMatch2 = emphasisRegexpMatches[2]
				}

				matchesOnMeter := MatchesOnMeter{}
				if emphasisRegexpMatches != nil {
					countWordsInMatch := func(s string) int { 
						trimmed := strings.TrimSpace(s)
						num := 0
						if trimmed != "" {
							num = len( strings.Split( trimmed, " " ) ) 
						}
						return num
					}
					// assume we can rely on space-separated emphasis fragments to map to words...
					numBefore        := countWordsInMatch(emphasisRegexpMatches[1])
					numDuring        := countWordsInMatch(emphasisRegexpMatches[2])
					numAfter         := countWordsInMatch(emphasisRegexpMatches[3])
					numBeforeDuring  := numBefore + numDuring
					numTotal         := numBeforeDuring       + numAfter

					if numTotal != len(phraseWords) {
						fmt.Println("rhyme: rhymeAndMeterOfPhrase: matchesOnMeter: mismatched counts: numTotal=", numTotal, ", len(phraseWords)=", len(phraseWords))
					} else {
						matchBefore        := ""
						matchBeforeCropped := matchBefore
						if numBefore > 0 {
							matchBefore = strings.Join(phraseWords[0:numBefore], " ")
							numBeforeExcess := numBefore-cropBeforeAfterToMaxWords
							if numBeforeExcess > 0 {
								matchBeforeCropped = cropBeforeAfterDotDotDot + " " + strings.Join(phraseWords[numBeforeExcess:numBefore], " ")
							} else {
								matchBeforeCropped = matchBefore
							}
						}
						matchDuring := ""
						if numDuring > 0 {
							matchDuring = strings.Join(phraseWords[numBefore:numBeforeDuring], " ")
						}
						finalDuringWord       := phraseWords[numBeforeDuring-1]
						finalDuringSyllable   := finalSyllableFunc(finalDuringWord)
						finalDuringSyllableAZ := KeepAZString(finalDuringSyllable)
						finalDuringWordWord   := matchingWords[numBeforeDuring-1]

						matchAfter 		  := ""
						matchAfterCropped := matchAfter
						if numAfter > 0 {
							matchAfter = strings.Join(phraseWords[numBeforeDuring:len(phraseWords)], " ")
							numAfterExcess := numAfter-cropBeforeAfterToMaxWords
							if numAfterExcess > 0 {
								matchAfterCropped = strings.Join(phraseWords[numBeforeDuring:len(phraseWords)-numAfterExcess], " ") + " " + cropBeforeAfterDotDotDot
							} else {
								matchAfterCropped = matchAfter
							}
						}

// type SecondaryMatch struct {
// 	FullPhrase string
// 	EmphasisRegexpMatches *[]string
// 	NumWordsInEachMatch *[]int
// 	WordsInEachMatch *[]string
// 	PhraseInEachMatch *[]string
//  FinalWordWordInEachMatch *[]*Word
// }

						var secondaryMatch *SecondaryMatch

						// fmt.Println("rhymeAndMetersOfPhrase: matchDuring=", matchDuring)
						// fmt.Println("rhymeAndMetersOfPhrase: emphasisRegexpMatch2=", emphasisRegexpMatch2)
						// fmt.Println("rhymeAndMetersOfPhrase: emphasisRegexpSecondary=", emphasisRegexpSecondary)

						if emphasisRegexpSecondary != nil && emphasisRegexpMatch2 != "" {
							secondaryEmphMatches := emphasisRegexpSecondary.FindStringSubmatch( emphasisRegexpMatch2 )
							if secondaryEmphMatches != nil {
								secondaryEmphMatchesWordCounts := []int{}
								for i,sem := range secondaryEmphMatches {
									if i==0 { continue }
									c := countWordsInMatch(sem)
									secondaryEmphMatchesWordCounts = append( secondaryEmphMatchesWordCounts, c)
									// fmt.Println("rhymeAndMetersOfPhrase: secondaryEmphMatches: sem=", sem, ", c=", c)
								}
								countFrom := numBefore
								secondaryEmphMatchesWords := []*[]string{}
								finalWordWordInEachMatch := []*Word{}
								for _,c := range secondaryEmphMatchesWordCounts {
									ws := phraseWords[countFrom:(countFrom+c)]
									secondaryEmphMatchesWords = append(secondaryEmphMatchesWords, &ws)
									ww := matchingWords[countFrom+c-1]
									finalWordWordInEachMatch = append(finalWordWordInEachMatch, ww)
									countFrom = countFrom + c
								}
								secondaryEmphMatchesPhrases := []string{}
								for _,ws := range secondaryEmphMatchesWords {
									phrase := strings.Join(*ws, " ")
									secondaryEmphMatchesPhrases = append( secondaryEmphMatchesPhrases, phrase)
								}

								secondaryMatch = &SecondaryMatch{
									SecondaryRegexp:       emphasisRegexpSecondary,
									FullPhrase:            matchDuring,
									EmphasisRegexpMatches: &secondaryEmphMatches,
									NumWordsInEachMatch:   &secondaryEmphMatchesWordCounts,
									WordsInEachMatch:      &secondaryEmphMatchesWords,
									PhraseInEachMatch:     &secondaryEmphMatchesPhrases,
									FinalWordWordInEachMatch: &finalWordWordInEachMatch,
								}
							}
						}

						matchesOnMeter = MatchesOnMeter{
							Before:          matchBefore, 
							During:          matchDuring,
							After:           matchAfter,
							NumWordsBefore:  numBefore,
							NumWordsDuring:  numDuring,
							NumWordsAfter:   numAfter,
							NumWordsTotal:   numTotal,
							BeforeCropped:   matchBeforeCropped,
							AfterCropped:    matchAfterCropped,
							FinalDuringWord:       finalDuringWord,
							FinalDuringSyllable:   finalDuringSyllable,
							FinalDuringSyllableAZ: finalDuringSyllableAZ,
							FinalDuringWordWord: finalDuringWordWord,
							SecondaryMatch:  secondaryMatch,
						}
					}
				}
		 
				ram := RhymeAndMeter{
					Phrase:                       phrase,
					PhraseWords:                  &phraseWords,
					MatchingWords:                &emphasisPointsDetails.MatchingWords,
					EmphasisPointsStrings:        &emphasisPointsDetails.EmphasisPointsStrings,
					EmphasisPointsCombinedString: emphasisPointsCombinedString,
					FinalSyllable:                emphasisPointsDetails.FinalSyllable,
					FinalSyllableAZ:              emphasisPointsDetails.FinalSyllableAZ,
					ContainsUnmatchedWord:        emphasisPointsDetails.ContainsUnmatchedWord,
					FinalWord:                    finalWord,
					EmphasisRegexp:               emphasisRegexp,
					EmphasisRegexpString:         emphasisRegexp.String(),
					EmphasisRegexpMatches:        emphasisRegexpMatches,
					EmphasisRegexpMatch2:         emphasisRegexpMatch2,
					MatchesOnMeter:               &matchesOnMeter,
				}

				rams = append( rams, &ram )		
			}			
		}

		return &rams
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

	knownUnknownsFunc := func() *[]string {
		// 	knownUnknowns := map[string]int{}

		list := []string{}

		for k,_ := range knownUnknowns {
			list = append(list, k)
		}

		sort.Strings(list)

		return &list
	}

	syllabi := Syllabi{
		Stats:          stats,
		SourceFilenames: sourceFilenames,
		FindRhymes:     findRhymes,
		CountSyllables: countSyllables,
		EmphasisPoints: emphasisPoints,
		FinalSyllable:  finalSyllableFunc,
		FinalSyllableOfPhrase: finalSyllableOfPhraseFunc,
		SortPhrasesByFinalSyllable: sortPhrasesByFinalSyllable,
		RhymeAndMetersOfPhrase:      rhymeAndMetersOfPhrase,
		FindMatchingWord:           findMatchingWord,
		KnownUnknowns:              knownUnknownsFunc,
		PhraseWordsRegexp:            wordsRegexp,
		PhraseWordsRegexpString:      wordsRegexp.String(),
		FindAllEmphasisPointsDetails: findAllEmphasisPointsDetails,
	}

	return &syllabi
}

func main() {
	syllabi := ConstructSyllabi(&[]string{SyllableFilename})

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
