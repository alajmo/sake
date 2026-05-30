package pages

import (
	"fmt"
	"io"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/run"
	"github.com/alajmo/sake/core/tui/components"
	"github.com/alajmo/sake/core/tui/misc"
	"github.com/alajmo/sake/core/tui/views"
)

type TRunPage struct {
	focusable  []*misc.TItem
	taskData   *views.TTask
	serverData *views.TServer
	outputView *components.TOutput
	spec       *views.TSpec
}

func CreateRunPage(
	tasks []dao.Task,
	servers []dao.Server,
	serverTags []string,
) *tview.Flex {
	r := &TRunPage{}

	// Data
	r.taskData = views.CreateTasksData(
		tasks,
		[]string{"Task", "Name", "Description"},
		1,
		true,
		true,
		true,
	)

	r.serverData = views.CreateServersData(
		servers,
		serverTags,
		[]string{"Server", "Host", "Tags"},
		2,
		true,
		true,
		true,
		len(serverTags) > 0,
	)

	// Views
	r.outputView = &components.TOutput{Title: "Output"}
	r.outputView.Create()

	// Spec options
	r.spec = views.CreateSpecView()

	// Shortcut info views
	runInfoView := views.CreateRunInfoView()
	execInfoView := views.CreateExecInfoView()

	// Selection page (tasks and servers)
	selectionPage := r.createSelectionPage(runInfoView)

	// Output page with info
	outputPage := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(r.outputView.Root, 0, 1, true).
		AddItem(execInfoView, 1, 0, false)

	// Pages container
	pages := tview.NewPages().
		AddPage("run-selection", selectionPage, true, true).
		AddPage("run-output", outputPage, true, false)

	// Main page
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		AddItem(misc.Search, 1, 0, false)

	// Focus
	r.focusable = r.updateSelectionFocusable()
	misc.RunLastFocus = &r.focusable[0].Primitive

	// Shortcuts
	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			r.focusable = r.switchView(pages)
			misc.App.SetFocus(r.focusable[0].Primitive)
			misc.RunLastFocus = &r.focusable[0].Primitive
			return nil
		case tcell.KeyCtrlR:
			r.focusable = r.switchBeforeRun(pages)
			misc.App.SetFocus(r.focusable[0].Primitive)
			misc.RunLastFocus = &r.focusable[0].Primitive
			r.runTasks()
			return nil
		case tcell.KeyTab:
			nextPrimitive := misc.FocusNext(r.focusable)
			misc.RunLastFocus = nextPrimitive
			return nil
		case tcell.KeyBacktab:
			nextPrimitive := misc.FocusPrevious(r.focusable)
			misc.RunLastFocus = nextPrimitive
			return nil
		case tcell.KeyCtrlO:
			components.OpenModal("spec-modal", "Options", r.spec.View, 35, 14)
			return nil
		case tcell.KeyCtrlX:
			r.outputView.Clear()
			return nil
		case tcell.KeyRune:
			if _, ok := misc.App.GetFocus().(*tview.InputField); ok {
				return event
			}
			name, _ := pages.GetFrontPage()
			if name == "run-selection" {
				switch event.Rune() {
				case 'd': // Toggle describe modal
					if components.CloseDescribeModal() {
						return nil
					}
					r.describeItem()
					return nil
				case 'C': // Clear filters
					r.serverData.Emitter.PublishAndWait(misc.Event{Name: "remove_tag_filter", Data: ""})
					r.serverData.Emitter.PublishAndWait(misc.Event{Name: "remove_tag_selections", Data: ""})
					r.serverData.Emitter.PublishAndWait(misc.Event{Name: "remove_server_filter", Data: ""})
					r.serverData.Emitter.PublishAndWait(misc.Event{Name: "remove_server_selections", Data: ""})
					r.serverData.Emitter.Publish(misc.Event{Name: "filter_servers", Data: ""})

					r.taskData.Emitter.PublishAndWait(misc.Event{Name: "remove_task_filter", Data: ""})
					r.taskData.Emitter.PublishAndWait(misc.Event{Name: "remove_task_selections", Data: ""})
					r.taskData.Emitter.Publish(misc.Event{Name: "filter_tasks", Data: ""})
					return nil
				case '1', '2', '3', '4', '5', '6', '7', '8', '9':
					misc.FocusPage(event, r.focusable)
					return nil
				}
			}
		}

		return event
	})

	return page
}

