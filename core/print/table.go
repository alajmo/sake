package print

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/alajmo/sake/core/dao"
)

func CreateTable(
	options PrintTableOptions,
	defaultHeaders []string,
	taskHeaders []string,
) table.Writer {
	t := table.NewWriter()

	theme := options.Theme

	t.SetOutputMirror(os.Stdout)
	t.SetStyle(FormatTable(theme))
	if options.SuppressEmptyColumns {
		t.SuppressEmptyColumns()
	}

	headerStyles := make(map[string]table.ColumnConfig)
	for _, h := range defaultHeaders {
		switch h {
		case "server":
			headerStyles[h] = table.ColumnConfig{
				Name: "server",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.Server.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.Server.Fg, theme.Table.Color.Header.Server.Bg, theme.Table.Color.Header.Server.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.Server.Align),
				Colors: combineColors(theme.Table.Color.Row.Server.Fg, theme.Table.Color.Row.Server.Bg, theme.Table.Color.Row.Server.Attr),
			}
		case "tag":
			headerStyles[h] = table.ColumnConfig{
				Name: "tag",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.Tag.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.Tag.Fg, theme.Table.Color.Header.Tag.Bg, theme.Table.Color.Header.Tag.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.Tag.Align),
				Colors: combineColors(theme.Table.Color.Row.Tag.Fg, theme.Table.Color.Row.Tag.Bg, theme.Table.Color.Row.Tag.Attr),
			}
		case "description":
			headerStyles[h] = table.ColumnConfig{
				Name: "description",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.Desc.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.Desc.Fg, theme.Table.Color.Header.Desc.Bg, theme.Table.Color.Header.Desc.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.Desc.Align),
				Colors: combineColors(theme.Table.Color.Row.Desc.Fg, theme.Table.Color.Row.Desc.Bg, theme.Table.Color.Row.Desc.Attr),
			}
		case "host":
			headerStyles[h] = table.ColumnConfig{
				Name: "host",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.Host.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.Host.Fg, theme.Table.Color.Header.Host.Bg, theme.Table.Color.Header.Host.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.Host.Align),
				Colors: combineColors(theme.Table.Color.Row.Host.Fg, theme.Table.Color.Row.Host.Bg, theme.Table.Color.Row.Host.Attr),
			}
		case "user":
			headerStyles[h] = table.ColumnConfig{
				Name: "user",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.User.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.User.Fg, theme.Table.Color.Header.User.Bg, theme.Table.Color.Header.User.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.User.Align),
				Colors: combineColors(theme.Table.Color.Row.User.Fg, theme.Table.Color.Row.User.Bg, theme.Table.Color.Row.User.Attr),
			}
		case "port":
			headerStyles[h] = table.ColumnConfig{
				Name: "port",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.Port.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.Port.Fg, theme.Table.Color.Header.Port.Bg, theme.Table.Color.Header.Port.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.Port.Align),
				Colors: combineColors(theme.Table.Color.Row.Port.Fg, theme.Table.Color.Row.Port.Bg, theme.Table.Color.Row.Port.Attr),
			}
		case "local":
			headerStyles[h] = table.ColumnConfig{
				Name: "local",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.Local.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.Local.Fg, theme.Table.Color.Header.Local.Bg, theme.Table.Color.Header.Local.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.Local.Align),
				Colors: combineColors(theme.Table.Color.Row.Local.Fg, theme.Table.Color.Row.Local.Bg, theme.Table.Color.Row.Local.Attr),
			}
		case "task":
			headerStyles[h] = table.ColumnConfig{
				Name: "task",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.Task.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.Task.Fg, theme.Table.Color.Header.Task.Bg, theme.Table.Color.Header.Task.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.Task.Align),
				Colors: combineColors(theme.Table.Color.Row.Task.Fg, theme.Table.Color.Row.Task.Bg, theme.Table.Color.Row.Task.Attr),
			}
		}
	}

	headers := []table.ColumnConfig{}
	for _, h := range defaultHeaders {
		headers = append(headers, headerStyles[h])
	}

	for i := range taskHeaders {
		hh := table.ColumnConfig{
			Number:       len(defaultHeaders) + 1 + i,
			AlignHeader:  GetAlign(*theme.Table.Color.Header.Output.Align),
			ColorsHeader: combineColors(theme.Table.Color.Header.Output.Fg, theme.Table.Color.Header.Output.Bg, theme.Table.Color.Header.Output.Attr),

			Align:  GetAlign(*theme.Table.Color.Row.Output.Align),
			Colors: combineColors(theme.Table.Color.Row.Output.Fg, theme.Table.Color.Row.Output.Bg, theme.Table.Color.Row.Output.Attr),
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
			Header: GetFormat(*theme.Table.Format.Header),
			Row:    GetFormat(*theme.Table.Format.Row),
		},

		Options: table.Options{
			DrawBorder:      *theme.Table.Options.DrawBorder,
			SeparateColumns: *theme.Table.Options.SeparateColumns,
			SeparateHeader:  *theme.Table.Options.SeparateHeader,
			SeparateRows:    *theme.Table.Options.SeparateRows,
		},

		// Border colors
		Color: table.ColorOptions{
			Header:       combineColors(theme.Table.Color.Border.Header.Fg, theme.Table.Color.Border.Header.Bg, theme.Table.Color.Border.Header.Attr),
			Row:          combineColors(theme.Table.Color.Border.Row.Fg, theme.Table.Color.Border.Row.Bg, theme.Table.Color.Border.Row.Attr),
			RowAlternate: combineColors(theme.Table.Color.Border.RowAlternate.Fg, theme.Table.Color.Border.RowAlternate.Bg, theme.Table.Color.Border.RowAlternate.Attr),
			Footer:       combineColors(theme.Table.Color.Border.Footer.Fg, theme.Table.Color.Border.Footer.Bg, theme.Table.Color.Border.Footer.Attr),
		},
	}
}

func RenderTable(t table.Writer, output string) {
	fmt.Println()
	switch output {
	case "markdown":
		t.RenderMarkdown()
	case "html":
		t.RenderHTML()
	default:
		t.Render()
	}
	fmt.Println()
}
