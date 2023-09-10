package normalize

import (
	"math"
	"testing"
)

func TestNormalizeFraction(t *testing.T) {
	ratings := []string{
		"5/5",
		"3.2/10",
		"10.0/10.0",
		"100 / 100",
		"0.0/20",
	}
	expectedVals := []float32{
		5.0 / 5.0,
		3.2 / 10.,
		10.0 / 10.0,
		100.0 / 100.0,
		0.0 / 20.0,
	}

	eps := 0.00001

	for idx, rating := range ratings {
		expected := expectedVals[idx]
		actual, err := normalizeRating(rating)
		if err != nil {
			t.Errorf("Couldn't normalize '%s'", rating)
		}
		if math.Abs(float64(expected-actual)) > eps {
			t.Errorf("Expected %f for rating '%s'. Got %f", expected, rating, actual)
		}
	}
}

func TestNormalizeGrade(t *testing.T) {
	ratings := []string{
		"A",
		"b+",
		"c",
		"-F",
	}
	expectedVals := []float32{
		gradesMap["A"],
		gradesMap["+B"],
		gradesMap["C"],
		gradesMap["-F"],
	}

	eps := 0.000001

	for idx, rating := range ratings {
		expected := expectedVals[idx]
		actual, err := normalizeRating(rating)
		if err != nil {
			t.Errorf("%v", err)
		}
		if math.Abs(float64(expected-actual)) > eps {
			t.Errorf("Expected %f for rating '%s'. Got %f", expected, rating, actual)
		}
	}

}
