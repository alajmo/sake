package tui

import (
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/tui/components"
	"github.com/alajmo/sake/core/tui/misc"
	"github.com/alajmo/sake/core/tui/pages"
	"github.com/alajmo/sake/core/tui/views"
	"github.com/rivo/tview"
)

func createPages(
	servers []dao.Server,
	serverTags []string,
	tasks []dao.Task,
) *tview.Pages {
	appPages := tview.NewPages()
	navPane := createNav()
	search := components.CreateSearch()
	misc.Search = search

	runPage := pages.CreateRunPage(tasks, servers, serverTags)
	execPage := pages.CreateExecPage(servers, serverTags)
	serversPage := pages.CreateServersPage(servers, serverTags)
	tasksPage := pages.CreateTasksPage(tasks)

	misc.MainPage = tview.NewPages().
		AddPage("run", runPage, true, true).
		AddPage("exec", execPage, true, false).
		AddPage("servers", serversPage, true, false).
		AddPage("tasks", tasksPage, true, false)

	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(navPane, 2, 1, false).
		AddItem(misc.MainPage, 0, 1, true)
	appPages.AddPage("main", mainLayout, true, true)

	SwitchToPage("run")

	return appPages
}

func createNav() *tview.Flex {
	// Buttons
	misc.RunBtn = components.CreateButton("Run")
	misc.RunBtn.SetSelectedFunc(func() {
		SwitchToPage("run")
		misc.App.SetFocus(*misc.RunLastFocus)
	})

	misc.ExecBtn = components.CreateButton("Exec")
	misc.ExecBtn.SetSelectedFunc(func() {
		SwitchToPage("exec")
		misc.App.SetFocus(*misc.ExecLastFocus)
	})

	misc.ServerBtn = components.CreateButton("Servers")
	misc.ServerBtn.SetSelectedFunc(func() {
		SwitchToPage("servers")
		misc.App.SetFocus(*misc.ServersLastFocus)
	})

	misc.TaskBtn = components.CreateButton("Tasks")
	misc.TaskBtn.SetSelectedFunc(func() {
		SwitchToPage("tasks")
		misc.App.SetFocus(*misc.TasksLastFocus)
	})

	misc.HelpBtn = components.CreateButton("Help")
	misc.HelpBtn.SetSelectedFunc(func() {
		views.ShowHelpModal()
	})

	// Left
	left := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(misc.RunBtn, 7, 0, false).
		AddItem(misc.ExecBtn, 8, 0, false).
		AddItem(misc.ServerBtn, 11, 0, false).
		AddItem(misc.TaskBtn, 9, 0, false)

	// Right
	right := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(misc.HelpBtn, 8, 0, false)

	// Nav
	navPane := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(left, 0, 1, false).
		AddItem(nil, 0, 1, false).
		AddItem(right, 8, 0, false)
	navPane.SetBorderPadding(0, 1, 1, 1)

	return navPane
}

func SwitchToPage(pageName string) {
	misc.MainPage.SwitchToPage(pageName)

	switch pageName {
	case "servers":
		components.SetActiveButtonStyle(misc.ServerBtn)
		components.SetInactiveButtonStyle(misc.TaskBtn)
		components.SetInactiveButtonStyle(misc.RunBtn)
		components.SetInactiveButtonStyle(misc.ExecBtn)
		components.SetInactiveButtonStyle(misc.HelpBtn)
	case "tasks":
		components.SetActiveButtonStyle(misc.TaskBtn)
		components.SetInactiveButtonStyle(misc.ServerBtn)
		components.SetInactiveButtonStyle(misc.RunBtn)
		components.SetInactiveButtonStyle(misc.ExecBtn)
		components.SetInactiveButtonStyle(misc.HelpBtn)
	case "run":
		components.SetActiveButtonStyle(misc.RunBtn)
		components.SetInactiveButtonStyle(misc.ServerBtn)
		components.SetInactiveButtonStyle(misc.TaskBtn)
		components.SetInactiveButtonStyle(misc.ExecBtn)
		components.SetInactiveButtonStyle(misc.HelpBtn)
	case "exec":
		components.SetActiveButtonStyle(misc.ExecBtn)
		components.SetInactiveButtonStyle(misc.ServerBtn)
		components.SetInactiveButtonStyle(misc.TaskBtn)
		components.SetInactiveButtonStyle(misc.RunBtn)
		components.SetInactiveButtonStyle(misc.HelpBtn)
	}

	_, page := misc.MainPage.GetFrontPage()
	misc.App.SetFocus(page)
}
