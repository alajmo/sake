package views

import (
	"fmt"

	"github.com/alajmo/sake/core/tui/components"
	"github.com/alajmo/sake/core/tui/misc"
	"github.com/rivo/tview"
)

var Version = "v0.15.1"

func ShowHelpModal() {
	t, table := createShortcutsTable()
	components.OpenModal("help-modal", "Help", t, 65, 37)
	misc.App.SetFocus(table)
}

func shortcutRow(shortcut string, description string) (*tview.TableCell, *tview.TableCell) {
	shortcut = fmt.Sprintf("[%s:%s:%s]%s[-:-:-]",
		misc.STYLE_SHORTCUT_LABEL.FgStr, misc.STYLE_SHORTCUT_LABEL.BgStr, misc.STYLE_SHORTCUT_LABEL.AttrStr, shortcut,
	)

	description = fmt.Sprintf("[%s:%s:%s]%s[-:-:-]",
		misc.STYLE_SHORTCUT_TEXT.FgStr, misc.STYLE_SHORTCUT_TEXT.BgStr, misc.STYLE_SHORTCUT_TEXT.AttrStr, description,
	)

	r1 := tview.NewTableCell(shortcut + "  ").
		SetTextColor(misc.STYLE_SHORTCUT_TEXT.Fg).
		SetAlign(tview.AlignRight).
		SetSelectable(false)

	r2 := tview.NewTableCell(description).
		SetAlign(tview.AlignLeft).
		SetSelectable(false)

	return r1, r2
}

func titleRow(title string) (*tview.TableCell, *tview.TableCell) {
	r1 := tview.NewTableCell("").
		SetTextColor(misc.STYLE_SHORTCUT_TEXT.Fg).
		SetAlign(tview.AlignRight).
		SetSelectable(false)

	r2 := tview.NewTableCell(title).
		SetTextColor(misc.STYLE_TABLE_HEADER.Fg).
		SetAttributes(misc.STYLE_TABLE_HEADER.Attr).
		SetAlign(tview.AlignLeft).
		SetSelectable(false)

	return r1, r2
}

func createShortcutsTable() (*tview.Flex, *tview.Table) {
	table := tview.NewTable()
	table.SetEvaluateAllRows(true)
	table.SetBackgroundColor(misc.STYLE_DEFAULT.Bg)

	sections := []struct {
		title     string
		shortcuts [][2]string
	}{
		{
			title: "--- Global ---",
			shortcuts: [][2]string{
				{"?", "Show this help"},
				{"q, Ctrl + c", "Quit program"},
				{"F5", "Reload app"},
				{"F6", "Re-sync screen buffer"},
			},
		},
		{
			title: "--- Navigation ---",
			shortcuts: [][2]string{
				{"r, F1", "Switch to run page"},
				{"e, F2", "Switch to exec page"},
				{"s, F3", "Switch to servers page"},
				{"t, F4", "Switch to tasks page"},
				{"1-9", "Focus specific pane"},
				{"Tab", "Focus next pane"},
				{"Shift + Tab", "Focus previous pane"},
				{"g", "Go to first item in the current pane"},
				{"G", "Go to last item in the current pane"},
				{"Ctrl + e", "Toggle Table/Tree view"},
				{"Ctrl + o", "Show task options"},
				{"Ctrl + s", "Toggle between selection and output view"},
			},
		},
		{
			title: "--- Actions ---",
			shortcuts: [][2]string{
				{"Escape", "Close modal/filter/search"},
				{"/", "Free text search"},
				{"n/N", "Next/previous search result"},
				{"f", "Filter items for the current pane"},
				{"F", "Clear filter for the current selected pane"},
				{"C", "Clear all filters and selections"},
				{"a", "Select all items in the current pane"},
				{"c", "Clear all selections in the current pane"},
				{"d", "Describe the selected item"},
				{"o", "Open the current selected item in $EDITOR"},
				{"Shift + S", "SSH to server (servers view)"},
				{"Space, Enter", "Toggle selection"},
				{"Ctrl + r", "Run tasks"},
				{"Ctrl + x", "Clear output"},
			},
		},
	}

	// Populate table with sections
	currentRow := 0
	for i, section := range sections {
		// Add spacing between sections except for the first one
		if i > 0 {
			r1, r2 := titleRow("")
			table.SetCell(currentRow, 0, r1)
			table.SetCell(currentRow, 1, r2)
			currentRow++
		}

		// Add section title
		r1, r2 := titleRow(section.title)
		table.SetCell(currentRow, 0, r1)
		table.SetCell(currentRow, 1, r2)
		currentRow++

		// Add shortcuts for this section
		for _, shortcut := range section.shortcuts {
			r1, r2 := shortcutRow(shortcut[0], shortcut[1])
			table.SetCell(currentRow, 0, r1)
			table.SetCell(currentRow, 1, r2)
			currentRow++
		}
	}

	versionString := fmt.Sprintf("[-:-:b]Sake %s", Version)
	text := tview.NewTextView()
	text.SetDynamicColors(true)
	text.SetText(versionString).SetTextAlign(tview.AlignRight)

	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(text, 1, 0, true).
		AddItem(table, 0, 1, true)

	return root, table
}
