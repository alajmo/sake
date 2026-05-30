package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/sake/core/tui/misc"
)

type TOutput struct {
	Root   *tview.Flex
	Output *tview.TextView
	Title  string
}

func (t *TOutput) Create() {
	output := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)

	output.SetBackgroundColor(misc.STYLE_ITEM.Bg)
	output.SetTextColor(misc.STYLE_ITEM.Fg)

	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(output, 0, 1, true)

	root.SetTitleColor(misc.STYLE_TITLE.Fg)
	root.SetTitleAlign(misc.STYLE_TITLE.Align).
		SetBorder(true).
		SetBorderPadding(1, 0, 1, 1)

	t.Output = output
	t.Root = root

	if t.Title != "" {
		misc.SetActive(t.Root.Box, t.Title, false)
	}

	// Input
	t.Output.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlD:
			row, _ := output.GetScrollOffset()
			_, _, _, height := output.GetInnerRect()
			output.ScrollTo(row+height/2, 0)
			return nil
		case tcell.KeyCtrlU:
			row, _ := output.GetScrollOffset()
			_, _, _, height := output.GetInnerRect()
			newRow := row - height/2
			if newRow < 0 {
				newRow = 0
			}
			output.ScrollTo(newRow, 0)
			return nil
		case tcell.KeyCtrlF:
			row, _ := output.GetScrollOffset()
			_, _, _, height := output.GetInnerRect()
			output.ScrollTo(row+height, 0)
			return nil
		case tcell.KeyCtrlB:
			row, _ := output.GetScrollOffset()
			_, _, _, height := output.GetInnerRect()
			newRow := row - height
			if newRow < 0 {
				newRow = 0
			}
			output.ScrollTo(newRow, 0)
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				row, col := output.GetScrollOffset()
				output.ScrollTo(row+1, col)
				return nil
			case 'k':
				row, col := output.GetScrollOffset()
				if row > 0 {
					output.ScrollTo(row-1, col)
				}
				return nil
			case 'g':
				output.ScrollToBeginning()
				return nil
			case 'G':
				output.ScrollToEnd()
				return nil
			}
		}
		return event
	})

	t.Output.SetFocusFunc(func() {
		misc.PreviousPane = t.Output
		misc.SetActive(t.Root.Box, t.Title, true)
	})

	t.Output.SetBlurFunc(func() {
		misc.PreviousPane = t.Output
		misc.SetActive(t.Root.Box, t.Title, false)
	})
}

func (t *TOutput) Clear() {
	t.Output.Clear()
}

func (t *TOutput) Write(text string) {
	t.Output.SetText(text)
}

func (t *TOutput) GetWriter() *misc.ThreadSafeWriter {
	return misc.NewThreadSafeWriter(t.Output)
}
