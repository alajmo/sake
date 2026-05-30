package views

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/tui/components"
	"github.com/alajmo/sake/core/tui/misc"
)

type TTask struct {
	// UI
	Page          *tview.Flex
	TaskTableView *components.TTable
	TaskTreeView  *components.TTree

	// Task
	Tasks           []dao.Task
	TasksFiltered   []dao.Task
	TasksSelected   map[string]bool
	taskFilterValue *string
	Headers         []string
	ShowHeaders     bool
	TaskStyle       string

	// Misc
	Emitter *misc.EventEmitter
}

func CreateTasksData(
	tasks []dao.Task,
	headers []string,
	prefixNumber int,
	showTitle bool,
	showHeaders bool,
	selectEnabled bool,
) *TTask {
	t := &TTask{
		Tasks:           tasks,
		TasksFiltered:   tasks,
		TasksSelected:   make(map[string]bool),
		taskFilterValue: new(string),

		ShowHeaders: showHeaders,
		Headers:     headers,
		TaskStyle:   "task-table",

		Emitter: misc.NewEventEmitter(),
	}

	for _, task := range t.Tasks {
		t.TasksSelected[task.ID] = false
	}

	title := ""
	if showTitle && prefixNumber > 0 {
		title = fmt.Sprintf("[%d] Tasks (%d)", prefixNumber, len(tasks))
	} else if showTitle {
		title = fmt.Sprintf("Tasks (%d)", len(tasks))
	}

	rows := t.getTableRows()
	taskTable := t.CreateTasksTable(selectEnabled, title, headers, rows)
	t.TaskTableView = taskTable

	nodes := t.getTreeHierarchy()
	taskTree := t.CreateTasksTree(selectEnabled, title, nodes)
	t.TaskTreeView = taskTree

	// Events
	t.Emitter.Subscribe("remove_task_filter", func(e misc.Event) {
		t.TaskTableView.ClearFilter()
		t.TaskTreeView.ClearFilter()
	})
	t.Emitter.Subscribe("remove_task_selections", func(event misc.Event) {
		t.unselectAllTasks()
	})
	t.Emitter.Subscribe("filter_tasks", func(e misc.Event) {
		t.filterTasks()
	})

	return t
}

func (t *TTask) CreateTasksTable(
	selectEnabled bool,
	title string,
	headers []string,
	rows [][]string,
) *components.TTable {
	table := &components.TTable{
		Title:         title,
		ToggleEnabled: selectEnabled,
		ShowHeaders:   t.ShowHeaders,
		FilterValue:   t.taskFilterValue,
	}
	table.Create()
	table.Update(headers, rows)

	// Methods
	table.IsRowSelected = func(name string) bool {
		return t.TasksSelected[name]
	}
	table.ToggleSelectRow = func(name string) {
		t.toggleSelectTask(name)
	}
	table.SelectAll = func() {
		t.selectAllTasks()
	}
	table.UnselectAll = func() {
		t.unselectAllTasks()
	}
	table.FilterRows = func() {
		t.filterTasks()
	}
	table.DescribeRow = func(taskName string) {
		if taskName != "" {
			t.showTaskDescModal(taskName)
		}
	}
	table.EditRow = func(taskName string) {
		if taskName != "" {
			t.editTask(taskName)
		}
	}
	return table
}

func (t *TTask) CreateTasksTree(
	selectEnabled bool,
	title string,
	nodes []components.TNode,
) *components.TTree {
	tree := &components.TTree{
		Title:         title,
		RootTitle:     "",
		SelectEnabled: selectEnabled,
		FilterValue:   t.taskFilterValue,
	}
	tree.Create()
	tree.UpdateTasks(nodes)
	tree.UpdateTasksStyle()

	tree.IsNodeSelected = func(name string) bool {
		return t.TasksSelected[name]
	}
	tree.ToggleSelectNode = func(name string) {
		t.toggleSelectTask(name)
	}
	tree.SelectAll = func() {
		t.selectAllTasks()
	}
	tree.UnselectAll = func() {
		t.unselectAllTasks()
	}
	tree.FilterNodes = func() {
		t.filterTasks()
	}
	tree.DescribeNode = func(taskName string) {
		if taskName != "" {
			t.showTaskDescModal(taskName)
		}
	}
	tree.EditNode = func(taskName string) {
		if taskName != "" {
			t.editTask(taskName)
		}
	}

	return tree
}

func (t *TTask) getTableRows() [][]string {
	var rows = make([][]string, len(t.TasksFiltered))
	for i, task := range t.TasksFiltered {
		rows[i] = make([]string, len(t.Headers))
		for j, header := range t.Headers {
			rows[i][j] = task.GetValue(header, 0)
		}
	}
	return rows
}

