package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	gitlab "gctl/gitlab"
)

// UI represents a user interface
type UI struct {
	namespace       string
	app             *tview.Application
	glClient        *gitlab.Client
	projectID       int
	pages           *tview.Pages
	currentJob      int
	currentNodeText string
	currentPage     string
}

type pipelineReference struct {
	ID         int
	isPipeline bool
}

// New creates a new tui
func New(glClient *gitlab.Client, namespace string, projectID int) {
	app := tview.NewApplication()

	ui := UI{
		namespace: namespace,
		app:       app,
		glClient:  glClient,
		projectID: projectID,
	}

	// So we can control our app effectively, we need to define some global keys
	setGlobalKeys(&ui)

	// We use the pages functionality of the library so we can switch back and forth between screens
	ui.pages = tview.NewPages()

	// We need to show a loading screen as data can take a while to pull for the pipelines
	setupLoadingScreen(&ui)

	// This allows us to get data in the background before displaying it
	go func(ui *UI) {
		ui.app.QueueUpdateDraw(func() {
			pipelines := getPipelines(ui)
			displayHeader(ui.namespace, pipelineSidebar, ui, pipelines, "pipelines")
		})
	}(&ui)

	// Now we need to display the loading screen
	displayLoadingScreen(&ui)
}

func displayHeader(header string, sidebarText string, ui *UI, primitive tview.Primitive, page string) {
	var headerText strings.Builder
	headerText.WriteString(header)

	title := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(headerText.String())

	sidebar := tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetText(sidebarText)

	grid := tview.NewGrid().
		SetRows(3, 0).
		SetColumns(40, 0).
		SetBorders(true).
		AddItem(title, 0, 0, 1, 2, 0, 0, false).
		AddItem(sidebar, 1, 0, 1, 1, 0, 0, false)

	grid.AddItem(primitive, 1, 1, 1, 1, 0, 0, true)

	ui.pages.AddPage(page, grid, true, true)
	ui.pages.SwitchToPage(page)
	ui.currentPage = page
	if err := ui.app.SetRoot(ui.pages, true).Run(); err != nil {
		panic(err)
	}
}

func displayLoadingScreen(ui *UI) {
	ui.pages.SwitchToPage("loading")
	if err := ui.app.SetRoot(ui.pages, true).Run(); err != nil {
		panic(err)
	}
}

// hyphenate strings
func hyphenate(in string) string {
	out := fmt.Sprintf("%s - ", in)
	return out
}

func setGlobalKeys(ui *UI) {
	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			ui.pages.SwitchToPage("pipelines")
		case tcell.KeyCtrlC:
			//Escape properly
			ui.app.Stop()
			os.Exit(0)
		case tcell.KeyCtrlR:
			//Refresh Trace
			traceFile := getTrace(ui, ui.currentJob)
			displayHeader(ui.currentNodeText, traceSidebar, ui, traceFile, "trace")
		case tcell.KeyCtrlK:
			//Cancel job
			if ui.currentPage == "trace" {
				_, _ = ui.glClient.CancelJob(ui.projectID, ui.currentJob)
				traceFile := getTrace(ui, ui.currentJob)
				displayHeader(ui.currentNodeText, traceSidebar, ui, traceFile, "trace")
			}
		case tcell.KeyCtrlT:
			//Retry job
			if ui.currentPage == "trace" {
				_, _ = ui.glClient.RetryJob(ui.projectID, ui.currentJob)
				traceFile := getTrace(ui, ui.currentJob)
				displayHeader(ui.currentNodeText, traceSidebar, ui, traceFile, "trace")
			}
		case tcell.KeyCtrlE:
			//Execute job
			if ui.currentPage == "trace" {
				_, _ = ui.glClient.RunJob(ui.projectID, ui.currentJob)
				traceFile := getTrace(ui, ui.currentJob)
				displayHeader(ui.currentNodeText, traceSidebar, ui, traceFile, "trace")
			}
		}
		return event
	})
}
