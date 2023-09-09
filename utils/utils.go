package utils

import (
	"encoding/gob"
	"fmt"
	"io"
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

func WriteCritics(critics []Critic, outFile string) {
	fo, err := os.Create(outFile)
	if err != nil {
		fmt.Println(critics)
		panic(err)
	}
	defer fo.Close()

	enc := gob.NewEncoder(fo)

	for idx, c := range critics {
		if idx%10 == 0 {
			fmt.Printf("\rWriting to file: %.2f%%", float32(idx)/float32(len(critics)))
		}

		err := enc.Encode(c)
		if err != nil {
			fmt.Printf("\rCouldn't write %s\n", c.String())
			continue
		}
	}
	fmt.Println("\r Writing to file: 100%")
}

// Reads all the critics from a given file
func ReadCritics(criticsFile string, verbose bool) []Critic {
	inFile, err := os.Open(criticsFile)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	dec := gob.NewDecoder(inFile)

	var critics []Critic
	var skippedLines = 0

	if verbose {
		fmt.Println("Scanning critics file...")
	}
	for {
		var c Critic
		err := dec.Decode(&c)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			skippedLines++
			continue
		}

		critics = append(critics, c)
	}

	if verbose {
		fmt.Printf("Found %d critics\n", len(critics))
		fmt.Printf("Skipped %d lines\n", skippedLines)
	}

	return critics
}

type Review struct {
	Score      string
	MediaTitle string
	MediaInfo  string
	MediaUrl   string
}

func (r Review) String() string {
	return fmt.Sprintf("%s;%s;%s;%s",
		strings.ReplaceAll(r.Score, ";", "\\;"),
		strings.ReplaceAll(r.MediaTitle, ";", "\\;"),
		strings.ReplaceAll(r.MediaInfo, ";", "\\;"),
		strings.ReplaceAll(r.MediaUrl, ";", "\\;"))
}
