// Fetches all the data about the critics and the movies they rated and stores them in a database

package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
)

type Critic struct {
	name string
	url  string
}

func (c Critic) String() string {
	return fmt.Sprintf("%s, %s", c.name, c.url)
}

func fetch_critics(outFile string) {
	fmt.Println("Fetching critics...")

	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	const url = "https://www.rottentomatoes.com/critics/authors?letter=%s"

	// collect them
	var critics []Critic

	regExp, err := regexp.Compile("<a class=\"critic-authors__name\" href=\"/critics/(.+)\" data-qa=\"critic-item-link\">(.+)</a>")
	if err != nil {
		panic(err)
	}

	for _, letter := range alphabet {
		fmt.Printf("\rCritics of letter %s", string(letter))

		resp, err := http.Get(fmt.Sprintf(url, string(letter)))
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		raw_body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			continue
		}

		body := string(raw_body)

		matches := regExp.FindAllStringSubmatch(body, -1)

		for _, match := range matches {
			critics = append(critics, Critic{name: match[2], url: match[1]})
		}
	}

	fmt.Printf("\rFound %d critics.\n", len(critics))

	// write them to a file

	fo, err := os.Create(outFile)
	if err != nil {
		fmt.Println(critics)
		panic(err)
	}

	for idx, c := range critics {
		if idx%10 == 0 {
			fmt.Printf("\rWriting to file: %.2f%%", float32(idx)/float32(len(critics)))
		}

		_, err := fo.WriteString(fmt.Sprintf("%s\n", c.String()))
		if err != nil {
			fmt.Printf("\rCouldn't write %s\n", c.String())
			continue
		}
	}
	fmt.Println("\r Writing to file: 100%")

}

func main() {
	fetchCriticsSet := flag.NewFlagSet("fetch-critics", flag.ExitOnError)

	var outFile = fetchCriticsSet.String("o", "./tmp/out.txt", "Path to the out-file")

	if len(os.Args) < 2 {
		fmt.Println("Expect arguments")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "fetch-critics":
		fetchCriticsSet.Parse(os.Args[2:])
		fetch_critics(*outFile)
	default:
		fmt.Printf("Unkown command \"%s\"\n", os.Args[1])
		os.Exit(1)
	}
}
