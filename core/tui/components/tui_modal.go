package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/term"

	"github.com/alajmo/sake/core/tui/misc"
)

// OpenTextModal opens a modal with text content
func OpenTextModal(pageTitle string, text string, title string) {
	width, height := getTextModalSize(text)
	text = strings.TrimSpace(text)

	// Text
	contentPane := tview.NewTextView().
		SetText(text).
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)

	// Border
	formattedTitle := misc.ColorizeTitle(title, misc.STYLE_TITLE_ACTIVE.FgStr, misc.STYLE_TITLE_ACTIVE.BgStr, "b")
	contentPane.SetBorder(true).
		SetTitle(formattedTitle).
		SetTitleAlign(misc.STYLE_TITLE.Align).
		SetBorderColor(misc.STYLE_BORDER_FOCUS.Fg).
		SetBorderPadding(1, 1, 2, 2)

	// Use a background box with draw function for proper rendering (no color = terminal default)
	background := tview.NewBox()
	containerFlex := tview.NewFlex().
		AddItem(contentPane, 0, 1, true)
	containerFlex.SetDrawFunc(func(screen tcell.Screen, x, y, w, h int) (int, int, int, int) {
		background.SetRect(x, y, w, h)
		background.Draw(screen)
		contentPane.SetRect(x, y, w, h)
		contentPane.Draw(screen)
		return x, y, w, h
	})

	// Container
	modal := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(nil, 0, 1, false).
				AddItem(containerFlex, width, 1, true).
				AddItem(nil, 0, 1, false),
			height, 1, true,
		).
		AddItem(nil, 0, 1, false)

	modal.SetFullScreen(true)

	EmptySearch()

	misc.Pages.AddPage(pageTitle, modal, false, true)
	misc.App.SetFocus(contentPane)
}

// CloseModal closes the front modal
func CloseModal() {
	previousPane := misc.PreviousPane
	frontPageName, _ := misc.Pages.GetFrontPage()
	misc.Pages.RemovePage(frontPageName)
	misc.App.SetFocus(previousPane)
}

// IsModalOpen checks if a modal is currently open
func IsModalOpen() bool {
	frontPageName, _ := misc.Pages.GetFrontPage()
	return strings.Contains(frontPageName, "-modal")
}

// IsDescribeModalOpen checks if the describe modal is open
func IsDescribeModalOpen() bool {
	frontPageName, _ := misc.Pages.GetFrontPage()
	return strings.Contains(frontPageName, "describe") || strings.Contains(frontPageName, "description")
}

// CloseDescribeModal closes the describe modal if open
func CloseDescribeModal() bool {
	frontPageName, _ := misc.Pages.GetFrontPage()
	if strings.Contains(frontPageName, "describe") || strings.Contains(frontPageName, "description") {
		CloseModal()
		return true
	}
	return false
}

// OpenModal opens a modal with custom content
func OpenModal(pageTitle string, title string, content tview.Primitive, width int, height int) {
	termWidth, termHeight, _ := term.GetSize(0)
	if termWidth == 0 {
		termWidth = 80
	}
	if termHeight == 0 {
		termHeight = 24
	}
	if width > termWidth {
		width = termWidth - 5
	}
	if height > termHeight {
		height = termHeight - 5
	}

	// Border wrapper
	wrapper := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(content, 0, 1, true)

	formattedTitle := misc.ColorizeTitle(title, misc.STYLE_TITLE_ACTIVE.FgStr, misc.STYLE_TITLE_ACTIVE.BgStr, "b")
	wrapper.SetBorder(true).
		SetTitle(formattedTitle).
		SetTitleAlign(misc.STYLE_TITLE.Align).
		SetBorderColor(misc.STYLE_BORDER_FOCUS.Fg).
		SetBorderPadding(1, 1, 2, 2)

	// Use a background box with draw function for proper rendering (no color = terminal default)
	background := tview.NewBox()
	containerFlex := tview.NewFlex().
		AddItem(wrapper, 0, 1, true)
	containerFlex.SetDrawFunc(func(screen tcell.Screen, x, y, w, h int) (int, int, int, int) {
		background.SetRect(x, y, w, h)
		background.Draw(screen)
		wrapper.SetRect(x, y, w, h)
		wrapper.Draw(screen)
		return x, y, w, h
	})

	// Container
	modal := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(nil, 0, 1, false).
				AddItem(containerFlex, width, 1, true).
				AddItem(nil, 0, 1, false),
			height, 1, true,
		).
		AddItem(nil, 0, 1, false)

	modal.SetFullScreen(true)

	EmptySearch()

	misc.Pages.AddPage(pageTitle, modal, false, true)
	misc.App.SetFocus(content)
}

// getTextModalSize calculates appropriate modal dimensions based on content
func getTextModalSize(text string) (int, int) {
	termWidth, termHeight, _ := term.GetSize(0)
	if termWidth == 0 {
		termWidth = 80
	}
	if termHeight == 0 {
		termHeight = 24
	}

	lines := strings.Split(text, "\n")
	height := len(lines) + 6 // padding for borders and title

	width := 40
	for _, line := range lines {
		lineLen := len(line) + 6
		if lineLen > width {
			width = lineLen
		}
	}

	maxWidth := termWidth - 10
	maxHeight := termHeight - 5

	if width > maxWidth {
		width = maxWidth
	}
	if height > maxHeight {
		height = maxHeight
	}

	return width, height
}
