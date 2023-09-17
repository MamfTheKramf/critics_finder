package tui

import (
	"math"
	"slices"

	"github.com/MamfTheKramf/critics_finder/internal/utils"
)

// contains a critic together with a score describing how close the critic rates movies to the reference ratings.
type ScoredCritic struct {
	Score  float64
	Critic utils.Critic
}

// Compares the ratings of each critic with the userRatings and assigns each critic a score.
// smaller scores are better. Critics that didn't rate any of the movies rated by the user get a score of infinity
// The returned slice is already sorted
func evaluate(userRatings []utils.NumericReview, criticsRatings map[string][]utils.NumericReview, critics []utils.Critic, workers int) []ScoredCritic {
	resultsChannel := make(chan []ScoredCritic, workers)
	defer close(resultsChannel)

	stepSize := int(math.Ceil(float64(len(critics)) / float64(workers)))
	for i := 0; i < workers; i++ {
		lower := i * stepSize
		upper := utils.Min(len(critics), lower+stepSize)

		go evalWorker(resultsChannel, userRatings, criticsRatings, critics[lower:upper])
	}

	var scoredCritics []ScoredCritic
	for i := 0; i < workers; i++ {
		workerRes := <-resultsChannel
		scoredCritics = append(scoredCritics, workerRes...)
	}

	slices.SortFunc(scoredCritics, func(a, b ScoredCritic) int {
		if a.Score < b.Score {
			return -1
		} else if a.Score > b.Score {
			return 1
		}
		return 0
	})

	return scoredCritics
}

// Scores each critics ratings against the user ratings and writes the ScoredCritics to the resChannel.
// The returned slice is not sorted
func evalWorker(resChannel chan []ScoredCritic, userRatings []utils.NumericReview, criticsRatings map[string][]utils.NumericReview, critics []utils.Critic) {
	scoredCritics := make([]ScoredCritic, 0, len(critics))

	for _, critic := range critics {
		score := math.Inf(1)
		criticRatings, prs := criticsRatings[critic.Url]
		if prs {
			score = eval(userRatings, criticRatings)
		}

		scoredCritics = append(scoredCritics, ScoredCritic{Score: score, Critic: critic})
	}

	resChannel <- scoredCritics
}

func eval(userRatings, criticRatings []utils.NumericReview) float64 {
	totalErr := 0.0
	totalMatches := 0
	for _, userRating := range userRatings {
		matchIdx := slices.IndexFunc(criticRatings, func(candidate utils.NumericReview) bool { return candidate.MediaUrl == userRating.MediaUrl })
		if matchIdx < 0 {
			continue
		}

		criticRating := criticRatings[matchIdx]

		totalErr += math.Pow(float64(userRating.Score)-float64(criticRating.Score), 2.0)
		totalMatches++
	}

	if totalMatches == 0 {
		return math.Inf(1)
	}
	return totalErr / float64(totalMatches)
}
