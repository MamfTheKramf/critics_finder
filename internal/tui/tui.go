package tui

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/MamfTheKramf/critics_finder/internal/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sahilm/fuzzy"
)

const contentLabel = "content"
const modalLabel = "model"
const instructions = `Use the input field to find media to rate.
After one is selected, press [ENTER] to open a window to enter the rating.

Already rated media is shown on the left.
Select one of them and press [ENTER] to update it (alternatively you can search it again on the right and input a new rating).
Pressing [BACKSPACE] will remove the rating from the list.`

var app = tview.NewApplication()

// To be able to display rating modal
var layers = tview.NewPages()

var ratingModal = tview.NewFlex()
var modalPrompt = tview.NewTextView()
var modalForm = tview.NewForm()

var content = tview.NewFlex()
var mainSections = tview.NewFlex()
var ratedMediaSection = tview.NewFlex()
var selectMediaSection = tview.NewFlex()
var selectedSection = selectMediaSection // reference to selection that is currently selected
var searchQuery = tview.NewInputField()
var controls = tview.NewTextView()

var userRatings []utils.NumericReview
var critics []utils.Critic
var criticsRatings = make(map[string][]utils.NumericReview)

// critic ratings are read in a separate routing -> we need an indicator that we're done before comparing
var doneReadingcriticsRatings = make(chan bool, 1)
var media []utils.Media
var selected utils.Media
var currRating = 0.0

// list of media names used for fuzzy finding
var mediaNames []string

var outFile *os.File

func StartTui(args []string) {
	file, err := os.Create("./tmp/deineMudda.txt")
	if err != nil {
		panic(err)
	}
	outFile = file
	userRatingsFile := flag.String("u", utils.DefaultUserRatingsFile, "Path to the user ratings file (if non-existing it will be created)")
	criticsFile := flag.String("c", utils.DefaultCriticsFile, "Path to crtics file")
	inDir := flag.String("i", utils.DefaultNormalizedDir, "Path to directory containing normalized reviews")
	mediaFile := flag.String("m", utils.DefaultMediaFile, "Path to media file")
	os.Args = append(os.Args[:1], args...)
	flag.Parse()

	setup(*userRatingsFile, *criticsFile, *inDir, *mediaFile)

	defer writeUserRatings(*userRatingsFile)

	if err := app.SetRoot(layers, true).EnableMouse(true).SetFocus(searchQuery).Run(); err != nil {
		panic(err)
	}
}

func writeUserRatings(outFile string) {
	utils.WriteStructs[utils.NumericReview](userRatings, outFile, false)
}

func setup(userRatingsFile, criticsFile, criticsRatingDir, mediaFile string) {
	readUserRatings(userRatingsFile)
	readCritics(criticsFile)
	readMedia(mediaFile)

	go readCriticsRatings(criticsRatingDir)

	setupApp()
}

func modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}

func setupApp() {
	controls.SetBackgroundColor(tcell.ColorLightGray)
	controls.SetTextColor(tcell.ColorBlack)
	controls.SetText("(Shift + ArrowKey) focus different window; (Ctrl + c) exit")

	inputLabel := tview.NewTextView().SetText("Media Title:")
	inputLabel.SetTextColor(tcell.ColorGreen)
	searchQuery.SetAutocompleteFunc(autocomplete)
	searchQuery.SetFieldWidth(0)
	searchQuery.SetDoneFunc(func(_ tcell.Key) { selectMedium() })

	ratedMediaSection.SetBorder(true)
	ratedMediaSection.SetTitle(" Rated Media ")
	ratedMediaSection.SetTitleAlign(tview.AlignLeft)
	selectMediaSection.SetDirection(tview.FlexRow)
	selectMediaSection.SetBorderPadding(0, 0, 1, 1)
	selectMediaSection.SetBorder(true)
	selectMediaSection.SetTitle(" Select Media To Rate ")
	selectMediaSection.SetTitleAlign(tview.AlignLeft)
	selectMediaSection.AddItem(inputLabel, 1, 0, false)
	selectMediaSection.AddItem(searchQuery, 0, 3, true)
	selectMediaSection.AddItem(tview.NewTextView().SetText(instructions), 0, 1, false)

	mainSections.AddItem(ratedMediaSection, 0, 1, true)
	mainSections.AddItem(selectMediaSection, 0, 1, false)
	mainSections.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Modifiers()&tcell.ModShift != 0 {
			if event.Key() == tcell.KeyLeft {
				selectedSection = ratedMediaSection
			}
			if event.Key() == tcell.KeyRight {
				selectedSection = selectMediaSection
			}
			app.SetFocus(selectedSection)
		}
		return event
	})

	content.SetDirection(tview.FlexRow)
	content.AddItem(mainSections, 0, 1, true)
	content.AddItem(controls, 1, 0, false)

	ratingModal.SetBorder(true)
	ratingModal.SetTitle(" Rate given Medium ")
	ratingModal.SetDirection(tview.FlexRow)
	ratingModal.AddItem(modalPrompt, 0, 2, false)
	ratingModal.AddItem(modalForm, 0, 1, true)

	modalPrompt.SetTextAlign(tview.AlignCenter)
	modalPrompt.SetBorderPadding(1, 1, 1, 1)

	layers.AddPage(contentLabel, content, true, true)
	layers.AddPage(modalLabel, modal(ratingModal, 80, 20), true, false)
}

