package print

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/alajmo/sake/core/dao"
)

func CreateTable(
	options PrintTableOptions,
	tableHeaders []string,
) table.Writer {
	t := table.NewWriter()

	theme := options.Theme

	t.SetOutputMirror(os.Stdout)

	t.SetStyle(FormatTable(theme))
	if options.OmitEmptyColumns {
		t.SuppressEmptyColumns()
	}

	var headers []table.ColumnConfig
	for i := range tableHeaders {
		hh := table.ColumnConfig{
			Number:       i + 1,
			AlignHeader:  GetAlign(*theme.Table.Header.Align),
			Align:        GetAlign(*theme.Table.Row.Align),
			ColorsHeader: combineColors(theme.Table.Header.Fg, theme.Table.Header.Bg, theme.Table.Header.Attr),
			Colors:       combineColors(theme.Table.Row.Fg, theme.Table.Row.Bg, theme.Table.Row.Attr),
			ColorsFooter: combineColors(theme.Table.Footer.Fg, theme.Table.Footer.Bg, theme.Table.Footer.Attr),
		}

		headers = append(headers, hh)
	}

	t.SetColumnConfigs(headers)

	return t
}

func FormatTable(theme dao.Theme) table.Style {
	return table.Style{
		Name: theme.Name,
		Box:  theme.Table.Box,

		Format: table.FormatOptions{
			Header: GetFormat(*theme.Table.Header.Format),
			Row:    GetFormat(*theme.Table.Row.Format),
			Footer: GetFormat(*theme.Table.Footer.Format),
		},

		Options: table.Options{
			DrawBorder:      *theme.Table.Options.DrawBorder,
			SeparateColumns: *theme.Table.Options.SeparateColumns,
			SeparateHeader:  *theme.Table.Options.SeparateHeader,
			SeparateRows:    *theme.Table.Options.SeparateRows,
			SeparateFooter:  *theme.Table.Options.SeparateFooter,
		},

		Title: table.TitleOptions{
			Align:  GetAlign(*theme.Table.Title.Align),
			Colors: combineColors(theme.Table.Title.Fg, theme.Table.Title.Bg, theme.Table.Title.Attr),
		},

		// Border colors
		Color: table.ColorOptions{
			Header:       combineColors(theme.Table.Border.Header.Fg, theme.Table.Border.Header.Bg, theme.Table.Border.Header.Attr),
			Row:          combineColors(theme.Table.Border.Row.Fg, theme.Table.Border.Row.Bg, theme.Table.Border.Row.Attr),
			RowAlternate: combineColors(theme.Table.Border.RowAlternate.Fg, theme.Table.Border.RowAlternate.Bg, theme.Table.Border.RowAlternate.Attr),
			Footer:       combineColors(theme.Table.Border.Footer.Fg, theme.Table.Border.Footer.Bg, theme.Table.Border.Footer.Attr),
		},
	}
}

func RenderTable(
	t table.Writer,
	output string,
	padTop bool,
	padBottom bool,
) {
	switch output {
	case "markdown":
		t.RenderMarkdown()
	case "html":
		t.RenderHTML()
	case "csv":
		t.RenderCSV()
	default:
		if padTop {
			fmt.Println()
		}
		t.Render()
		if padBottom {
			fmt.Println()
		}
	}
}