func (t *TTask) getTreeHierarchy() []components.TNode {
	var nodes = []components.TNode{}
	for _, task := range t.TasksFiltered {
		parentNode := &components.TNode{
			DisplayName: task.ID,
			ID:          task.ID,
			Type:        "task",
			Children:    &[]components.TNode{},
		}

		// Sub-tasks
		for _, subTask := range task.Tasks {
			var node *components.TNode
			if subTask.Name != "" {
				node = &components.TNode{
					DisplayName: subTask.Name,
					ID:          task.ID,
					Type:        "command",
					Children:    &[]components.TNode{},
				}
			} else {
				node = &components.TNode{
					DisplayName: "cmd",
					ID:          task.ID,
					Type:        "command",
					Children:    &[]components.TNode{},
				}
			}
			*parentNode.Children = append(*parentNode.Children, *node)
		}

		// Task refs
		for _, ref := range task.TaskRefs {
			node := &components.TNode{
				DisplayName: ref.Task,
				ID:          task.ID,
				Type:        "task-ref",
				Children:    &[]components.TNode{},
			}
			*parentNode.Children = append(*parentNode.Children, *node)
		}

		nodes = append(nodes, *parentNode)
	}

	return nodes
}

func (t *TTask) toggleSelectTask(name string) {
	t.TasksSelected[name] = !t.TasksSelected[name]
	t.TaskTableView.ToggleSelectCurrentRow(name)
	t.TaskTreeView.ToggleSelectCurrentNode(name)
}

func (t *TTask) filterTasks() {
	var finalTasks []dao.Task
	for _, task := range t.Tasks {
		if strings.Contains(strings.ToLower(task.ID), strings.ToLower(*t.taskFilterValue)) ||
			strings.Contains(strings.ToLower(task.Name), strings.ToLower(*t.taskFilterValue)) {
			finalTasks = append(finalTasks, task)
		}
	}
	t.TasksFiltered = finalTasks

	// Table
	rows := t.getTableRows()
	t.TaskTableView.Update(t.Headers, rows)
	t.TaskTableView.Table.ScrollToBeginning()
	t.TaskTableView.Table.Select(1, 0)

	// Tree
	taskTree := t.getTreeHierarchy()
	t.TaskTreeView.UpdateTasks(taskTree)
	t.TaskTreeView.UpdateTasksStyle()
	t.TaskTreeView.FocusFirst()
}

func (t *TTask) selectAllTasks() {
	for _, task := range t.TasksFiltered {
		t.TasksSelected[task.ID] = true
	}
	t.TaskTableView.UpdateRowStyle()
	t.TaskTreeView.UpdateTasksStyle()
}

func (t *TTask) unselectAllTasks() {
	for _, task := range t.TasksFiltered {
		t.TasksSelected[task.ID] = false
	}
	t.TaskTableView.UpdateRowStyle()
	t.TaskTreeView.UpdateTasksStyle()
}

func (t *TTask) showTaskDescModal(name string) {
	task, err := misc.Config.GetTask(name)
	if err != nil {
		return
	}

	// Build sub-tasks info from TaskRefs
	var subTasks []misc.SubTaskInfo
	for _, ref := range task.TaskRefs {
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
	for _, tc := range task.Tasks {
		st := misc.SubTaskInfo{
			Name:  tc.Name,
			Desc:  tc.Desc,
			Cmd:   tc.Cmd,
			IsRef: false,
		}
		subTasks = append(subTasks, st)
	}

	description := misc.FormatTaskBlock(
		task.Name,
		task.Desc,
		task.Cmd,
		task.Local,
		task.TTY,
		task.Attach,
		task.WorkDir,
		task.Shell,
		task.Envs,
		task.Target.Tags,
		subTasks,
		task.Spec.Name,
		task.Target.Name,
		task.Theme.Name,
	)
	components.OpenTextModal("task-description-modal", description, task.Name)
}

func (t *TTask) editTask(taskName string) {
	misc.App.Suspend(func() {
		err := misc.Config.EditTask(taskName)
		if err != nil {
			return
		}
	})
}

// GetSelectedTasks returns IDs of selected tasks
func (t *TTask) GetSelectedTasks() []string {
	var selected []string
	for id, isSelected := range t.TasksSelected {
		if isSelected {
			selected = append(selected, id)
		}
	}
	return selected
}

// GetSelectedTaskObjects returns the selected task objects
func (t *TTask) GetFilteredTasks() []dao.Task {
	return t.TasksFiltered
}

func (t *TTask) GetSelectedTaskObjects() []dao.Task {
	var selected []dao.Task
	for _, task := range t.Tasks {
		if t.TasksSelected[task.ID] {
			selected = append(selected, task)
		}
	}
	return selected
}
