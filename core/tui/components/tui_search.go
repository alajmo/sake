package components

import (
	"strings"

	"github.com/rivo/tview"

	"github.com/alajmo/sake/core/tui/misc"
)

func CreateSearch() *tview.InputField {
	search := tview.NewInputField().
		SetLabel("").
		SetLabelStyle(misc.STYLE_SEARCH_LABEL.Style).
		SetFieldStyle(misc.STYLE_SEARCH_TEXT.Style)
	return search
}

func ShowSearch() {
	misc.Search.SetLabel(misc.Colorize("Search:", misc.STYLE_SEARCH_LABEL.FgStr, misc.STYLE_SEARCH_LABEL.BgStr, "-"))
	misc.Search.SetText("")
	misc.App.SetFocus(misc.Search)
}

func EmptySearch() {
	misc.Search.SetLabel("")
	misc.Search.SetText("")
}

func SearchInTable(table *tview.Table, query string, lastFoundRow, lastFoundCol *int, direction int) {
	query = strings.ToLower(query)
	rowCount := table.GetRowCount()
	colCount := table.GetColumnCount()
	startRow := *lastFoundRow

	if startRow == -1 {
		startRow = 0
	} else {
		startRow += direction
	}

	searchRow := startRow
	for i := 0; i < rowCount; i++ {
		if searchRow < 0 {
			searchRow = rowCount - 1
		} else if searchRow >= rowCount {
			searchRow = 0
		}

		for col := 0; col < colCount; col++ {
			if cell := table.GetCell(searchRow, col); cell != nil {
				if strings.Contains(strings.ToLower(strings.TrimSpace(cell.Text)), query) {
					table.Select(searchRow, col)
					*lastFoundRow, *lastFoundCol = searchRow, col
					return
				}
			}
		}

		searchRow += direction
	}

	*lastFoundRow, *lastFoundCol = -1, -1
}

func SearchInList(list *tview.List, query string, lastFoundIndex *int, direction int) {
	query = strings.ToLower(query)
	itemCount := list.GetItemCount()
	startIndex := *lastFoundIndex

	if startIndex == -1 {
		startIndex = 0
	} else {
		startIndex += direction
	}

	searchIndex := startIndex
	for i := 0; i < itemCount; i++ {
		if searchIndex < 0 {
			searchIndex = itemCount - 1
		} else if searchIndex >= itemCount {
			searchIndex = 0
		}

		mainText, secondaryText := list.GetItemText(searchIndex)
		if strings.Contains(strings.ToLower(mainText), query) ||
			strings.Contains(strings.ToLower(secondaryText), query) {
			list.SetCurrentItem(searchIndex)
			*lastFoundIndex = searchIndex
			return
		}

		searchIndex += direction
	}

	*lastFoundIndex = -1
}

func SearchInTree(tree *tview.TreeView, query string, lastFoundIndex *int, direction int) {
	query = strings.ToLower(query)

	// Get all selectable nodes
	var nodes []*tview.TreeNode
	var walk func(*tview.TreeNode)
	walk = func(node *tview.TreeNode) {
		if node == nil {
			return
		}
		// Only include selectable nodes (ones with references)
		ref := node.GetReference()
		if ref != nil && ref.(string) != "" {
			nodes = append(nodes, node)
		}
		for _, child := range node.GetChildren() {
			walk(child)
		}
	}
	walk(tree.GetRoot())

	if len(nodes) == 0 {
		return
	}

	startIndex := *lastFoundIndex
	if startIndex == -1 {
		startIndex = 0
	} else {
		startIndex += direction
	}

	searchIndex := startIndex
	for i := 0; i < len(nodes); i++ {
		if searchIndex < 0 {
			searchIndex = len(nodes) - 1
		} else if searchIndex >= len(nodes) {
			searchIndex = 0
		}

		node := nodes[searchIndex]
		text := node.GetText()
		// Strip color codes for matching
		text = stripColorCodes(text)
		if strings.Contains(strings.ToLower(text), query) {
			tree.SetCurrentNode(node)
			*lastFoundIndex = searchIndex
			return
		}

		searchIndex += direction
	}

	*lastFoundIndex = -1
}

// stripColorCodes removes tview color tags from text
func stripColorCodes(text string) string {
	result := ""
	inTag := false
	for _, r := range text {
		if r == '[' {
			inTag = true
			continue
		}
		if r == ']' && inTag {
			inTag = false
			continue
		}
		if !inTag {
			result += string(r)
		}
	}
	return strings.TrimSpace(result)
}
