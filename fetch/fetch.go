// Fetches all the data about the critics and the movies they rated and stores them in a database

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
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

// Fetches a list of all available critics and places them in the given outFile
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

type genericError struct {
	Message string
}

func (err *genericError) Error() string { return err.Message }

// / converts a review JSON into a ReviewBatch instance
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

var USER_AGENTS = [15]string{
	"Mozilla/5.0 (Linux; Android 12; moto g stylus 5G)",
	"AppleWebKit/537.36 (KHTML, like Gecko)",
	"Chrome/112.0.0.0 Mobile Safari/537.36v",
	"Mozilla/5.0 (Linux; Android 10; MAR-LX1A)",
	"AppleWebKit/537.36 (KHTML, like Gecko)",
	"Chrome/112.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (iPhone9,4; U; CPU iPhone OS 10_0_1 like Mac OS X)",
	"AppleWebKit/602.1.50 (KHTML, like Gecko)",
	"Version/10.0 Mobile/14A403 Safari/602.1",
	"Mozilla/5.0 (Linux; Android 7.0; Pixel C Build/NRD90M; wv)",
	"AppleWebKit/537.36 (KHTML, like Gecko)",
	"Version/4.0 Chrome/52.0.2743.98 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
	"AppleWebKit/537.36 (KHTML, like Gecko)",
	"Chrome/42.0.2311.135 Safari/537.36 Edge/12.246",
}

func sendRequest(url string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	idx := rand.Int31n(15)
	userAgent := USER_AGENTS[idx]

	// add some user agent because without some spam filters kick in
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw_body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		// fmt.Println("FUCK!")
		// body := string(raw_body)
		// fmt.Println(body)
		return nil, &genericError{"Got bad status code!"}
	}

	return raw_body, nil
}

// Fetch the first batch of reviews of a critic
func getFirstBatch(critic *Critic) (*ReviewBatch, error) {

	// The first "movies" page of a critic has the reviews as a json hardcoded somewhere in the HTML.
	// THis line is what we're interested in.
	regexp, err := regexp.Compile("<script id=\"reviews-json\" .+?>({.+)</script>")
	if err != nil {
		return nil, err
	}

	const url = "https://www.rottentomatoes.com/critics/%s/movies"
	reqUrl := fmt.Sprintf(url, critic.url)
	raw_body, err := sendRequest(reqUrl)
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

// Get the ReviewBatch of the given critic after the provided afterCursor.
// afterCursor is used by RottenTomates for the pagination
func getBatch(critic *Critic, afterCursor string) (*ReviewBatch, error) {
	const url = "https://www.rottentomatoes.com/napi/critics/%s/movies?after=%s&pagecount=50"
	reqUrl := fmt.Sprintf(url, critic.url, afterCursor)

	raw_body, err := sendRequest(reqUrl)
	if err != nil {
		return nil, err
	}

	batch := parseReviewBatch(raw_body)
	return &batch, nil
}

// Fetch all the reviews of a given critic
func fetch_reviews(critic *Critic, verbose bool) ([]Review, error) {
	var reviews []Review

	if verbose {
		fmt.Print("\rLoad Review page 1...")
	}
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

// Worker that getch a slice of critics, fetches all their reviews and writes them into the outDir.
// Each time a critic is done, a bool is sent to the channel indicating success of failure for the critic
func fetch_worker(channel chan<- bool, critics []Critic, outDir string, failChan chan<- []Critic) {
	var failedCrititcs []Critic
	for _, critic := range critics {

		reviews, err := fetch_reviews(&critic, false)
		if err != nil {
			channel <- false
			failedCrititcs = append(failedCrititcs, critic)
			continue
		}
		if len(reviews) == 0 {
			channel <- false
			failedCrititcs = append(failedCrititcs, critic)
			continue
		}

		fileName := fmt.Sprintf("%s/%s.csv", outDir, critic.url)
		outFile, err := os.Create(fileName)
		if err != nil {
			channel <- false
			failedCrititcs = append(failedCrititcs, critic)
			continue
		}

		writtenLines := 0
		for _, review := range reviews {
			_, err := outFile.WriteString(fmt.Sprintf("%s\n", review.String()))
			if err != nil {
				fmt.Println(err)
				continue
			}
			writtenLines++
		}

		if writtenLines == 0 {
			failedCrititcs = append(failedCrititcs, critic)
			channel <- false
		} else {
			channel <- true
		}
	}

	failChan <- failedCrititcs
}

// Fetch the reviews of all the critivs in the criticsFile and write for each of the critics a file into outDir.
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
	defer close(channel)
	failChannel := make(chan []Critic, workers)
	defer close(failChannel)

	stepSize := int(math.Ceil(float64(len(critics)) / float64(workers)))

	for i := 0; i < workers; i++ {
		lower := i * stepSize
		upper := min(len(critics), lower+stepSize)

		go fetch_worker(channel, critics[lower:upper], outDir, failChannel)
	}

	doneTotal := 0
	finished := 0
	errors := 0

	for doneTotal < len(critics) {
		res := <-channel
		doneTotal++
		if res {
			finished++
		} else {
			errors++
		}

		if verbose && doneTotal&10 == 0 {
			fmt.Printf("\r%.2f%% done; %d finished; %d errors",
				100.0*float32(doneTotal)/float32(len(critics)),
				finished,
				errors)
		}
	}
	if verbose {
		fmt.Printf("\r%.2f%% done; %d finished; %d errors",
			100.0*float32(doneTotal)/float32(len(critics)),
			finished,
			errors)
	}

	fmt.Println("Critics where issues occured")
	for i := 0; i < workers; i++ {
		failed := <-failChannel

		for _, failedCritic := range failed {
			fmt.Println(failedCritic.url)
		}
	}
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
		fmt.Println("Available commands are: fetch-critics, fetch-reviews, fetch-all-reviews")
		os.Exit(1)
	}
}
