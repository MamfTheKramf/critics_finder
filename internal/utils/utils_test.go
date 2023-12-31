package utils

import (
	"fmt"
	"os"
	"testing"
)

func compCritics(t *testing.T, actual, expected Critic, name string) {
	if actual.Name != expected.Name {
		t.Errorf("Wrong Critic-Name for %s. Expected '%s'. Got '%s'.", name, actual.Name, expected.Name)
	}
	if actual.Url != expected.Url {
		t.Errorf("Wrong Critic-Url for %s. Expected '%s'. Got '%s'.", name, actual.Url, expected.Url)
	}
}

func TestSerDe(t *testing.T) {
	expectedCriritcs := make([]Critic, 2)
	expectedCriritcs[0] = Critic{
		Name: "Hutzi",
		Url:  "Butzi",
	}
	expectedCriritcs[1] = Critic{
		Name: "Butzi",
		Url:  "Hutzi",
	}

	f, err := os.CreateTemp("", "testSerDe")
	if err != nil {
		t.Fatalf("Can't create temp file: %v\n", err)
	}
	defer os.Remove(f.Name())

	WriteStructs(expectedCriritcs, f.Name(), false)

	critics := ReadStructs[Critic](f.Name(), false)

	if len(critics) != len(expectedCriritcs) {
		t.Fatalf("Expected %d critics. Got %d", len(expectedCriritcs), len(critics))
	}

	for idx := range critics {
		actual := critics[idx]
		expected := expectedCriritcs[idx]

		compCritics(t, actual, expected, fmt.Sprintf("c%d", idx))
	}
}

func arrComp(t *testing.T, expected, actual []int) {
	if len(expected) != len(actual) {
		t.Errorf("expected: %v\ngot: %v", expected, actual)
	}
	for idx, expectedVal := range expected {
		actualVal := actual[idx]
		if expectedVal != actualVal {
			t.Errorf("expected: %v\ngot: %v", expected, actual)
		}
	}
}

func TestRemove(t *testing.T) {
	input := []int{0, 1, 2, 3, 4, 5, 6}

	actual, err := Remove[int](input, 0)
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
	arrComp(t, []int{1, 2, 3, 4, 5, 6}, actual)

	actual, err = Remove[int](input, 4)
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
	arrComp(t, []int{0, 1, 2, 3, 5, 6}, actual)

	actual, err = Remove[int](input, 6)
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
	arrComp(t, []int{0, 1, 2, 3, 4, 5}, actual)

	actual, err = Remove[int](input, 7)
	if err == nil {
		t.Errorf("expected error but got %v", actual)
	}
}
