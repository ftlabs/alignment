package main

import (
    "bufio"
    "os"
    "strings"
    "fmt"
)

const (
	SyllableFilename = "./cmudict-0.7b"
)


func readSyllables(filename string) (*map[string][]string) {

	words := map[string][]string{}
	countSyllables := 0

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
			syllables        := strings.Split(remainder, " ")
			countSyllables = countSyllables + len(syllables)
			words[name] = syllables
		}
    }

    fmt.Println("num words = ", len(words), ", num syllables = ", countSyllables) 

    return &words
}

func main() {
	readSyllables(SyllableFilename)
}

