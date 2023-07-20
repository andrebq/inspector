package dashboard

import (
	_ "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func Run() error {
	app := tview.NewApplication()
	requests := tview.NewList().SetBorder(true).SetTitle("Requests").SetTitleAlign(tview.AlignLeft)

	layout := tview.NewFlex()
	layout.AddItem(requests, 50, 1, true)
	app.SetRoot(layout, true)

	return app.Run()
}
