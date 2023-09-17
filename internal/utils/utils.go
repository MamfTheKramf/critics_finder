package utils

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	DefaultCriticsFile     = "./tmp/critics.gob"
	DefaultReviewsDir      = "./tmp/reviews"
	DefaultNormalizedDir   = "./tmp/normalized"
	DefaultMediaFile       = "./tmp/movies.gob"
	DefaultUserRatingsFile = "./tmp/userRatings.gob"
)

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Retuns a new slice containing the elements of arr exept for the one with the given index
func Remove[T any](arr []T, idx uint) ([]T, error) {
	intIdx := int(idx)
	if intIdx >= len(arr) {
		return nil, fmt.Errorf("index out of range! Slice has len %d, biut got %d", len(arr), idx)
	}

	if intIdx == len(arr)-1 {
		return arr[:idx], nil
	}

	var ret []T
	for i, val := range arr {
		if i != intIdx {
			ret = append(ret, val)
		}
	}
	return ret, nil
}

type Critic struct {
	Name string
	Url  string
}

func (c Critic) String() string {
	return fmt.Sprintf("%s, %s", c.Name, c.Url)
}

func WriteStructs[T fmt.Stringer](structs []T, outFile string, verbose bool) int {
	fo, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer fo.Close()

	enc := gob.NewEncoder(fo)

	writtenStructs := 0
	for idx, s := range structs {
		if verbose && idx%10 == 0 {
			fmt.Printf("\rWriting to file: %.2f%%", float32(idx)/float32(len(structs)))
		}

		err := enc.Encode(s)
		if err != nil {
			if verbose {
				fmt.Printf("\rCouldn't write struct %s\n", s.String())
			}
			continue
		}

		writtenStructs++
	}
	if verbose {
		fmt.Println("\r Writing to file: 100%")
	}

	return writtenStructs
}

// Reads all the critics from a given file
func ReadStructs[T any](filePath string, verbose bool) []T {
	inFile, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	dec := gob.NewDecoder(inFile)

	var structs []T
	var skippedStructs = 0

	if verbose {
		fmt.Println("Scanning structs file...")
	}
	for {
		var s T
		err := dec.Decode(&s)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			skippedStructs++
			continue
		}

		structs = append(structs, s)
	}

	if verbose {
		fmt.Printf("Found %d structs\n", len(structs))
		fmt.Printf("Skipped %d structs\n", skippedStructs)
	}

	return structs
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

type NumericReview struct {
	Score    float32
	MediaUrl string
}

func (r NumericReview) String() string {
	return fmt.Sprintf("%f;%s",
		r.Score,
		strings.ReplaceAll(r.MediaUrl, ";", "\\;"))
}

type Media struct {
	MediaTitle string
	MediaInfo  string
	MediaUrl   string
}

func (m Media) String() string {
	return fmt.Sprintf("%s;%s;%s",
		strings.ReplaceAll(m.MediaTitle, ";", "\\;"),
		strings.ReplaceAll(m.MediaInfo, ";", "\\;"),
		strings.ReplaceAll(m.MediaUrl, ";", "\\;"))
}
