package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/sake/core/tui/misc"
)

type TTree struct {
	Tree     *tview.TreeView
	Root     *tview.Flex
	RootNode *tview.TreeNode
	Filter   *tview.InputField

	List          []*TNode
	Title         string
	RootTitle     string
	FilterValue   *string
	SelectEnabled bool

	IsNodeSelected   func(name string) bool
	ToggleSelectNode func(name string)
	SelectAll        func()
	UnselectAll      func()
	FilterNodes      func()
	DescribeNode     func(name string)
	EditNode         func(name string)
}

type TNode struct {
	ID          string // The reference
	DisplayName string // What is shown
	Type        string

	TreeNode *tview.TreeNode
	Children *[]TNode
}

func (t *TTree) Create() {
	title := misc.Colorize(t.RootTitle, "", "", "-")
	rootNode := tview.NewTreeNode(title)
	rootNode.SetColor(misc.STYLE_DEFAULT.Fg)
	rootNode.SetSelectable(false)

	t.IsNodeSelected = func(name string) bool { return false }
	t.ToggleSelectNode = func(name string) {}
	t.SelectAll = func() {}
	t.UnselectAll = func() {}
	t.FilterNodes = func() {}
	t.DescribeNode = func(name string) {}
	t.EditNode = func(name string) {}

	tree := tview.NewTreeView().
		SetRoot(rootNode).
		SetCurrentNode(rootNode)
	tree.SetGraphics(true)

	filter := CreateFilter()

	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tree, 0, 1, true).
		AddItem(filter, 1, 0, false)
	root.SetTitleAlign(misc.STYLE_TITLE.Align).
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 1)

	t.Root = root
	t.Filter = filter
	t.RootNode = rootNode
	t.Tree = tree

	if t.Title != "" {
		misc.SetActive(t.Root.Box, t.Title, false)
	}

	// Filter
	t.Filter.SetChangedFunc(func(_ string) {
		t.applyFilter()
		t.FilterNodes()
	})

	t.Filter.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentFocus := misc.App.GetFocus()
		if currentFocus == filter {
			switch event.Key() {
			case tcell.KeyEscape:
				t.ClearFilter()
				t.FilterNodes()
				misc.App.SetFocus(tree)
				return nil
			case tcell.KeyEnter:
				t.applyFilter()
				t.FilterNodes()
				misc.App.SetFocus(tree)
			}
			return event
		}
		return event
	})

	// Input
	t.Tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if t.SelectEnabled {
				node := t.Tree.GetCurrentNode()
				ref := node.GetReference()
				if ref != nil {
					name := ref.(string)
					t.ToggleSelectNode(name)
				}
			}
			return nil
		case tcell.KeyCtrlD:
			current := t.Tree.GetCurrentNode()
			_, _, _, height := t.Tree.GetInnerRect()
			visibleNodes := t.getVisibleNodes()
			currentIndex := t.findNodeIndex(visibleNodes, current)
			newIndex := min(currentIndex+height/2, len(visibleNodes)-1)
			if newIndex > 0 && newIndex < len(visibleNodes) {
				t.Tree.SetCurrentNode(visibleNodes[newIndex])
			}
			return nil
		case tcell.KeyCtrlU:
			current := t.Tree.GetCurrentNode()
			_, _, _, height := t.Tree.GetInnerRect()
			visibleNodes := t.getVisibleNodes()
			currentIndex := t.findNodeIndex(visibleNodes, current)
			newIndex := max(currentIndex-height/2, 0)
			if newIndex >= 0 && newIndex < len(visibleNodes) {
				t.Tree.SetCurrentNode(visibleNodes[newIndex])
			}
			return nil
		case tcell.KeyCtrlF:
			current := t.Tree.GetCurrentNode()
			_, _, _, height := t.Tree.GetInnerRect()
			visibleNodes := t.getVisibleNodes()
			currentIndex := t.findNodeIndex(visibleNodes, current)
			newIndex := min(currentIndex+height, len(visibleNodes)-1)
			if newIndex > 0 && newIndex < len(visibleNodes) {
				t.Tree.SetCurrentNode(visibleNodes[newIndex])
			}
			return nil
		case tcell.KeyCtrlB:
			current := t.Tree.GetCurrentNode()
			_, _, _, height := t.Tree.GetInnerRect()
			visibleNodes := t.getVisibleNodes()
			currentIndex := t.findNodeIndex(visibleNodes, current)
			newIndex := max(currentIndex-height, 0)
			if newIndex >= 0 && newIndex < len(visibleNodes) {
				t.Tree.SetCurrentNode(visibleNodes[newIndex])
			}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case ' ': // Toggle item (space)
				if t.SelectEnabled {
					node := t.Tree.GetCurrentNode()
					ref := node.GetReference()
					if ref != nil {
						name := ref.(string)
						t.ToggleSelectNode(name)
					}
				}
				return nil
			case 'a': // Select all
				if t.SelectEnabled {
					t.SelectAll()
				}
				return nil
			case 'c': // Unselect all
				if t.SelectEnabled {
					t.UnselectAll()
				}
				return nil
			case 'f': // Filter rows
				ShowFilter(filter, *t.FilterValue)
				return nil
			case 'F': // Remove filter
				CloseFilter(filter)
				*t.FilterValue = ""
				return nil
			case 'o': // Edit in editor
				item := tree.GetCurrentNode()
				ref := item.GetReference()
				if ref != nil {
					name := ref.(string)
					t.EditNode(name)
				}
				return nil
			case 'd': // Toggle description modal
				if CloseDescribeModal() {
					return nil
				}
				item := tree.GetCurrentNode()
				ref := item.GetReference()
				if ref != nil {
					name := ref.(string)
					t.DescribeNode(name)
				}
				return nil
			case 'g': // Top
				tree.SetCurrentNode(rootNode)
				misc.App.QueueEvent(tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone))
				return nil
			case 'G': // Bottom
				children := rootNode.GetChildren()
				if len(children) > 0 {
					last := children[len(children)-1]
					ref := last.GetReference()

					if ref == nil || ref.(string) == "" {
						children = last.GetChildren()
						if len(children) > 0 {
							last = children[len(children)-1]
						}
					}

					tree.SetCurrentNode(last)
					misc.App.QueueEvent(tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone))
				}
				return nil
			}
		}
		return event
	})

	// Events
	var previousNode *tview.TreeNode
	var previousColor tcell.Color
	tree.SetChangedFunc(func(node *tview.TreeNode) {
		if previousNode != nil {
			previousNode.SetColor(previousColor)
		}
		if node != nil {
			previousColor = node.GetColor()
			previousNode = node
			node.SetColor(misc.STYLE_ITEM_FOCUSED.Bg)
		}
	})

	t.Tree.SetFocusFunc(func() {
		InitFilter(t.Filter, *t.FilterValue)
		misc.PreviousPane = t.Tree
		misc.SetActive(t.Root.Box, t.Title, true)
	})

	t.Tree.SetBlurFunc(func() {
		misc.PreviousPane = t.Tree
		misc.SetActive(t.Root.Box, t.Title, false)
	})
}

