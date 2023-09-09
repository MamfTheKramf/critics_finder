// Functions form normalizing review scores and collecting all movies
package normalize

import "fmt"

func NormalizeMain(args []string) {
	fmt.Println("HeyHo")
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
