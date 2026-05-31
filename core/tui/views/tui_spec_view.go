package views

import (
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/tui/components"
	"github.com/alajmo/sake/core/tui/misc"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TSpec holds the execution options for the TUI
type TSpec struct {
	View  *tview.Flex
	items []*tview.Box

	// Options
	Output            string
	Strategy          string
	ClearBeforeRun    bool
	IgnoreErrors      bool
	IgnoreUnreachable bool
	OmitEmptyRows     bool
	OmitEmptyColumns  bool
	AnyErrorsFatal    bool
}

// CreateSpecView creates the options view with all spec options
func CreateSpecView() *TSpec {
	defSpec := dao.DEFAULT_SPEC

	spec := &TSpec{
		Output:            defSpec.Output,
		Strategy:          defSpec.Strategy,
		ClearBeforeRun:    false,
		IgnoreErrors:      defSpec.IgnoreErrors,
		IgnoreUnreachable: defSpec.IgnoreUnreachable,
		OmitEmptyRows:     defSpec.OmitEmptyRows,
		OmitEmptyColumns:  defSpec.OmitEmptyColumns,
		AnyErrorsFatal:    defSpec.AnyErrorsFatal,
	}

	view := tview.NewFlex().SetDirection(tview.FlexRow)
	view.SetBorder(false).SetBorderPadding(0, 0, 0, 0)
	view.SetBackgroundColor(misc.STYLE_DEFAULT.Bg)
	spec.View = view

	// Output type toggle (text/table)
	outputType := &components.TToggleText{
		Value:   &spec.Output,
		Option1: "text",
		Option2: "table",
		Label1:  " Output: text ",
		Label2:  " Output: table ",
	}
	outputType.Create()

	// Strategy toggle (linear/free/host_pinned)
	strategyType := &components.TToggleThree{
		Value:   &spec.Strategy,
		Option1: "linear",
		Option2: "free",
		Option3: "host_pinned",
		Label1:  " Strategy: linear ",
		Label2:  " Strategy: free ",
		Label3:  " Strategy: host_pinned ",
	}
	strategyType.Create()

	// Checkboxes
	clearBeforeRun := spec.AddCheckbox("Clear Before Run", &spec.ClearBeforeRun)
	ignoreErrors := spec.AddCheckbox("Ignore Errors", &spec.IgnoreErrors)
	ignoreUnreachable := spec.AddCheckbox("Ignore Unreachable", &spec.IgnoreUnreachable)
	omitEmptyRows := spec.AddCheckbox("Omit Empty Rows", &spec.OmitEmptyRows)
	omitEmptyColumns := spec.AddCheckbox("Omit Empty Columns", &spec.OmitEmptyColumns)
	anyErrorsFatal := spec.AddCheckbox("Any Errors Fatal", &spec.AnyErrorsFatal)

	// Add items to view
	view.AddItem(outputType.TextView, 1, 0, false)
	view.AddItem(strategyType.TextView, 1, 0, false)
	view.AddItem(clearBeforeRun, 1, 0, false)
	view.AddItem(ignoreErrors, 1, 0, false)
	view.AddItem(ignoreUnreachable, 1, 0, false)
	view.AddItem(omitEmptyRows, 1, 0, false)
	view.AddItem(omitEmptyColumns, 1, 0, false)
	view.AddItem(anyErrorsFatal, 1, 0, false)

	focusItems := []*tview.Box{
		outputType.TextView.Box,
		strategyType.TextView.Box,
		clearBeforeRun.Box,
		ignoreErrors.Box,
		ignoreUnreachable.Box,
		omitEmptyRows.Box,
		omitEmptyColumns.Box,
		anyErrorsFatal.Box,
	}

	// Input handling for navigation
	currentFocus := 0
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown:
			if currentFocus < (len(focusItems) - 1) {
				currentFocus += 1
				misc.App.SetFocus(focusItems[currentFocus])
			}
			return nil
		case tcell.KeyUp:
			if currentFocus > 0 {
				currentFocus -= 1
				misc.App.SetFocus(focusItems[currentFocus])
			}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'g': // top
				currentFocus = 0
				misc.App.SetFocus(focusItems[currentFocus])
				return nil
			case 'G': // bottom
				currentFocus = len(focusItems) - 1
				misc.App.SetFocus(focusItems[currentFocus])
				return nil
			case 'j': // down
				if currentFocus < (len(focusItems) - 1) {
					currentFocus += 1
					misc.App.SetFocus(focusItems[currentFocus])
				}
				return nil
			case 'k': // up
				if currentFocus > 0 {
					currentFocus -= 1
					misc.App.SetFocus(focusItems[currentFocus])
				}
				return nil
			}
		}

		return event
	})

	view.SetFocusFunc(func() {
		currentFocus = 0
		misc.App.SetFocus(outputType.TextView)
	})

	return spec
}

// AddCheckbox adds a checkbox to the spec view
func (spec *TSpec) AddCheckbox(title string, checked *bool) *tview.Checkbox {
	onFocus := func() {}
	onBlur := func() {}

	checkbox := components.Checkbox(title, checked, onFocus, onBlur)
	spec.items = append(spec.items, checkbox.Box)
	return checkbox
}
