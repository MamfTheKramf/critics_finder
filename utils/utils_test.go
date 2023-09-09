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

	WriteCritics(expectedCriritcs, f.Name())

	critics := ReadCritics(f.Name(), true)

	if len(critics) != len(expectedCriritcs) {
		t.Fatalf("Expected %d critics. Got %d", len(expectedCriritcs), len(critics))
	}

	for idx := range critics {
		actual := critics[idx]
		expected := expectedCriritcs[idx]

		compCritics(t, actual, expected, fmt.Sprintf("c%d", idx))
	}
}
