package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type Critic struct {
	Name string
	Url  string
}

func (c Critic) String() string {
	return fmt.Sprintf("%s, %s", c.Name, c.Url)
}

// Reads all the critics from a given file
func ReadCritics(criticsFile string, verbose bool) []Critic {
	inFile, err := os.Open(criticsFile)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)

	var critics []Critic
	var skippedLines = 0

	if verbose {
		fmt.Println("Scanning critics file...")
	}
	for scanner.Scan() {
		criticsLine := strings.Split(scanner.Text(), ",")

		if len(criticsLine) < 2 {
			skippedLines += 1
			continue
		}

		critics = append(critics, Critic{
			Name: strings.TrimSpace(criticsLine[0]),
			Url:  strings.TrimSpace(criticsLine[1]),
		})
	}

	if verbose {
		fmt.Printf("Found %d critics\n", len(critics))
		fmt.Printf("Skipped %d lines\n", skippedLines)
	}

	return critics
}
