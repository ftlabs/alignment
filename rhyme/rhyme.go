package main

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
}

var (
	syllableRegexp      = regexp.MustCompile(`^[A-Z]+\d+$`)
	finalSyllableRegexp = regexp.MustCompile(`([A-Z]+\d+(?:[^\d]*))$`)
)

func readSyllables(filename string) (*map[string]*Word) {

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

			numSyllables := 0
			for _,f := range fragments {
				if syllableRegexp.FindStringSubmatch(f) != nil {
					numSyllables = numSyllables + 1
	    		}
	    	}

	    	if numSyllables == 0 {
	    		fmt.Println("WARNING: no syllables found for name=", name) 
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
			}
		}
    }

    fmt.Println("num fragments = ", countFragments, ", num syllables = ", countSyllables) 

    return &words
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

type Syllabi struct {
    Words          *map[string]*Word
    FinalOnes      *map[string][]*Word
    SourceFilename string
    FindRhymesFunc func(string) []string
}

func ConstructSyllabi(sourceFilename string) (*Syllabi){
	words          := readSyllables(SyllableFilename)
	finalSyllables := processFinalSyllables(words)

	findRhymesFunc := func(s string) []string {
		upperS          := strings.ToUpper(s)
		matchingStrings := []string{}

	    // fmt.Println("ConstructSyllabi: upperS=", upperS ) 

		if matchingWord, ok := (*words)[upperS]; ok {
		     // fmt.Println("ConstructSyllabi: matchingWord=", matchingWord ) 
		     // fmt.Println("ConstructSyllabi: matchingWord.FragmentsString=", matchingWord.FragmentsString ) 
		     // fmt.Println("ConstructSyllabi: matchingWord.FinalSyllable=", matchingWord.FinalSyllable ) 
			finalSyllable := matchingWord.FinalSyllable
		 	if rhymingWords, ok := (*finalSyllables)[finalSyllable]; ok {
		 		for _,w := range rhymingWords {
		 			matchingStrings = append(matchingStrings, (*w).Name)
				}
			}
		}

		return matchingStrings
	}

	syllabi := Syllabi{
		Words:          words,
		FinalOnes:      finalSyllables,
		SourceFilename: SyllableFilename,
		FindRhymesFunc: findRhymesFunc,
	}

	return &syllabi
}

func main() {
	syllabi := ConstructSyllabi(SyllableFilename)

    fmt.Println("num words = ", len(*syllabi.Words) ) 
	fmt.Println("num unique final syllables = ", len(*syllabi.FinalOnes))

	s := "carnival"
	rhymesWith := (*syllabi).FindRhymesFunc(s)
	sort.Strings(rhymesWith)
	fmt.Println(`"`, s, `" rhymes with: `, strings.Join(rhymesWith, ","))
}
