package pages

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/tui/misc"
	"github.com/alajmo/sake/core/tui/views"
)

type TTaskPage struct {
	focusable []*misc.TItem
}

func CreateTasksPage(tasks []dao.Task) *tview.Flex {
	p := &TTaskPage{}

	// Data
	taskData := views.CreateTasksData(
		tasks,
		[]string{"Task", "Name", "Description"},
		1,
		true,
		true,
		false,
	)

	// Shortcut info view
	infoView := views.CreateTasksInfoView()

	// Pages for table/tree toggle
	taskViewPages := p.createTaskViewPages(taskData)

	// Page
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(taskViewPages, 0, 1, true).
		AddItem(infoView, 1, 0, false).
		AddItem(misc.Search, 1, 0, false)

	// Focusable
	p.focusable = p.updateTaskFocusable(taskData)
	misc.TasksLastFocus = &p.focusable[0].Primitive

	// Shortcuts
	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if misc.App.GetFocus() == misc.Search {
			return event
		}

		switch event.Key() {
		case tcell.KeyTab:
			nextPrimitive := misc.FocusNext(p.focusable)
			misc.TasksLastFocus = nextPrimitive
			return nil
		case tcell.KeyBacktab:
			nextPrimitive := misc.FocusPrevious(p.focusable)
			misc.TasksLastFocus = nextPrimitive
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'C': // Clear filters
				taskData.Emitter.PublishAndWait(misc.Event{Name: "remove_task_filter", Data: ""})
				taskData.Emitter.PublishAndWait(misc.Event{Name: "remove_task_selections", Data: ""})
				taskData.Emitter.Publish(misc.Event{Name: "filter_tasks", Data: ""})
				return nil
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				misc.FocusPage(event, p.focusable)
				return nil
			}
		}
		return event
	})

	return page
}

func (p *TTaskPage) createTaskViewPages(taskData *views.TTask) *tview.Flex {
	isTable := taskData.TaskStyle == "task-table"

	pages := tview.NewPages().
		AddPage("task-table", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(taskData.TaskTableView.Root, 0, 1, true), true, isTable).
		AddPage("task-tree", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(taskData.TaskTreeView.Root, 0, 1, false), true, !isTable)

	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true)

	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if misc.App.GetFocus() == misc.Search {
			return event
		}

		switch event.Key() {
		case tcell.KeyCtrlE:
			if taskData.TaskStyle == "task-table" {
				taskData.TaskStyle = "task-tree"
			} else {
				taskData.TaskStyle = "task-table"
			}
			pages.SwitchToPage(taskData.TaskStyle)
			p.focusable = p.updateTaskFocusable(taskData)
			misc.App.SetFocus(p.focusable[0].Primitive)
			misc.TasksLastFocus = &p.focusable[0].Primitive
			return nil
		}
		return event
	})

	return page
}

func (p *TTaskPage) updateTaskFocusable(data *views.TTask) []*misc.TItem {
	focusable := []*misc.TItem{}

	if data.TaskStyle == "task-table" {
		focusable = append(
			focusable,
			misc.GetTUIItem(
				data.TaskTableView.Table,
				data.TaskTableView.Table.Box,
			))
	} else {
		focusable = append(
			focusable,
			misc.GetTUIItem(
				data.TaskTreeView.Tree,
				data.TaskTreeView.Tree.Box,
			))
	}

	return focusable
}
