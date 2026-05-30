package misc

import (
	"strings"

	"github.com/rivo/tview"
)

// StyleFormat applies formatting to a string (uppercase, lowercase, title case)
func StyleFormat(value, format string) string {
	switch strings.ToLower(format) {
	case "upper", "uppercase":
		return strings.ToUpper(value)
	case "lower", "lowercase":
		return strings.ToLower(value)
	case "title", "titlecase":
		return strings.Title(value)
	default:
		return value
	}
}

// GetModalSize calculates appropriate modal dimensions based on content
func GetModalSize(content string, minWidth, minHeight, maxWidth, maxHeight int) (int, int) {
	lines := strings.Split(content, "\n")
	height := len(lines) + 4 // padding

	width := minWidth
	for _, line := range lines {
		lineLen := len(line) + 4
		if lineLen > width {
			width = lineLen
		}
	}

	if width > maxWidth {
		width = maxWidth
	}
	if height > maxHeight {
		height = maxHeight
	}
	if width < minWidth {
		width = minWidth
	}
	if height < minHeight {
		height = minHeight
	}

	return width, height
}

// SetupStyles configures global tview styles
func SetupStyles() {
	// Foreground / Background
	tview.Styles.PrimaryTextColor = STYLE_DEFAULT.Fg
	tview.Styles.PrimitiveBackgroundColor = STYLE_DEFAULT.Bg

	// Borders Colors
	tview.Styles.BorderColor = STYLE_BORDER.Fg

	// Border style
	tview.Borders.HorizontalFocus = tview.BoxDrawingsLightHorizontal
	tview.Borders.VerticalFocus = tview.BoxDrawingsLightVertical
	tview.Borders.TopLeftFocus = tview.BoxDrawingsLightDownAndRight
	tview.Borders.TopRightFocus = tview.BoxDrawingsLightDownAndLeft
	tview.Borders.BottomLeftFocus = tview.BoxDrawingsLightUpAndRight
	tview.Borders.BottomRightFocus = tview.BoxDrawingsLightUpAndLeft
}
