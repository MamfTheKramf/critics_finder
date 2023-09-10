// Functions form normalizing review scores and collecting all movies
package normalize

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/critics_finder/utils"
)

var (
	singleNumRegExp = regexp.MustCompile(`^\d+(?:\.\d+)?$`)
	fractionRegexp  = regexp.MustCompile(`(\d+(?:\.\d+)?)/(\d+(?:\.\d+)?)`)
	gradesMap       = map[string]float32{
		"A+": 1.0,
		"+A": 1.0,
		"A":  0.9285714285714285,
		"-A": 0.8571428571428571,
		"A-": 0.8571428571428571,
		"B+": 0.7857142857142857,
		"+B": 0.7857142857142857,
		"B":  0.7142857142857142,
		"B-": 0.6428571428571428,
		"-B": 0.6428571428571428,
		"C+": 0.5714285714285714,
		"+C": 0.5714285714285714,
		"C":  0.5,
		"C-": 0.42857142857142855,
		"-C": 0.42857142857142855,
		"D+": 0.3571428571428571,
		"+D": 0.3571428571428571,
		"D":  0.2857142857142857,
		"D-": 0.21428571428571427,
		"-D": 0.21428571428571427,
		"F+": 0.14285714285714285,
		"+F": 0.14285714285714285,
		"F":  0.07142857142857142,
		"F-": 0.0,
		"-F": 0.0,
	}
)

func preprocessRating(rating string) string {
	inter := strings.ToUpper(rating)
	inter = strings.ReplaceAll(inter, "STARS", "")
	inter = strings.ReplaceAll(inter, "STAR", "")
	inter = strings.ReplaceAll(inter, "OUT OF", "/")
	inter = strings.ReplaceAll(inter, "OF", "/")
	inter = strings.ReplaceAll(inter, "\\", "/")
	inter = strings.ReplaceAll(inter, "-MINUS", "-")
	inter = strings.ReplaceAll(inter, "-PLUS", "+")
	inter = strings.ReplaceAll(inter, " ", "")
	return inter
}

// Normalizes the given rating. Rating can either be in fraction form (e.g. 4.5/10) or in school grades (e.g. B+)
func normalizeRating(rating string) (float32, error) {
	processed := preprocessRating(rating)
	match := singleNumRegExp.FindStringSubmatch(processed)
	if match != nil {
		num, numErr := strconv.ParseFloat(match[0], 32)
		if numErr == nil {
			denom := 5.0
			if num >= 5.0 {
				if num <= 10. {
					denom = 10.0
				} else if num <= 100. {
					denom = 100.0
				}
			}

			return float32(num / denom), nil
		}
	}
	match = fractionRegexp.FindStringSubmatch(processed)
	if match != nil {
		num, numErr := strconv.ParseFloat(match[1], 32)
		denom, denomErr := strconv.ParseFloat(match[2], 32)
		if numErr == nil && denomErr == nil {
			return float32(num / denom), nil
		}
	}
	// wasn't a rating check for grade
	score, prs := gradesMap[processed]
	if !prs {
		return 0.0, fmt.Errorf("couldn't normalize rating '%s'", rating)
	}
	return score, nil
}

func normalizeReviews(reviewFile, outDir string) (int, error) {
	errors := strings.Builder{}
	emptyScores := 0

	reviews := utils.ReadStructs[utils.Review](reviewFile, false)
	for _, review := range reviews {
		if review.Score == "" {
			emptyScores++
			continue
		}
		_, err := normalizeRating(review.Score)
		if err != nil {
			errors.WriteString(err.Error())
			errors.WriteString("\n")
		}
	}

	if errors.Len() > 0 {
		return emptyScores, fmt.Errorf(errors.String())
	}
	return emptyScores, nil
}

// normalizes each review inside each of the review files and writes them to a new file in outDir
func normalizeWorker(channel chan<- bool, reviewFiles []os.DirEntry, inDir, outDir string, emptyScoresChannel chan<- int) {
	totalEmptyScores := 0
	for _, reviewFile := range reviewFiles {
		path := path.Join(inDir, reviewFile.Name())
		emptyScores, err := normalizeReviews(path, outDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
		channel <- err == nil
		totalEmptyScores += emptyScores
	}

	emptyScoresChannel <- totalEmptyScores
}

func NormalizeMain(args []string) {
	var inDir = flag.String("i", "./tmp/reviews", "Path to the directory containing the reviews")
	var outDir = flag.String("o", "./tmp/normalized", "Path to the directory to write normalized reviews to")
	var moviesFile = flag.String("m", "./tmp/movies.gob", "Path to file to store movies in")
	var workers = flag.Int("w", 1, "Number of workers to normalize reviews")
	fmt.Println(os.Args)
	fmt.Println(args)
	os.Args = append(os.Args[:1], args...)
	fmt.Println(os.Args)
	flag.Parse()

	fmt.Println(*inDir, *outDir, *moviesFile, *workers)

	entries, err := os.ReadDir(*inDir)
	if err != nil {
		panic(err)
	}

	progressChannel := make(chan bool, 1)
	defer close(progressChannel)
	emptyScoresChannel := make(chan int, *workers)
	defer close(emptyScoresChannel)

	stepSize := int(math.Ceil(float64(len(entries)) / float64(*workers)))

	for i := 0; i < *workers; i++ {
		lower := i * stepSize
		upper := utils.Min(len(entries), lower+stepSize)

		go normalizeWorker(progressChannel, entries[lower:upper], *inDir, *outDir, emptyScoresChannel)
	}

	doneTotal := 0
	finished := 0
	errors := 0

	for doneTotal < len(entries) {
		res := <-progressChannel
		doneTotal++
		if res {
			finished++
		} else {
			errors++
		}

		if doneTotal&10 == 0 {
			fmt.Printf("%.2f%% done; %d finished; %d errors\r",
				100.0*float32(doneTotal)/float32(len(entries)),
				finished,
				errors)
		}
	}
	fmt.Printf("%.2f%% done; %d finished; %d errors\n",
		100.0*float32(doneTotal)/float32(len(entries)),
		finished,
		errors)

	totalEmptyScores := 0
	for i := 0; i < *workers; i++ {
		totalEmptyScores += <-emptyScoresChannel
	}

	fmt.Printf("totalEmptyScores: %d\n", totalEmptyScores)
}
