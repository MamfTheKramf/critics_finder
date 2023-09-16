package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var app = tview.NewApplication()
var content = tview.NewFlex()
var ratedMediaSection = tview.NewBox()
var selectMediaSection = tview.NewBox()
var controls = tview.NewTextView()

func StartTui(args []string) {
	setup()

	if err := app.SetRoot(content, true).Run(); err != nil {
		panic(err)
	}
}

func setup() {
	controls.SetBackgroundColor(tcell.ColorLightGray)
	controls.SetTextColor(tcell.ColorBlack)
	controls.SetText("(Shift + ArrowKey) focus different window; (Ctrl + c) exit")

	ratedMediaSection.SetBorder(true)
	ratedMediaSection.SetTitle("Rated Media")
	ratedMediaSection.SetTitleAlign(tview.AlignLeft)
	selectMediaSection.SetBorder(true)
	selectMediaSection.SetTitle("Select Media To Rate")
	selectMediaSection.SetTitleAlign(tview.AlignLeft)

	content.SetDirection(tview.FlexRow)
	content.AddItem(ratedMediaSection, 0, 1, false)
	content.AddItem(selectMediaSection, 0, 1, false)
	content.AddItem(controls, 1, 0, false)
}
