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

type TExecPage struct {
	focusable   []*misc.TItem
	serverData  *views.TServer
	commandArea *tview.TextArea
	outputView  *components.TOutput
	spec        *views.TSpec
}

func CreateExecPage(
	servers []dao.Server,
	serverTags []string,
) *tview.Flex {
	e := &TExecPage{}

	// Data
	e.serverData = views.CreateServersData(
		servers,
		serverTags,
		[]string{"Server", "Host", "Tags"},
		1,
		true,
		true,
		true,
		len(serverTags) > 0,
	)

	// Command input area
	e.commandArea = tview.NewTextArea()
	e.commandArea.SetBorder(true)
	e.commandArea.SetTitle(" Command ")
	e.commandArea.SetTitleAlign(tview.AlignLeft)
	e.commandArea.SetBorderColor(misc.STYLE_BORDER.Fg)
	e.commandArea.SetPlaceholder("Enter command to execute...")

	// Output view
	e.outputView = &components.TOutput{Title: "Output"}
	e.outputView.Create()

	// Spec options
	e.spec = views.CreateSpecView()

	// Shortcut info view
	execInfoView := views.CreateExecInfoView()

	// Left panel: Command input + Output
	leftPanel := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(e.commandArea, 5, 0, true).
		AddItem(e.outputView.Root, 0, 1, false)

	// Right panel: Servers with table/tree toggle
	isServerTable := e.serverData.ServerStyle == "server-table"
	serverPages := tview.NewPages().
		AddPage("server-table", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(e.serverData.ServerTableView.Root, 0, 1, true), true, isServerTable).
		AddPage("server-tree", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(e.serverData.ServerTreeView.Root, 0, 1, false), true, !isServerTable)

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
			if e.serverData.ServerStyle == "server-table" {
				e.serverData.ServerStyle = "server-tree"
			} else {
				e.serverData.ServerStyle = "server-table"
			}
			serverPages.SwitchToPage(e.serverData.ServerStyle)
			e.focusable = e.updateExecFocusable()
			// Find the server view in focusable and focus it
			for _, item := range e.focusable {
				if item.Primitive == e.serverData.ServerTableView.Table ||
					item.Primitive == e.serverData.ServerTreeView.Tree {
					misc.App.SetFocus(item.Primitive)
					misc.ExecLastFocus = &item.Primitive
					break
				}
			}
			return nil
		}
		return event
	})

	rightPanel := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(serverViewPane, 0, 1, true)

	if e.serverData.TagView != nil && len(serverTags) > 0 {
		rightPanel.AddItem(e.serverData.TagView.Root, 20, 0, false)
	}

	// Main layout
	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(leftPanel, 0, 1, true).
		AddItem(rightPanel, 0, 1, false)

	// Page with info at bottom
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(mainLayout, 0, 1, true).
		AddItem(execInfoView, 1, 0, false).
		AddItem(misc.Search, 1, 0, false)

	// Focus
	e.focusable = e.updateExecFocusable()
	misc.ExecLastFocus = &e.focusable[0].Primitive

	// Focus handlers for command area
	e.commandArea.SetFocusFunc(func() {
		misc.PreviousPane = e.commandArea
		e.commandArea.SetBorderColor(misc.STYLE_BORDER_FOCUS.Fg)
	})
	e.commandArea.SetBlurFunc(func() {
		e.commandArea.SetBorderColor(misc.STYLE_BORDER.Fg)
	})

	// Shortcuts
	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlR:
			e.execCommand()
			return nil
		case tcell.KeyCtrlO:
			components.OpenModal("spec-modal", "Options", e.spec.View, 35, 14)
			return nil
		case tcell.KeyTab:
			// Don't switch focus if in command area and not at end
			if misc.App.GetFocus() == e.commandArea {
				// Allow Tab to cycle focus
			}
			nextPrimitive := misc.FocusNext(e.focusable)
			misc.ExecLastFocus = nextPrimitive
			return nil
		case tcell.KeyBacktab:
			nextPrimitive := misc.FocusPrevious(e.focusable)
			misc.ExecLastFocus = nextPrimitive
			return nil
		case tcell.KeyCtrlX:
			e.outputView.Clear()
			return nil
		case tcell.KeyRune:
			// Allow typing in command area
			if misc.App.GetFocus() == e.commandArea {
				return event
			}
			switch event.Rune() {
			case 'd': // Toggle describe modal
				if components.CloseDescribeModal() {
					return nil
				}
				e.describeServer()
				return nil
			case 'C': // Clear filters
				e.serverData.Emitter.PublishAndWait(misc.Event{Name: "remove_tag_filter", Data: ""})
				e.serverData.Emitter.PublishAndWait(misc.Event{Name: "remove_tag_selections", Data: ""})
				e.serverData.Emitter.PublishAndWait(misc.Event{Name: "remove_server_filter", Data: ""})
				e.serverData.Emitter.PublishAndWait(misc.Event{Name: "remove_server_selections", Data: ""})
				e.serverData.Emitter.Publish(misc.Event{Name: "filter_servers", Data: ""})
				return nil
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				misc.FocusPage(event, e.focusable)
				return nil
			}
		}

		return event
	})

	return page
}