func (t *TTree) UpdateTasks(nodes []TNode) {
	t.RootNode.ClearChildren()
	t.List = []*TNode{}

	for _, parentNode := range nodes {
		// Parent
		displayName := misc.Colorize(parentNode.DisplayName, misc.STYLE_ITEM.FgStr, misc.STYLE_ITEM.BgStr, "-")
		parentTreeNode := tview.NewTreeNode(displayName).
			SetReference(parentNode.ID).
			SetSelectable(true)
		t.RootNode.AddChild(parentTreeNode)

		parentListNode := &TNode{
			DisplayName: parentNode.DisplayName,
			ID:          parentNode.ID,
			Type:        parentNode.Type,
			TreeNode:    parentTreeNode,
			Children:    &[]TNode{},
		}
		t.List = append(t.List, parentListNode)

		// Children
		if parentNode.Children != nil {
			for _, childNode := range *parentNode.Children {
				var childDisplayName string
				if childNode.Type == "task-ref" {
					// Use magenta (TABLE_HEADER color) for task refs
					childDisplayName = misc.Colorize(childNode.DisplayName, misc.STYLE_TABLE_HEADER.FgStr, "", "-")
				} else {
					childDisplayName = misc.Colorize(childNode.DisplayName, misc.STYLE_ITEM.FgStr, misc.STYLE_ITEM.BgStr, "-")
				}
				childTreeNode := tview.NewTreeNode(childDisplayName).
					SetSelectable(false)
				parentTreeNode.AddChild(childTreeNode)

				listChildNode := &TNode{
					DisplayName: childNode.DisplayName,
					Type:        childNode.Type,
					TreeNode:    childTreeNode,
					Children:    &[]TNode{},
				}
				*parentListNode.Children = append(*parentListNode.Children, *listChildNode)
			}
		}
	}
}

