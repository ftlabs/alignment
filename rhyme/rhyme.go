package main

import (
    "bufio"
    "os"
    "strings"
    "fmt"
    "regexp"
)

const (
	SyllableFilename = "./cmudict-0.7b"
)

type Word struct {
    Name           string
    FragmentString string
    Fragments      []string
    NumSyllables   int
    FinalSyllable  string
}

func readSyllables(filename string) (*map[string]Word) {

	words := map[string]Word{}
	countFragments      := 0
	countSyllables      := 0
    syllableRegexp      := regexp.MustCompile(`^[A-Z]+\d+$`)
    finalSyllableRegexp := regexp.MustCompile(`([A-Z]+\d+(?:[^\d]*))$`)

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
			words[name] = Word{
				Name:           name,
				FragmentString: remainder,
				Fragments:      fragments,
				NumSyllables:   numSyllables,
				FinalSyllable:  finalSyllable,
			}
		}
    }

    fmt.Println("num words = ", len(words), ", num fragments = ", countFragments, ", num syllables = ", countSyllables) 

    return &words
}

func main() {
	readSyllables(SyllableFilename)
}

