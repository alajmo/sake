package misc

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TItem struct {
	Primitive tview.Primitive
	Box       *tview.Box
}

func FocusNext(elements []*TItem) *tview.Primitive {
	if len(elements) == 0 {
		return nil
	}

	currentFocus := App.GetFocus()
	nextIndex := -1
	for i, element := range elements {
		if element.Primitive == currentFocus {
			nextIndex = (i + 1) % len(elements)
		}
		element.Box.SetBorderColor(STYLE_BORDER.Fg)
	}

	// Current focus isn't among these elements (e.g. right after a page switch):
	// fall back to the previously focused pane as the anchor.
	if nextIndex < 0 {
		for i, element := range elements {
			if element.Primitive == PreviousPane {
				nextIndex = (i + 1) % len(elements)
			}
		}
	}

	// Neither anchor was found; default to the first element so Tab keeps
	// working and we never dereference a nil Box.
	if nextIndex < 0 {
		nextIndex = 0
	}

	nextFocusItem := elements[nextIndex]
	nextFocusItem.Box.SetBorderColor(STYLE_BORDER_FOCUS.Fg)
	App.SetFocus(nextFocusItem.Primitive)

	return &nextFocusItem.Primitive
}

func FocusPrevious(elements []*TItem) *tview.Primitive {
	if len(elements) == 0 {
		return nil
	}

	currentFocus := App.GetFocus()
	prevIndex := -1
	for i, element := range elements {
		if element.Primitive == currentFocus {
			prevIndex = (i - 1 + len(elements)) % len(elements)
		}
		element.Box.SetBorderColor(STYLE_BORDER.Fg)
	}

	// Only fall back to the previous pane when current focus wasn't found,
	// otherwise this would override the correct result.
	if prevIndex < 0 {
		for i, element := range elements {
			if element.Primitive == PreviousPane {
				prevIndex = (i - 1 + len(elements)) % len(elements)
			}
		}
	}

	// Neither anchor was found; default to the first element so Shift-Tab keeps
	// working and we never dereference a nil Box.
	if prevIndex < 0 {
		prevIndex = 0
	}

	nextFocusItem := elements[prevIndex]
	nextFocusItem.Box.SetBorderColor(STYLE_BORDER_FOCUS.Fg)
	App.SetFocus(nextFocusItem.Primitive)

	return &nextFocusItem.Primitive
}

func FocusPage(event *tcell.EventKey, focusable []*TItem) {
	i := int(event.Rune()-'0') - 1
	if i < len(focusable) {
		App.SetFocus(focusable[i].Box)
	}
}

func FocusPreviousPage() {
	App.SetFocus(PreviousPane)
}

func GetTUIItem(primitive tview.Primitive, box *tview.Box) *TItem {
	return &TItem{
		Primitive: primitive,
		Box:       box,
	}
}