func (r *TRunPage) createSelectionPage(info *tview.TextView) *tview.Flex {
	// Left: Tasks with table/tree toggle
	isTaskTable := r.taskData.TaskStyle == "task-table"
	taskPages := tview.NewPages().
		AddPage("task-table", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(r.taskData.TaskTableView.Root, 0, 1, true), true, isTaskTable).
		AddPage("task-tree", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(r.taskData.TaskTreeView.Root, 0, 1, false), true, !isTaskTable)

	taskPane := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(taskPages, 0, 1, true)

	// Handle Ctrl+E for task view toggle
	taskPane.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if misc.App.GetFocus() == misc.Search {
			return event
		}

		switch event.Key() {
		case tcell.KeyCtrlE:
			if r.taskData.TaskStyle == "task-table" {
				r.taskData.TaskStyle = "task-tree"
			} else {
				r.taskData.TaskStyle = "task-table"
			}
			taskPages.SwitchToPage(r.taskData.TaskStyle)
			r.focusable = r.updateSelectionFocusable()
			misc.App.SetFocus(r.focusable[0].Primitive)
			misc.RunLastFocus = &r.focusable[0].Primitive
			return nil
		}
		return event
	})

	// Right: Servers with table/tree toggle
	isServerTable := r.serverData.ServerStyle == "server-table"
	serverPages := tview.NewPages().
		AddPage("server-table", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(r.serverData.ServerTableView.Root, 0, 1, true), true, isServerTable).
		AddPage("server-tree", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(r.serverData.ServerTreeView.Root, 0, 1, false), true, !isServerTable)

	serverViewPane := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(serverPages, 0, 1, true)

	// Handle Ctrl+E for server view toggle
	serverViewPane.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if misc.App.GetFocus() == misc.Search {
			return event
		}

		switch event.Key() {
		case tcell.KeyCtrlE:
			if r.serverData.ServerStyle == "server-table" {
				r.serverData.ServerStyle = "server-tree"
			} else {
				r.serverData.ServerStyle = "server-table"
			}
			serverPages.SwitchToPage(r.serverData.ServerStyle)
			r.focusable = r.updateSelectionFocusable()
			// Find the server view in focusable and focus it
			for _, item := range r.focusable {
				if item.Primitive == r.serverData.ServerTableView.Table ||
					item.Primitive == r.serverData.ServerTreeView.Tree {
					misc.App.SetFocus(item.Primitive)
					misc.RunLastFocus = &item.Primitive
					break
				}
			}
			return nil
		}
		return event
	})

	serverPane := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(serverViewPane, 0, 1, true)

	if r.serverData.TagView != nil && len(r.serverData.ServerTags) > 0 {
		serverPane.AddItem(r.serverData.TagView.Root, 20, 0, false)
	}

	// Main layout
	mainContent := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(taskPane, 0, 1, true).
		AddItem(serverPane, 0, 1, false)

	// Page with info at bottom
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(mainContent, 0, 1, true).
		AddItem(info, 1, 0, false)

	return page
}

func (r *TRunPage) updateSelectionFocusable() []*misc.TItem {
	var focusable []*misc.TItem

	// Add task view based on current style
	if r.taskData.TaskStyle == "task-table" {
		focusable = append(focusable, misc.GetTUIItem(
			r.taskData.TaskTableView.Table,
			r.taskData.TaskTableView.Table.Box,
		))
	} else {
		focusable = append(focusable, misc.GetTUIItem(
			r.taskData.TaskTreeView.Tree,
			r.taskData.TaskTreeView.Tree.Box,
		))
	}

	// Add server view based on current style
	if r.serverData.ServerStyle == "server-table" {
		focusable = append(focusable, misc.GetTUIItem(
			r.serverData.ServerTableView.Table,
			r.serverData.ServerTableView.Table.Box,
		))
	} else {
		focusable = append(focusable, misc.GetTUIItem(
			r.serverData.ServerTreeView.Tree,
			r.serverData.ServerTreeView.Tree.Box,
		))
	}

	if r.serverData.TagView != nil && len(r.serverData.ServerTags) > 0 {
		focusable = append(
			focusable,
			misc.GetTUIItem(
				r.serverData.TagView.List,
				r.serverData.TagView.List.Box,
			),
		)
	}

	return focusable
}

