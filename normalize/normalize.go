// Functions form normalizing review scores and collecting all movies
package normalize

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	fractionRegexp = regexp.MustCompile(`(\d+(?:\.\d+)?)/(\d+(?:\.\d+)?)`)
	gradesMap      = map[string]float32{
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

// Normalizes the given rating. Rating can either be in fraction form (e.g. 4.5/10) or in school grades (e.g. B+)
func normalizeRating(rating string) (float32, error) {
	trimmed := strings.ReplaceAll(rating, " ", "")
	match := fractionRegexp.FindStringSubmatch(trimmed)
	if match != nil {
		num, numErr := strconv.ParseFloat(match[1], 32)
		denom, denomErr := strconv.ParseFloat(match[2], 32)
		if numErr == nil && denomErr == nil {
			return float32(num / denom), nil
		}
	}
	// wasn't a rating check for grade
	uppercase := strings.ToUpper(trimmed)
	score, prs := gradesMap[uppercase]
	if !prs {
		return 0.0, fmt.Errorf("couldn't normalize rating '%s'", rating)
	}
	return score, nil
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
	// fetchCriticsSet := flag.NewFlagSet(FETCH_CRITICS, flag.ExitOnError)
	// var outFile = fetchCriticsSet.String("o", "./tmp/critics.csv", "Path to the out-file")

	// fetchReviewsSet := flag.NewFlagSet(FETCH_REVIEWS, flag.ExitOnError)
	// var criticUrl = fetchReviewsSet.String("c", "", "URL of critic to get reviews from")

	// fetchAllReviewsSet := flag.NewFlagSet(FETCH_ALL_REVIEWS, flag.ExitOnError)
	// var criticsFile = fetchAllReviewsSet.String("i", "./tmp/critics.csv", "Path to critics file (CSV)")
	// var outDir = fetchAllReviewsSet.String("o", "./tmp/reviews", "Path to output directory (will be created if doesn't exist)")
	// var workers = fetchAllReviewsSet.Int("w", 1, "Number of workers to fetch all reviews")

	// if len(args) < 1 {
	// 	fmt.Fprintf(os.Stderr, "Expect arguments")
	// 	os.Exit(1)
	// }

	// switch args[0] {
	// case FETCH_CRITICS:
	// 	fetchCriticsSet.Parse(args[1:])
	// 	fetch_critics(*outFile)
	// case FETCH_REVIEWS:
	// 	fetchReviewsSet.Parse(args[1:])
	// 	reviews, err := fetch_reviews(&Critic{Name: "", Url: *criticUrl}, true)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	fmt.Printf("Found %d reviews\n", len(reviews))
	// 	for _, rev := range reviews[:utils.Min(len(reviews), 10)] {
	// 		fmt.Println(rev.String())
	// 	}
	// case FETCH_ALL_REVIEWS:
	// 	fetchAllReviewsSet.Parse(args[1:])
	// 	fetch_all_reviews(*criticsFile, *outDir, *workers, true)
	// default:
	// 	fmt.Printf("Unkown command \"%s\"\n", args[0])
	// 	fmt.Printf("Available commands are: %s, %s, %s\n", FETCH_CRITICS, FETCH_REVIEWS, FETCH_ALL_REVIEWS)
	// 	os.Exit(1)
	// }
}