func autocomplete(currText string) []string {
	matches := fuzzy.Find(currText, mediaNames)
	var ret []string
	for _, match := range matches {
		ret = append(ret, match.Str)
	}
	return ret
}

func submitRating() {
	addRating()
	showContent()
}

// checks that currText is a float an within [0; 100]
func checkFloat(currText string, lastChar rune) bool {
	if !tview.InputFieldFloat(currText, lastChar) {
		return false
	}
	val, err := strconv.ParseFloat(currText, 64)
	if err != nil {
		return false
	}

	return val >= 0.0 && val <= 100.0
}

// checks if the selected medium exists.
// if it does, the modal windows will be opened.
// else we go back to the search query
func selectMedium() {
	val := searchQuery.GetText()
	found := false

	for _, candidate := range media {
		if val == getAutocompleteVal(candidate) {
			selected = candidate
			found = true
			break
		}
	}
	if !found {
		app.SetFocus(searchQuery)
		return
	}

	modalPrompt.Clear()
	modalPrompt.SetText(val)

	modalForm.AddInputField(" Score from 0 to 100 (decimals are allowed): ", "", 10, checkFloat, func(rating string) {
		parsed, err := strconv.ParseFloat(rating, 32)
		if err != nil {
			return
		}
		currRating = parsed
	})
	modalForm.AddButton("Submit", submitRating)

	layers.SwitchToPage(modalLabel)
}

func showUserRatings() {
	ratedMediaSection.Clear()
	li := tview.NewList()
	for _, userRating := range userRatings {
		li.AddItem(userRating.MediaUrl, fmt.Sprintf("    Score: %.2f", userRating.Score), ' ', nil)
	}

	li.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyDEL || event.Key() == tcell.KeyBackspace {
			idx := uint(li.GetCurrentItem())
			pruned, err := utils.Remove[utils.NumericReview](userRatings, idx)
			if err == nil {
				userRatings = pruned
				showUserRatings()
				app.SetFocus(selectedSection)
			}
		}
		return event
	})

	ratedMediaSection.AddItem(li, 0, 1, true)
}

// Go back to selectRating and ratingOverview view
func showContent() {
	showUserRatings()
	layers.SwitchToPage(contentLabel)
	app.SetFocus(selectedSection)
}

// add new user rating or update existing one
func addRating() {
	outFile.WriteString("Add Rating\n")
	defer func() {
		searchQuery.SetText("")
		modalPrompt.Clear()
		modalForm.Clear(true)
	}()

	for i, candidate := range userRatings {
		if selected.MediaUrl == candidate.MediaUrl {
			userRatings[i].Score = float32(currRating)
			outFile.WriteString(fmt.Sprintf("Found existing rating: %v", candidate))
			return
		}
	}
	outFile.WriteString("Create new rating")

	userRatings = append(userRatings, utils.NumericReview{
		MediaUrl: selected.MediaUrl,
		Score:    float32(currRating),
	})
	outFile.WriteString("Created new rating")

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
	showUserRatings()
}

func readCritics(criticsFile string) {
	fmt.Println("Reading critics...")
	critics = append(critics, utils.ReadStructs[utils.Critic](criticsFile, false)...)
	fmt.Printf("Read critics. Have %d critics now\n", len(critics))
}

func getAutocompleteVal(medium utils.Media) string {
	return fmt.Sprintf("%s (%s)", medium.MediaTitle, medium.MediaUrl)
}

func readMedia(mediaFile string) {
	fmt.Println("Reading media...")
	media = append(media, utils.ReadStructs[utils.Media](mediaFile, false)...)
	fmt.Printf("Read media. Have %d medias now\n", len(media))
	for _, medium := range media {
		mediaNames = append(mediaNames, getAutocompleteVal(medium))
	}
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
	doneReadingcriticsRatings <- true
}