func (t *TTree) UpdateTasksStyle() {
	for _, node := range t.List {
		if t.IsNodeSelected(node.ID) {
			displayName := misc.Colorize(node.DisplayName, misc.STYLE_ITEM_SELECTED.FgStr, "", "-")
			node.TreeNode.SetText(displayName)
			for _, child := range *node.Children {
				displayName := misc.Colorize(child.DisplayName, misc.STYLE_ITEM_SELECTED.FgStr, "", "-")
				child.TreeNode.SetText(displayName)
			}
		} else {
			displayName := misc.Colorize(node.DisplayName, misc.STYLE_ITEM.FgStr, misc.STYLE_ITEM.BgStr, "-")
			node.TreeNode.SetText(displayName)
			for _, child := range *node.Children {
				if child.Type == "task-ref" {
					// Use magenta (TABLE_HEADER color) for task refs
					displayName := misc.Colorize(child.DisplayName, misc.STYLE_TABLE_HEADER.FgStr, "", "-")
					child.TreeNode.SetText(displayName)
				} else {
					displayName := misc.Colorize(child.DisplayName, misc.STYLE_ITEM.FgStr, misc.STYLE_ITEM.BgStr, "-")
					child.TreeNode.SetText(displayName)
				}
			}
		}
	}
}

func (t *TTree) UpdateServers(nodes []TNode) {
	t.RootNode.ClearChildren()
	t.List = []*TNode{}

	for _, parentNode := range nodes {
		if parentNode.Type == "group" {
			// Group node (IP prefix or hostname domain)
			displayName := misc.Colorize(parentNode.DisplayName, misc.STYLE_TABLE_HEADER.FgStr, "", "-")
			parentTreeNode := tview.NewTreeNode(displayName).
				SetReference("").
				SetSelectable(false)
			t.RootNode.AddChild(parentTreeNode)

			parentListNode := &TNode{
				DisplayName: parentNode.DisplayName,
				ID:          "",
				Type:        "group",
				TreeNode:    parentTreeNode,
				Children:    &[]TNode{},
			}
			t.List = append(t.List, parentListNode)

			// Children (actual servers)
			if parentNode.Children != nil {
				for _, childNode := range *parentNode.Children {
					childDisplayName := misc.Colorize(childNode.DisplayName, misc.STYLE_ITEM.FgStr, misc.STYLE_ITEM.BgStr, "-")
					childTreeNode := tview.NewTreeNode(childDisplayName).
						SetReference(childNode.ID).
						SetSelectable(true)
					parentTreeNode.AddChild(childTreeNode)

					listChildNode := &TNode{
						DisplayName: childNode.DisplayName,
						ID:          childNode.ID,
						Type:        "server",
						TreeNode:    childTreeNode,
						Children:    &[]TNode{},
					}
					*parentListNode.Children = append(*parentListNode.Children, *listChildNode)
				}
			}
		} else {
			// Flat server node
			displayName := misc.Colorize(parentNode.DisplayName, misc.STYLE_ITEM.FgStr, misc.STYLE_ITEM.BgStr, "-")
			parentTreeNode := tview.NewTreeNode(displayName).
				SetReference(parentNode.ID).
				SetSelectable(true)
			t.RootNode.AddChild(parentTreeNode)

			parentListNode := &TNode{
				DisplayName: parentNode.DisplayName,
				ID:          parentNode.ID,
				Type:        "server",
				TreeNode:    parentTreeNode,
				Children:    &[]TNode{},
			}
			t.List = append(t.List, parentListNode)

			// Host info as child
			if parentNode.Children != nil {
				for _, childNode := range *parentNode.Children {
					childDisplayName := misc.Colorize(childNode.DisplayName, misc.STYLE_TABLE_HEADER.FgStr, "", "-")
					childTreeNode := tview.NewTreeNode(childDisplayName).
						SetSelectable(false)
					parentTreeNode.AddChild(childTreeNode)

					listChildNode := &TNode{
						DisplayName: childNode.DisplayName,
						ID:          childNode.ID,
						Type:        "host",
						TreeNode:    childTreeNode,
						Children:    &[]TNode{},
					}
					*parentListNode.Children = append(*parentListNode.Children, *listChildNode)
				}
			}
		}
	}
}