func (r *TRunPage) updateOutputFocusable() []*misc.TItem {
	return []*misc.TItem{
		misc.GetTUIItem(r.outputView.Output, r.outputView.Output.Box),
	}
}

func (r *TRunPage) switchView(pages *tview.Pages) []*misc.TItem {
	name, _ := pages.GetFrontPage()
	if name == "run-output" {
		pages.SwitchToPage("run-selection")
		return r.updateSelectionFocusable()
	}
	pages.SwitchToPage("run-output")
	return r.updateOutputFocusable()
}

func (r *TRunPage) switchBeforeRun(pages *tview.Pages) []*misc.TItem {
	name, _ := pages.GetFrontPage()
	if name == "run-selection" {
		pages.SwitchToPage("run-output")
		return r.updateOutputFocusable()
	}
	return r.focusable
}

func (r *TRunPage) runTasks() {
	// Get selected servers
	selectedServers := r.serverData.GetSelectedServerObjects()
	if len(selectedServers) == 0 {
		r.outputView.Write("[yellow]No servers selected[-]\n")
		return
	}

	// Get selected tasks
	selectedTasks := r.taskData.GetSelectedTaskObjects()
	if len(selectedTasks) == 0 {
		r.outputView.Write("[yellow]No tasks selected[-]\n")
		return
	}

	// Clear output if option is set
	if r.spec.ClearBeforeRun {
		r.outputView.Clear()
	}

	// Get writer for output
	writer := r.outputView.GetWriter()

	// Run each task
	go func() {
		for _, task := range selectedTasks {
			r.runSingleTask(&task, selectedServers, writer)
		}
		misc.App.QueueUpdateDraw(func() {})
	}()
}

func (r *TRunPage) runSingleTask(task *dao.Task, servers []dao.Server, writer io.Writer) {
	config := misc.Config
	spec := r.spec

	// Create run flags with spec options
	runFlags := &core.RunFlags{
		Output:   spec.Output,
		Strategy: spec.Strategy,
	}
	setRunFlags := &core.SetRunFlags{}

	// Create Run struct
	runner := &run.Run{
		LocalClients:  make(map[string]run.Client),
		RemoteClients: make(map[string]run.Client),
		Servers:       servers,
		Task:          task,
		Config:        *config,
	}

	// Evaluate config env
	configEnv, err := dao.EvaluateEnv(config.Envs)
	if err != nil {
		fmt.Fprintf(writer, "[red]Error evaluating config env: %s[-]\n", err.Error())
		return
	}

	// Parse task
	err = runner.ParseTask(configEnv, []string{}, runFlags, setRunFlags)
	if err != nil {
		fmt.Fprintf(writer, "[red]Error parsing task: %s[-]\n", err.Error())
		return
	}

	// Apply spec options to task
	task.Spec.Strategy = spec.Strategy
	task.Spec.Output = spec.Output
	task.Spec.IgnoreErrors = spec.IgnoreErrors
	task.Spec.IgnoreUnreachable = spec.IgnoreUnreachable
	task.Spec.OmitEmptyRows = spec.OmitEmptyRows
	task.Spec.OmitEmptyColumns = spec.OmitEmptyColumns
	task.Spec.AnyErrorsFatal = spec.AnyErrorsFatal

	// Parse servers
	errConnects, err := run.ParseServers(config.SSHConfigFile, &runner.Servers, runFlags, task.Spec.Order)
	if err != nil {
		fmt.Fprintf(writer, "[red]Error parsing servers: %s[-]\n", err.Error())
		return
	}

	if len(errConnects) > 0 {
		for _, e := range errConnects {
			fmt.Fprintf(writer, "[red]Parse error for %s: %s[-]\n", e.Name, e.Reason)
		}
		return
	}

	// Set up clients
	numClients := len(servers) * 2
	clientCh := make(chan run.Client, numClients)
	errCh := make(chan run.ErrConnect, numClients)

	errConnect, err := runner.SetClients(task, runFlags, numClients, clientCh, errCh)
	if err != nil {
		fmt.Fprintf(writer, "[red]Error setting up clients: %s[-]\n", err.Error())
		return
	}

	if len(errConnect) > 0 {
		fmt.Fprintf(writer, "[yellow]Unreachable hosts:[-]\n")
		for _, e := range errConnect {
			fmt.Fprintf(writer, "  - %s (%s): %s\n", e.Name, e.Host, e.Reason)
		}
		if !spec.IgnoreUnreachable {
			return
		}
	}

	// Get reachable servers
	var reachableServers []dao.Server
	for _, server := range runner.Servers {
		if server.Local {
			reachableServers = append(reachableServers, server)
			continue
		}

		_, reachable := runner.RemoteClients[server.Name]
		if reachable {
			reachableServers = append(reachableServers, server)
		}
	}
	runner.Servers = reachableServers

	if len(runner.Servers) == 0 {
		fmt.Fprintf(writer, "[yellow]No reachable servers[-]\n")
		runner.CleanupClients()
		return
	}

	// Execute task based on output type
	if spec.Output == "table" {
		// Table output
		data, _, _ := runner.Table(false)

		// Format table output for TUI
		if len(data.Headers) > 0 && len(data.Rows) > 0 {
			r.writeTableOutput(writer, data)
		}
	} else {
		// Text output (default)
		runner.TextTUI(false, writer, writer)
	}

	// Cleanup
	runner.CleanupClients()
}

