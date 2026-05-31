package views

import (
	"fmt"
	"strings"

	"github.com/alajmo/sake/core/tui/misc"
	"github.com/rivo/tview"
)

type Shortcut struct {
	label    string
	shortcut string
}

func getShortcutInfo(shortcuts []Shortcut) string {
	var formattedShortcuts []string
	for _, s := range shortcuts {
		value := fmt.Sprintf("[%s:%s:%s]%s[-:-:-] [%s:%s:%s]%s[-:-:-]",
			misc.STYLE_SHORTCUT_TEXT.FgStr, misc.STYLE_SHORTCUT_TEXT.BgStr, misc.STYLE_SHORTCUT_TEXT.AttrStr, s.shortcut,
			misc.STYLE_SHORTCUT_LABEL.FgStr, misc.STYLE_SHORTCUT_LABEL.BgStr, misc.STYLE_SHORTCUT_LABEL.AttrStr, s.label,
		)
		formattedShortcuts = append(formattedShortcuts, value)
	}
	return strings.Join(formattedShortcuts, "  ")
}

func CreateRunInfoView() *tview.TextView {
	shortcuts := []Shortcut{
		{"Ctrl-r", "Run"},
		{"Ctrl-s", "Toggle View"},
		{"Ctrl-e", "Toggle Table/Tree"},
		{"Ctrl-o", "Options"},
	}
	text := getShortcutInfo(shortcuts)

	helpInfo := tview.NewTextView().
		SetDynamicColors(true).
		SetText(text)
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)
	return helpInfo
}

func CreateExecInfoView() *tview.TextView {
	shortcuts := []Shortcut{
		{"Ctrl-r", "Run"},
		{"Ctrl-x", "Clear"},
		{"Ctrl-o", "Options"},
	}
	text := getShortcutInfo(shortcuts)

	helpInfo := tview.NewTextView().
		SetDynamicColors(true).
		SetText(text)
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)
	return helpInfo
}

func CreateServersInfoView() *tview.TextView {
	shortcuts := []Shortcut{
		{"Shift-S", "SSH to Server"},
		{"Ctrl-e", "Toggle Table/Tree"},
	}
	text := getShortcutInfo(shortcuts)

	helpInfo := tview.NewTextView().
		SetDynamicColors(true).
		SetText(text)
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)
	return helpInfo
}

func CreateTasksInfoView() *tview.TextView {
	shortcuts := []Shortcut{
		{"Ctrl-e", "Toggle Table/Tree"},
	}
	text := getShortcutInfo(shortcuts)

	helpInfo := tview.NewTextView().
		SetDynamicColors(true).
		SetText(text)
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)
	return helpInfo
}
