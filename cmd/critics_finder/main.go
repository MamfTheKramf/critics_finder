package main

import (
	"fmt"
	"os"

	"github.com/MamfTheKramf/critics_finder/internal/fetch"
	"github.com/MamfTheKramf/critics_finder/internal/normalize"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Expect arguments")
		os.Exit(1)
	}

	argMap := make(map[string]func([]string))
	argMap["fetch"] = fetch.FetchMain
	argMap["normalize"] = normalize.NormalizeMain

	fn, prs := argMap[os.Args[1]]
	if !prs {
		fmt.Fprintf(os.Stderr, "Unknown command '%s'\nPossible commands are:\n", os.Args[1])
		for cmd := range argMap {
			fmt.Fprintf(os.Stderr, "- '%s'\n", cmd)
		}
		os.Exit(1)
	}
	fn(os.Args[2:])
}
