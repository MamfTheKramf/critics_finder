package main

import (
	"fmt"
	"os"

	"github.com/MamfTheKramf/critics_finder/internal/fetch"
	"github.com/MamfTheKramf/critics_finder/internal/normalize"
	"github.com/MamfTheKramf/critics_finder/internal/tui"
)

var argMap = make(map[string]func([]string))

func main() {
	argMap["tui"] = tui.StartTui
	argMap["fetch"] = fetch.FetchMain
	argMap["normalize"] = normalize.NormalizeMain

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Expect arguments")
		printUsage()
		os.Exit(1)
	}

	fn, prs := argMap[os.Args[1]]
	if !prs {
		fmt.Fprintf(os.Stderr, "Unknown command '%s'\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
	fn(os.Args[2:])
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Possible commands are:")
	for cmd := range argMap {
		fmt.Fprintf(os.Stderr, "- '%s'\n", cmd)
	}
}
