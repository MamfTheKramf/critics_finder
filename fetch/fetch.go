// Fetches all the data about the critics and the movies they rated and stores them in a database

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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

type Review struct {
	score      string
	mediaTitle string
	mediaInfo  string
	mediaUrl   string
}

func (r Review) String() string {
	return fmt.Sprintf("%s;%s;%s;%s",
		strings.ReplaceAll(r.score, ";", "\\;"),
		strings.ReplaceAll(r.mediaTitle, ";", "\\;"),
		strings.ReplaceAll(r.mediaInfo, ";", "\\;"),
		strings.ReplaceAll(r.mediaUrl, ";", "\\;"))
}

type ReviewBatch struct {
	reviews []Review
	next    string
	prev    string
}

type pageInfo struct {
	HasNextPage     bool
	HasPreviousPage bool
	StartCursor     string
	EndCursor       string
}

type rawReview struct {
	OriginalScore string
	MediaInfo     string
	MediaTitle    string
	MediaUrl      string
}

type rawResp struct {
	PageInfo pageInfo
	Reviews  []rawReview
}

func parseReviewBatch(json_raw []byte) ReviewBatch {
	res := rawResp{}
	json.Unmarshal(json_raw, &res)

	var reviews []Review

	for _, rev := range res.Reviews {
		reviews = append(reviews, Review{
			score:      rev.OriginalScore,
			mediaTitle: rev.MediaTitle,
			mediaInfo:  rev.MediaInfo,
			mediaUrl:   rev.MediaUrl,
		})
	}

	batch := ReviewBatch{
		reviews: reviews,
		next:    res.PageInfo.EndCursor,
		prev:    res.PageInfo.StartCursor,
	}

	return batch
}

func getFirstBatch(critic *Critic) (*ReviewBatch, error) {
	regexp, err := regexp.Compile("<script id=\"reviews-json\" .+>(.+)</script>")
	if err != nil {
		return nil, err
	}

	const url = "https://www.rottentomatoes.com/critics/%s/movies"

	resp, err := http.Get(fmt.Sprintf(url, critic.url))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw_body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	body := string(raw_body)

	match := regexp.FindStringSubmatch(body)
	if len(match) < 2 {
		return &ReviewBatch{
			reviews: []Review{},
		}, nil
	}

	json_raw := []byte(match[1])

	batch := parseReviewBatch(json_raw)

	return &batch, nil
}

func getBatch(critic *Critic, afterCursor string) (*ReviewBatch, error) {
	const url = "https://www.rottentomatoes.com/napi/critics/%s/movies?after=%s&pagecount=50"

	resp, err := http.Get(fmt.Sprintf(url, critic.url, afterCursor))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw_body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	batch := parseReviewBatch(raw_body)
	return &batch, nil
}

func fetch_reviews(critic *Critic, verbose bool) ([]Review, error) {
	var reviews []Review
	batch, err := getFirstBatch(critic)
	if err != nil {
		return nil, err
	}

	reviews = batch.reviews
	var page_count = 2

	var next = batch.next
	for len(next) > 0 {
		if verbose {
			fmt.Printf("\rLoad Review page %d...", page_count)
		}
		page_count += 1

		batch, err = getBatch(critic, next)
		if err != nil {
			fmt.Println(err)
			break
		}

		reviews = append(reviews, batch.reviews...)
		next = batch.next
	}
	if verbose {
		fmt.Println()
	}

	return reviews, nil
}

func fetch_worker(channel chan<- bool, critics []Critic, outDir string) {
	for _, critic := range critics {
		fileName := fmt.Sprintf("%s/%s.csv", outDir, critic.url)
		outFile, err := os.Create(fileName)
		if err != nil {
			fmt.Println("Couldn't create file")
			fmt.Println(err)
			channel <- false
			continue
		}

		reviews, err := fetch_reviews(&critic, false)
		if err != nil {
			channel <- false
			continue
		}

		for _, review := range reviews {
			_, err := outFile.WriteString(fmt.Sprintf("%s\n", review.String()))
			if err != nil {
				continue
			}
		}

		channel <- true
	}
}

func fetch_all_reviews(criticsFile, outDir string, workers int, verbose bool) {
	// Still some issues with this one, but good enough
	err := os.MkdirAll(outDir, os.ModePerm)
	if err != nil {
		panic(err)
	}

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
			name: strings.TrimSpace(criticsLine[0]),
			url:  strings.TrimSpace(criticsLine[1]),
		})
	}

	if verbose {
		fmt.Printf("Found %d critics\n", len(critics))
		fmt.Printf("Skipped %d lines\n", skippedLines)
	}

	// set up workers
	channel := make(chan bool, 1)

	stepSize := int(math.Ceil(float64(len(critics)) / float64(workers)))

	for i := 0; i < workers; i++ {
		lower := i * stepSize
		upper := min(len(critics), lower+stepSize)

		go fetch_worker(channel, critics[lower:upper], outDir)
	}

	doneTotal := 0
	finished := 0
	errors := 0

	for i := 0; i < len(critics); i++ {
		res := <-channel
		doneTotal += 1
		if res {
			finished += 1
		} else {
			errors += 1
		}

		if verbose && doneTotal&10 == 0 {
			fmt.Printf("\r%.2f%% done; %d finished; %d errors",
				100.0*float32(doneTotal)/float32(len(critics)),
				finished,
				errors)
		}
	}

	close(channel)
}

func main() {
	fetchCriticsSet := flag.NewFlagSet("fetch-critics", flag.ExitOnError)
	var outFile = fetchCriticsSet.String("o", "./tmp/out.txt", "Path to the out-file")

	fetchReviewsSet := flag.NewFlagSet("fetch-reviews", flag.ExitOnError)
	var criticUrl = fetchReviewsSet.String("c", "", "URL of critic to get reviews from")

	fetchAllReviewsSet := flag.NewFlagSet("fetch-all-reviews", flag.ExitOnError)
	var criticsFile = fetchAllReviewsSet.String("i", "./tmp/critics.csv", "Path to critics file (CSV)")
	var outDir = fetchAllReviewsSet.String("o", "./tmp/reviews", "Path to output directory (will be created if doesn't exist)")
	var workers = fetchAllReviewsSet.Int("w", 1, "Number of workers to fetch all reviews")

	if len(os.Args) < 2 {
		fmt.Println("Expect arguments")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "fetch-critics":
		fetchCriticsSet.Parse(os.Args[2:])
		fetch_critics(*outFile)
	case "fetch-reviews":
		fetchReviewsSet.Parse(os.Args[2:])
		reviews, err := fetch_reviews(&Critic{name: "", url: *criticUrl}, true)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Found %d reviews\n", len(reviews))
		for _, rev := range reviews[:min(len(reviews), 10)] {
			fmt.Println(rev.String())
		}
	case "fetch-all-reviews":
		fetchAllReviewsSet.Parse(os.Args[2:])
		fetch_all_reviews(*criticsFile, *outDir, *workers, true)
	default:
		fmt.Printf("Unkown command \"%s\"\n", os.Args[1])
		os.Exit(1)
	}
}
