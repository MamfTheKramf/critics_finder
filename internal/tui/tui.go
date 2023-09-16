package tui

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/MamfTheKramf/critics_finder/internal/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var app = tview.NewApplication()
var content = tview.NewFlex()
var mainSections = tview.NewFlex()
var ratedMediaSection = tview.NewBox()
var selectMediaSection = tview.NewBox()
var controls = tview.NewTextView()

var userRatings []utils.NumericReview
var critics []utils.Critic
var criticsRatings = make(map[string][]utils.NumericReview)
var doneReadingcriticsRatings = false
var media []utils.Media

func StartTui(args []string) {
	userRatingsFile := flag.String("u", utils.DefaultUserRatingsFile, "Path to the user ratings file (if non-existing it will be created)")
	criticsFile := flag.String("c", utils.DefaultCriticsFile, "Path to crtics file")
	inDir := flag.String("i", utils.DefaultNormalizedDir, "Path to directory containing normalized reviews")
	mediaFile := flag.String("m", utils.DefaultMediaFile, "Path to media file")
	os.Args = append(os.Args[:1], args...)
	flag.Parse()

	setup(*userRatingsFile, *criticsFile, *inDir, *mediaFile)

	if err := app.SetRoot(content, true).Run(); err != nil {
		panic(err)
	}
}

func setup(userRatingsFile, criticsFile, criticsRatingDir, mediaFile string) {
	readUserRatings(userRatingsFile)
	readCritics(criticsFile)
	readMedia(mediaFile)

	go readCriticsRatings(criticsRatingDir)

	setupApp()
}

func setupApp() {
	controls.SetBackgroundColor(tcell.ColorLightGray)
	controls.SetTextColor(tcell.ColorBlack)
	controls.SetText("(Shift + ArrowKey) focus different window; (Ctrl + c) exit")

	ratedMediaSection.SetBorder(true)
	ratedMediaSection.SetTitle("Rated Media")
	ratedMediaSection.SetTitleAlign(tview.AlignLeft)
	selectMediaSection.SetBorder(true)
	selectMediaSection.SetTitle("Select Media To Rate")
	selectMediaSection.SetTitleAlign(tview.AlignLeft)

	mainSections.AddItem(ratedMediaSection, 0, 1, false)
	mainSections.AddItem(selectMediaSection, 0, 1, false)

	content.SetDirection(tview.FlexRow)
	content.AddItem(mainSections, 0, 1, true)
	content.AddItem(controls, 1, 0, false)
}

func readUserRatings(ratingsFile string) {
	if _, err := os.Stat(ratingsFile); err != nil {
		fmt.Printf("Creating empty ratingsFile, since %s doesn't exist...\nNo ratings so far.\n", ratingsFile)
		if _, err := os.Create(ratingsFile); err != nil {
			panic(err)
		}
		return
	}
	fmt.Printf("Reading user ratings from %s\n", ratingsFile)
	readRatings := utils.ReadStructs[utils.NumericReview](ratingsFile, false)
	userRatings = append(userRatings, readRatings...)

	fmt.Printf("Read user ratings. Have %d ratings now\n", len(userRatings))
}

func readCritics(criticsFile string) {
	fmt.Println("Reading critics...")
	critics = append(critics, utils.ReadStructs[utils.Critic](criticsFile, false)...)
	fmt.Printf("Read critics. Have %d critics now\n", len(critics))
}

func readMedia(mediaFile string) {
	fmt.Println("Reading media...")
	media = append(media, utils.ReadStructs[utils.Media](mediaFile, false)...)
	fmt.Printf("Read media. Have %d medias now\n", len(media))
}

func readCriticsRatings(ratingsDir string) {
	entries, err := os.ReadDir(ratingsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading critics dir %s:\n", ratingsDir)
		panic(err)
	}
	for _, entry := range entries {
		filePath := path.Join(ratingsDir, entry.Name())
		criticUrl := strings.ReplaceAll(entry.Name(), ".gob", "")

		reviews := utils.ReadStructs[utils.NumericReview](filePath, false)

		criticsRatings[criticUrl] = reviews
	}
	doneReadingcriticsRatings = true
}