// writeTableOutput formats table data for TUI display
func (r *TRunPage) writeTableOutput(writer io.Writer, data dao.TableOutput) {
	// Calculate column widths
	colWidths := make([]int, len(data.Headers))
	for i, header := range data.Headers {
		colWidths[i] = len(header)
	}
	for _, row := range data.Rows {
		for i, col := range row.Columns {
			if i < len(colWidths) {
				// Handle multi-line output - use first line for width calc
				lines := splitLines(col)
				for _, line := range lines {
					if len(line) > colWidths[i] {
						colWidths[i] = len(line)
					}
				}
			}
		}
	}

	// Cap column widths at reasonable max
	maxColWidth := 60
	for i := range colWidths {
		if colWidths[i] > maxColWidth {
			colWidths[i] = maxColWidth
		}
	}

	// Print header
	fmt.Fprintf(writer, "[#d787ff::b]")
	for i, header := range data.Headers {
		fmt.Fprintf(writer, "%-*s  ", colWidths[i], header)
	}
	fmt.Fprintf(writer, "[-:-:-]\n")

	// Print separator
	for i := range data.Headers {
		fmt.Fprintf(writer, "%s  ", repeatChar('-', colWidths[i]))
	}
	fmt.Fprintf(writer, "\n")

	// Print rows
	for _, row := range data.Rows {
		// Get max lines in this row
		maxLines := 1
		rowLines := make([][]string, len(row.Columns))
		for i, col := range row.Columns {
			rowLines[i] = splitLines(col)
			if len(rowLines[i]) > maxLines {
				maxLines = len(rowLines[i])
			}
		}

		// Print each line of the row
		for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
			for i := range row.Columns {
				var cellContent string
				if lineIdx < len(rowLines[i]) {
					cellContent = rowLines[i][lineIdx]
				}
				// Truncate if too long
				if len(cellContent) > colWidths[i] {
					cellContent = cellContent[:colWidths[i]-3] + "..."
				}
				fmt.Fprintf(writer, "%-*s  ", colWidths[i], cellContent)
			}
			fmt.Fprintf(writer, "\n")
		}
	}
}

func splitLines(s string) []string {
	if s == "" {
		return []string{""}
	}
	var lines []string
	current := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" || len(lines) == 0 {
		lines = append(lines, current)
	}
	return lines
}

