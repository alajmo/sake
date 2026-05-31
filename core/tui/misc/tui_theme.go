package misc

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// StyleOption holds styling information for TUI elements
type StyleOption struct {
	Fg    tcell.Color
	Bg    tcell.Color
	Attr  tcell.AttrMask
	Align int

	FgStr   string
	BgStr   string
	AttrStr string

	Style tcell.Style
}

// Default styles
var STYLE_DEFAULT StyleOption

// Border styles
var STYLE_BORDER StyleOption
var STYLE_BORDER_FOCUS StyleOption

// Title styles
var STYLE_TITLE StyleOption
var STYLE_TITLE_ACTIVE StyleOption

// Table Header
var STYLE_TABLE_HEADER StyleOption

// Item styles
var STYLE_ITEM StyleOption
var STYLE_ITEM_FOCUSED StyleOption
var STYLE_ITEM_SELECTED StyleOption

// Button styles
var STYLE_BUTTON StyleOption
var STYLE_BUTTON_ACTIVE StyleOption

// Search styles
var STYLE_SEARCH_LABEL StyleOption
var STYLE_SEARCH_TEXT StyleOption

// Filter styles
var STYLE_FILTER_LABEL StyleOption
var STYLE_FILTER_TEXT StyleOption

// Shortcut styles
var STYLE_SHORTCUT_LABEL StyleOption
var STYLE_SHORTCUT_TEXT StyleOption