func (t *TTree) UpdateServersStyle() {
	for _, node := range t.List {
		if node.Type == "group" {
			// Group headers stay magenta
			displayName := misc.Colorize(node.DisplayName, misc.STYLE_TABLE_HEADER.FgStr, "", "-")
			node.TreeNode.SetText(displayName)
			for _, child := range *node.Children {
				if t.IsNodeSelected(child.ID) {
					displayName := misc.Colorize(child.DisplayName, misc.STYLE_ITEM_SELECTED.FgStr, "", "-")
					child.TreeNode.SetText(displayName)
				} else {
					displayName := misc.Colorize(child.DisplayName, misc.STYLE_ITEM.FgStr, misc.STYLE_ITEM.BgStr, "-")
					child.TreeNode.SetText(displayName)
				}
			}
		} else {
			// Flat server
			if t.IsNodeSelected(node.ID) {
				displayName := misc.Colorize(node.DisplayName, misc.STYLE_ITEM_SELECTED.FgStr, "", "-")
				node.TreeNode.SetText(displayName)
				for _, child := range *node.Children {
					displayName := misc.Colorize(child.DisplayName, misc.STYLE_ITEM_SELECTED.FgStr, "", "-")
					child.TreeNode.SetText(displayName)
				}
			} else {
				displayName := misc.Colorize(node.DisplayName, misc.STYLE_ITEM.FgStr, misc.STYLE_ITEM.BgStr, "-")
				node.TreeNode.SetText(displayName)
				for _, child := range *node.Children {
					// Host info in magenta
					displayName := misc.Colorize(child.DisplayName, misc.STYLE_TABLE_HEADER.FgStr, "", "-")
					child.TreeNode.SetText(displayName)
				}
			}
		}
	}
}

func (t *TTree) ToggleSelectCurrentNode(id string) {
	for i := 0; i < len(t.List); i++ {
		node := t.List[i]
		if node.ID == id {
			t.setNodeSelect(node)
			return
		}
		// Also check children for grouped servers
		for _, child := range *node.Children {
			if child.ID == id {
				t.setServerNodeSelect(&child)
				return
			}
		}
	}
}

func (t *TTree) setServerNodeSelect(node *TNode) {
	if t.IsNodeSelected(node.ID) {
		displayName := misc.Colorize(node.DisplayName, misc.STYLE_ITEM_SELECTED.FgStr, "", "-")
		node.TreeNode.SetText(displayName)
	} else {
		displayName := misc.Colorize(node.DisplayName, misc.STYLE_ITEM.FgStr, misc.STYLE_ITEM.BgStr, "-")
		node.TreeNode.SetText(displayName)
	}
}

func (t *TTree) setNodeSelect(node *TNode) {
	if t.IsNodeSelected(node.ID) {
		displayName := misc.Colorize(node.DisplayName, misc.STYLE_ITEM_SELECTED.FgStr, "", "-")
		node.TreeNode.SetText(displayName)
		for _, childNode := range *node.Children {
			displayName := misc.Colorize(childNode.DisplayName, misc.STYLE_ITEM_SELECTED.FgStr, "", "-")
			childNode.TreeNode.SetText(displayName)
		}
		return
	}

	displayName := misc.Colorize(node.DisplayName, misc.STYLE_ITEM.FgStr, misc.STYLE_ITEM.BgStr, "-")
	node.TreeNode.SetText(displayName)
	for _, childNode := range *node.Children {
		if childNode.Type == "task-ref" {
			// Use magenta (TABLE_HEADER color) for task refs
			displayName := misc.Colorize(childNode.DisplayName, misc.STYLE_TABLE_HEADER.FgStr, "", "-")
			childNode.TreeNode.SetText(displayName)
		} else {
			displayName := misc.Colorize(childNode.DisplayName, misc.STYLE_ITEM.FgStr, misc.STYLE_ITEM.BgStr, "-")
			childNode.TreeNode.SetText(displayName)
		}
	}
}

func (t *TTree) FocusFirst() {
	children := t.RootNode.GetChildren()
	if len(children) > 0 {
		t.Tree.SetCurrentNode(children[0])
	}
}

func (t *TTree) FocusLast() {
	children := t.RootNode.GetChildren()
	if len(children) == 0 {
		return
	}
	last := children[len(children)-1]
	ref := last.GetReference()

	if ref == nil || ref.(string) == "" {
		children = last.GetChildren()
		if len(children) > 0 {
			last = children[len(children)-1]
		}
	}

	t.Tree.SetCurrentNode(last)
}

func (t *TTree) ClearFilter() {
	CloseFilter(t.Filter)
	*t.FilterValue = ""
}

func (t *TTree) applyFilter() {
	*t.FilterValue = t.Filter.GetText()
}

func (t *TTree) getVisibleNodes() []*tview.TreeNode {
	var nodes []*tview.TreeNode
	var walk func(*tview.TreeNode)
	walk = func(node *tview.TreeNode) {
		if node == nil {
			return
		}
		ref := node.GetReference()
		if ref != nil && ref.(string) != "" {
			nodes = append(nodes, node)
		}
		if node.IsExpanded() {
			for _, child := range node.GetChildren() {
				walk(child)
			}
		}
	}
	walk(t.RootNode)
	return nodes
}

func (t *TTree) findNodeIndex(nodes []*tview.TreeNode, target *tview.TreeNode) int {
	for i, node := range nodes {
		if node == target {
			return i
		}
	}
	return 0
}