func repeatChar(c rune, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += string(c)
	}
	return result
}

func (r *TRunPage) describeItem() {
	currentFocus := misc.App.GetFocus()

	// Check if task table or tree is focused
	if currentFocus == r.taskData.TaskTableView.Table || currentFocus == r.taskData.TaskTreeView.Tree {
		r.describeTask()
		return
	}

	// Check if server table or tree is focused
	if currentFocus == r.serverData.ServerTableView.Table || currentFocus == r.serverData.ServerTreeView.Tree {
		r.describeServer()
		return
	}
}

func (r *TRunPage) describeTask() {
	var taskID string

	if r.taskData.TaskStyle == "task-table" {
		row, _ := r.taskData.TaskTableView.Table.GetSelection()
		if row < 1 {
			return
		}

		tasks := r.taskData.GetFilteredTasks()
		if row-1 >= len(tasks) {
			return
		}
		taskID = tasks[row-1].ID
	} else {
		node := r.taskData.TaskTreeView.Tree.GetCurrentNode()
		if node == nil {
			return
		}
		ref := node.GetReference()
		if ref == nil {
			return
		}
		taskID = ref.(string)
	}

	// Get full task from config for complete info
	fullTask, err := misc.Config.GetTask(taskID)
	if err != nil {
		return
	}

	// Build sub-tasks info from TaskRefs
	var subTasks []misc.SubTaskInfo
	for _, ref := range fullTask.TaskRefs {
		st := misc.SubTaskInfo{
			Name:  ref.Name,
			Desc:  ref.Desc,
			Cmd:   ref.Cmd,
			Task:  ref.Task,
			IsRef: ref.Task != "",
		}
		subTasks = append(subTasks, st)
	}

	// Also include resolved Tasks (TaskCmd)
	for _, tc := range fullTask.Tasks {
		st := misc.SubTaskInfo{
			Name:  tc.Name,
			Desc:  tc.Desc,
			Cmd:   tc.Cmd,
			IsRef: false,
		}
		subTasks = append(subTasks, st)
	}

	description := misc.FormatTaskBlock(
		fullTask.Name,
		fullTask.Desc,
		fullTask.Cmd,
		fullTask.Local,
		fullTask.TTY,
		fullTask.Attach,
		fullTask.WorkDir,
		fullTask.Shell,
		fullTask.Envs,
		fullTask.Target.Tags,
		subTasks,
		fullTask.Spec.Name,
		fullTask.Target.Name,
		fullTask.Theme.Name,
	)
	components.OpenTextModal("describe-modal", description, fullTask.Name)
}

func (r *TRunPage) describeServer() {
	var serverName string

	if r.serverData.ServerStyle == "server-table" {
		row, _ := r.serverData.ServerTableView.Table.GetSelection()
		if row < 1 {
			return
		}

		servers := r.serverData.GetFilteredServers()
		if row-1 >= len(servers) {
			return
		}
		serverName = servers[row-1].Name
	} else {
		node := r.serverData.ServerTreeView.Tree.GetCurrentNode()
		if node == nil {
			return
		}
		ref := node.GetReference()
		if ref == nil {
			return
		}
		serverName = ref.(string)
		if serverName == "" {
			return
		}
	}

	// Get full server from config for complete info
	fullServer, err := misc.Config.GetServer(serverName)
	if err != nil {
		return
	}

	// Get bastion hosts as strings
	var bastions []string
	for _, bastion := range fullServer.Bastions {
		bastionStr := bastion.Host
		if bastion.User != "" {
			bastionStr = bastion.User + "@" + bastionStr
		}
		bastions = append(bastions, bastionStr)
	}

	// Get identity file
	identityFile := ""
	if fullServer.IdentityFile != nil {
		identityFile = *fullServer.IdentityFile
	}

	description := misc.FormatServerBlock(
		fullServer.Name,
		fullServer.Desc,
		fullServer.Host,
		fullServer.User,
		fullServer.Port,
		fullServer.Local,
		fullServer.Tags,
		bastions,
		identityFile,
		fullServer.WorkDir,
	)
	components.OpenTextModal("describe-modal", description, fullServer.Name)
}