// LoadStyles initializes all TUI styles with mani-style defaults
func LoadStyles() {
	// Mani-style colors
	magenta := tcell.GetColor("#d787ff")
	darkGray := tcell.GetColor("#262626")
	selectedBlue := tcell.GetColor("#5f87d7")
	labelYellow := tcell.GetColor("#d7d75f")
	shortcutGreen := tcell.GetColor("#00af5f")
	nearBlack := tcell.GetColor("#080808")

	// Default
	STYLE_DEFAULT = StyleOption{
		Fg:    tcell.ColorDefault,
		Bg:    tcell.ColorDefault,
		Attr:  tcell.AttrNone,
		Align: tview.AlignLeft,
		FgStr: "",
		BgStr: "",
		Style: tcell.StyleDefault,
	}

	// Border
	STYLE_BORDER = StyleOption{
		Fg:    tcell.ColorDefault,
		Bg:    tcell.ColorDefault,
		Attr:  tcell.AttrNone,
		FgStr: "",
		BgStr: "",
		Style: tcell.StyleDefault,
	}

	STYLE_BORDER_FOCUS = StyleOption{
		Fg:    magenta,
		Bg:    tcell.ColorDefault,
		Attr:  tcell.AttrNone,
		FgStr: "#d787ff",
		BgStr: "",
		Style: tcell.StyleDefault.Foreground(magenta),
	}

	// Title
	STYLE_TITLE = StyleOption{
		Fg:    tcell.ColorDefault,
		Bg:    tcell.ColorDefault,
		Attr:  tcell.AttrNone,
		Align: tview.AlignCenter,
		FgStr: "",
		BgStr: "",
		Style: tcell.StyleDefault,
	}

	STYLE_TITLE_ACTIVE = StyleOption{
		Fg:    tcell.ColorBlack,
		Bg:    magenta,
		Attr:  tcell.AttrNone,
		Align: tview.AlignCenter,
		FgStr: "#000000",
		BgStr: "#d787ff",
		Style: tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(magenta),
	}

	// Table Header
	STYLE_TABLE_HEADER = StyleOption{
		Fg:    magenta,
		Bg:    tcell.ColorDefault,
		Attr:  tcell.AttrBold,
		Align: tview.AlignLeft,
		FgStr: "#d787ff",
		BgStr: "",
		Style: tcell.StyleDefault.Foreground(magenta).Attributes(tcell.AttrBold),
	}

	// Item
	STYLE_ITEM = StyleOption{
		Fg:    tcell.ColorDefault,
		Bg:    tcell.ColorDefault,
		Attr:  tcell.AttrNone,
		FgStr: "",
		BgStr: "",
		Style: tcell.StyleDefault,
	}

	STYLE_ITEM_FOCUSED = StyleOption{
		Fg:    tcell.ColorWhite,
		Bg:    darkGray,
		Attr:  tcell.AttrNone,
		FgStr: "#ffffff",
		BgStr: "#262626",
		Style: tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(darkGray),
	}

	STYLE_ITEM_SELECTED = StyleOption{
		Fg:    selectedBlue,
		Bg:    tcell.ColorDefault,
		Attr:  tcell.AttrNone,
		FgStr: "#5f87d7",
		BgStr: "",
		Style: tcell.StyleDefault.Foreground(selectedBlue),
	}

	// Button
	STYLE_BUTTON = StyleOption{
		Fg:    tcell.ColorDefault,
		Bg:    tcell.ColorDefault,
		Attr:  tcell.AttrNone,
		FgStr: "",
		BgStr: "",
		Style: tcell.StyleDefault,
	}

	STYLE_BUTTON_ACTIVE = StyleOption{
		Fg:    nearBlack,
		Bg:    magenta,
		Attr:  tcell.AttrNone,
		FgStr: "#080808",
		BgStr: "#d787ff",
		Style: tcell.StyleDefault.Foreground(nearBlack).Background(magenta),
	}

	// Search
	STYLE_SEARCH_LABEL = StyleOption{
		Fg:    labelYellow,
		Bg:    tcell.ColorDefault,
		Attr:  tcell.AttrBold,
		FgStr: "#d7d75f",
		BgStr: "",
		Style: tcell.StyleDefault.Foreground(labelYellow).Attributes(tcell.AttrBold),
	}

	STYLE_SEARCH_TEXT = StyleOption{
		Fg:    tcell.ColorDefault,
		Bg:    tcell.ColorDefault,
		Attr:  tcell.AttrNone,
		FgStr: "",
		BgStr: "",
		Style: tcell.StyleDefault,
	}

	// Filter
	STYLE_FILTER_LABEL = StyleOption{
		Fg:    labelYellow,
		Bg:    tcell.ColorDefault,
		Attr:  tcell.AttrBold,
		FgStr: "#d7d75f",
		BgStr: "",
		Style: tcell.StyleDefault.Foreground(labelYellow).Attributes(tcell.AttrBold),
	}

	STYLE_FILTER_TEXT = StyleOption{
		Fg:    tcell.ColorDefault,
		Bg:    tcell.ColorDefault,
		Attr:  tcell.AttrNone,
		FgStr: "",
		BgStr: "",
		Style: tcell.StyleDefault,
	}

	// Shortcut
	STYLE_SHORTCUT_LABEL = StyleOption{
		Fg:    shortcutGreen,
		Bg:    tcell.ColorDefault,
		Attr:  tcell.AttrNone,
		FgStr: "#00af5f",
		BgStr: "",
		Style: tcell.StyleDefault.Foreground(shortcutGreen),
	}

	STYLE_SHORTCUT_TEXT = StyleOption{
		Fg:    tcell.ColorDefault,
		Bg:    tcell.ColorDefault,
		Attr:  tcell.AttrNone,
		FgStr: "",
		BgStr: "",
		Style: tcell.StyleDefault,
	}
}

// Colorize wraps a value with tview color tags
func Colorize(value, fg, bg, attr string) string {
	return " [-:-:-]" + fmt.Sprintf("[%s:%s:%s]%s", fg, bg, attr, value) + "[-:-:-] "
}

// ColorizeTitle wraps a title with tview color tags
func ColorizeTitle(value, fg, bg, attr string) string {
	return " [-:-:-]" + fmt.Sprintf("[%s:%s:%s] %s ", fg, bg, attr, value) + "[-:-:-] "
}

// PadString adds space padding around a string
func PadString(name string) string {
	return " " + strings.TrimSpace(name) + " "
}

// SetActive updates the title style based on active state
func SetActive(box *tview.Box, title string, active bool) {
	if active {
		box.SetTitle(ColorizeTitle(title, STYLE_TITLE_ACTIVE.FgStr, STYLE_TITLE_ACTIVE.BgStr, "b"))
		box.SetBorderColor(STYLE_BORDER_FOCUS.Fg)
	} else {
		box.SetTitle(ColorizeTitle(title, STYLE_TITLE.FgStr, STYLE_TITLE.BgStr, "-"))
		box.SetBorderColor(STYLE_BORDER.Fg)
	}
}
