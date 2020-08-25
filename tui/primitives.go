package tui

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func getPipelines(ui *UI) tview.Primitive {
	pipelines, err := ui.glClient.GetPipelines(ui.projectID)
	if err != nil {
		log.Fatal(err)
	}

	root := tview.NewTreeNode("Refresh Pipelines").
		// Set ID of 0 to signify root so we know when we hit it
		SetReference(pipelineReference{
			ID:         0,
			isPipeline: false}).
		SetColor(tcell.ColorRed)

	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	for _, v := range pipelines {
		var line strings.Builder
		// Do some conversions
		sID := strconv.Itoa(v.ID)
		fmtDate := v.CreatedAt.Format("Mon Jan _2 15:04:05 2006")

		line.WriteString(hyphenate(sID))
		line.WriteString(hyphenate(fmtDate))
		line.WriteString(hyphenate(v.Status))
		line.WriteString(hyphenate(v.Ref))
		line.WriteString(v.User.Name)

		node := tview.NewTreeNode(line.String()).
			SetReference(pipelineReference{
				ID:         v.ID,
				isPipeline: true,
			}).
			SetExpanded(false).
			SetSelectable(true)
		root.AddChild(node)
	}

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		// Need to modify type as GetReference returns interface{}
		reference := node.GetReference().(pipelineReference)
		if reference.ID == 0 {
			go func(ui *UI) {
				ui.app.QueueUpdateDraw(func() {
					pipelines := getPipelines(ui)
					displayHeader(ui.namespace, pipelineSidebar, ui, pipelines, "pipelines")
				})
			}(ui)
			displayLoadingScreen(ui)
		}
		children := node.GetChildren()
		if len(children) == 0 {
			if reference.isPipeline {
				jobs := getJobs(ui, reference.ID)
				for _, job := range jobs {
					node.AddChild(job)
				}
				node.SetExpanded(!node.IsExpanded())
			} else {
				ui.currentJob = reference.ID
				ui.currentNodeText = node.GetText()
				traceFile := getTrace(ui, ui.currentJob)
				displayHeader(ui.currentNodeText, traceSidebar, ui, traceFile, "trace")
			}
		} else {
			// Collapse if visible, expand if collapsed.
			node.SetExpanded(!node.IsExpanded())
		}
	})

	return tree

}

func getJobs(ui *UI, pipelineID int) []*tview.TreeNode {
	jobs, err := ui.glClient.GetPipelineJobs(ui.projectID, pipelineID)
	if err != nil {
		log.Fatal(err)
	}

	var jobsNodes []*tview.TreeNode

	for _, v := range jobs {
		var line strings.Builder
		// Do some conversions
		sID := strconv.Itoa(v.ID)
		fmtDate := v.CreatedAt.Format("Mon Jan _2 15:04:05 2006")

		line.WriteString(hyphenate(sID))
		line.WriteString(hyphenate(fmtDate))
		line.WriteString(hyphenate(v.Stage))
		line.WriteString(v.Status)

		node := tview.NewTreeNode(line.String()).
			SetReference(pipelineReference{
				ID:         v.ID,
				isPipeline: false,
			}).
			SetExpanded(false).
			SetSelectable(true)
		jobsNodes = append(jobsNodes, node)
	}

	return jobsNodes

}

func getTrace(ui *UI, jobID int) tview.Primitive {
	trace, err := ui.glClient.GetJobTrace(ui.projectID, jobID)
	if err != nil {
		log.Fatal(err)
	}

	var s string
	if trace == nil {
		s = "No trace found"
	} else {
		buf := new(bytes.Buffer)
		buf.ReadFrom(trace)
		s = buf.String()
	}

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			ui.app.Draw()
		})

	fmt.Fprintf(textView, "%s ", s)

	return textView

}

func setupLoadingScreen(ui *UI) {
	loadingPrimitive := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("\n\nLoading Data")

	ui.pages.AddPage("loading", loadingPrimitive, true, true)
}