func (e *TExecPage) updateExecFocusable() []*misc.TItem {
	focusable := []*misc.TItem{
		misc.GetTUIItem(e.commandArea, e.commandArea.Box),
		misc.GetTUIItem(e.outputView.Output, e.outputView.Output.Box),
	}

	// Add server view based on current style
	if e.serverData.ServerStyle == "server-table" {
		focusable = append(focusable, misc.GetTUIItem(
			e.serverData.ServerTableView.Table,
			e.serverData.ServerTableView.Table.Box,
		))
	} else {
		focusable = append(focusable, misc.GetTUIItem(
			e.serverData.ServerTreeView.Tree,
			e.serverData.ServerTreeView.Tree.Box,
		))
	}

	if e.serverData.TagView != nil && len(e.serverData.ServerTags) > 0 {
		focusable = append(
			focusable,
			misc.GetTUIItem(
				e.serverData.TagView.List,
				e.serverData.TagView.List.Box,
			),
		)
	}

	return focusable
}

func (e *TExecPage) execCommand() {
	// Get command
	command := e.commandArea.GetText()
	if command == "" {
		e.outputView.Write("[yellow]No command entered[-]\n")
		return
	}

	// Get selected servers
	selectedServers := e.serverData.GetSelectedServerObjects()
	if len(selectedServers) == 0 {
		e.outputView.Write("[yellow]No servers selected[-]\n")
		return
	}

	// Clear output if option is set
	if e.spec.ClearBeforeRun {
		e.outputView.Clear()
	}

	// Get writer for output
	writer := e.outputView.GetWriter()

	// Run command
	go func() {
		e.runCommand(command, selectedServers, writer)
		misc.App.QueueUpdateDraw(func() {})
	}()
}

func (e *TExecPage) describeServer() {
	var serverName string

	if e.serverData.ServerStyle == "server-table" {
		// Get currently focused server from table
		row, _ := e.serverData.ServerTableView.Table.GetSelection()
		if row < 1 {
			return
		}

		servers := e.serverData.GetFilteredServers()
		if row-1 >= len(servers) {
			return
		}
		serverName = servers[row-1].Name
	} else {
		// Get currently focused server from tree
		node := e.serverData.ServerTreeView.Tree.GetCurrentNode()
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

func (e *TExecPage) runCommand(command string, servers []dao.Server, writer io.Writer) {
	config := misc.Config
	spec := e.spec

	// Create a task for the command with spec options
	task := &dao.Task{
		Name: "exec",
		Tasks: []dao.TaskCmd{
			{
				Name: "exec",
				Cmd:  command,
			},
		},
		Spec:  dao.DEFAULT_SPEC,
		Theme: dao.DEFAULT_THEME,
	}

	// Apply spec options
	task.Spec.Strategy = spec.Strategy
	task.Spec.Output = spec.Output
	task.Spec.IgnoreErrors = spec.IgnoreErrors
	task.Spec.IgnoreUnreachable = spec.IgnoreUnreachable
	task.Spec.OmitEmptyRows = spec.OmitEmptyRows
	task.Spec.OmitEmptyColumns = spec.OmitEmptyColumns
	task.Spec.AnyErrorsFatal = spec.AnyErrorsFatal

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
		for _, ec := range errConnect {
			fmt.Fprintf(writer, "  - %s (%s): %s\n", ec.Name, ec.Host, ec.Reason)
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

	// Execute command based on output type
	if spec.Output == "table" {
		// Table output
		data, _, _ := runner.Table(false)

		// Format table output for TUI
		if len(data.Headers) > 0 && len(data.Rows) > 0 {
			e.writeTableOutput(writer, data)
		}
	} else {
		// Text output (default)
		runner.TextTUI(false, writer, writer)
	}

	// Cleanup
	runner.CleanupClients()
}

// writeTableOutput formats table data for TUI display
func (e *TExecPage) writeTableOutput(writer io.Writer, data dao.TableOutput) {
	// Calculate column widths
	colWidths := make([]int, len(data.Headers))
	for i, header := range data.Headers {
		colWidths[i] = len(header)
	}
	for _, row := range data.Rows {
		for i, col := range row.Columns {
			if i < len(colWidths) {
				// Handle multi-line output - use first line for width calc
				lines := splitLinesExec(col)
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
		fmt.Fprintf(writer, "%s  ", repeatCharExec('-', colWidths[i]))
	}
	fmt.Fprintf(writer, "\n")

	// Print rows
	for _, row := range data.Rows {
		// Get max lines in this row
		maxLines := 1
		rowLines := make([][]string, len(row.Columns))
		for i, col := range row.Columns {
			rowLines[i] = splitLinesExec(col)
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

func splitLinesExec(s string) []string {
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

func repeatCharExec(c rune, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += string(c)
	}
	return result
}
