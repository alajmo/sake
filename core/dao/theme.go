package dao

// TODO: This file needs refactoring

import (
	"errors"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"gopkg.in/yaml.v3"

	"github.com/alajmo/sake/core"
)

type Table struct {
	// Stylable via YAML
	Name    string        `yaml:"name"`
	Style   string        `yaml:"style"`
	Prefix  string        `yaml:"prefix"`
	Options *TableOptions `yaml:"options"`

	Border *BorderColors `yaml:"border"`

	Title  *CellColors `yaml:"title"`
	Header *CellColors `yaml:"header"`
	Row    *CellColors `yaml:"row"`
	Footer *CellColors `yaml:"footer"`

	// Not stylable via YAML
	Box table.BoxStyle `yaml:"-"`
}

type TableOptions struct {
	DrawBorder      *bool `yaml:"draw_border"`
	SeparateColumns *bool `yaml:"separate_columns"`
	SeparateHeader  *bool `yaml:"separate_header"`
	SeparateRows    *bool `yaml:"separate_rows"`
	SeparateFooter  *bool `yaml:"separate_footer"`
}

type BorderColors struct {
	Header       *CellColors `yaml:"header"`
	Row          *CellColors `yaml:"row"`
	RowAlternate *CellColors `yaml:"row_alt"`
	Footer       *CellColors `yaml:"footer"`
}

type CellColors struct {
	Fg     *string `yaml:"fg"`
	Bg     *string `yaml:"bg"`
	Align  *string `yaml:"align"`
	Attr   *string `yaml:"attr"`
	Format *string `yaml:"format"`
}

type Text struct {
	Prefix       string   `yaml:"prefix"`
	PrefixColors []string `yaml:"prefix_colors"`
	Header       string   `yaml:"header"`
	HeaderFiller string   `yaml:"header_filler"`
}

type Theme struct {
	Name  string `yaml:"name"`
	Table Table  `yaml:"table"`
	Text  Text   `yaml:"text"`

	context     string // config path
	contextLine int    // defined at
}

type Row struct {
	Columns []string
}

type TableOutput struct {
	Headers []string
	Rows    []Row
	Footers []string
}

func (t *Theme) GetContext() string {
	return t.context
}

func (t *Theme) GetContextLine() int {
	return t.contextLine
}

func (r Row) GetValue(_ string, i int) string {
	if i < len(r.Columns) {
		return r.Columns[i]
	}

	return ""
}

type Items interface {
	GetValue(string, int) string
}

func GetTableData[T Items](items []T, headers []string) []Row {
	var rows []Row
	for i, s := range items {
		rows = append(rows, Row{Columns: []string{}})
		for _, h := range headers {
			rows[i].Columns = append(rows[i].Columns, s.GetValue(h, 0))
		}
	}
	return rows
}

// Table Box Styles

var StyleBoxLight = table.BoxStyle{
	BottomLeft:       "└",
	BottomRight:      "┘",
	BottomSeparator:  "┴",
	EmptySeparator:   text.RepeatAndTrim(" ", text.RuneCount("┼")),
	Left:             "│",
	LeftSeparator:    "├",
	MiddleHorizontal: "─",
	MiddleSeparator:  "┼",
	MiddleVertical:   "│",
	PaddingLeft:      " ",
	PaddingRight:     " ",
	PageSeparator:    "\n",
	Right:            "│",
	RightSeparator:   "┤",
	TopLeft:          "┌",
	TopRight:         "┐",
	TopSeparator:     "┬",
	UnfinishedRow:    " ≈",
}

var StyleBoxASCII = table.BoxStyle{
	BottomLeft:       "+",
	BottomRight:      "+",
	BottomSeparator:  "+",
	EmptySeparator:   text.RepeatAndTrim(" ", text.RuneCount("+")),
	Left:             "|",
	LeftSeparator:    "+",
	MiddleHorizontal: "-",
	MiddleSeparator:  "+",
	MiddleVertical:   "|",
	PaddingLeft:      " ",
	PaddingRight:     " ",
	PageSeparator:    "\n",
	Right:            "|",
	RightSeparator:   "+",
	TopLeft:          "+",
	TopRight:         "+",
	TopSeparator:     "+",
	UnfinishedRow:    " ~",
}

var DefaultText = Text{
	Prefix:       `{{ .Host }}`,
	PrefixColors: []string{"green", "blue", "red", "yellow", "magenta", "cyan"},
	Header:       `{{ .Style "TASK" "bold" }}{{ if ne .NumTasks 1 }} ({{ .Index }}/{{ .NumTasks }}){{end}}{{ if and .Name .Desc }} [{{.Style .Name "bold"}}: {{ .Desc }}] {{ else if .Name }} [{{ .Name }}] {{ else if .Desc }} [{{ .Desc }}] {{end}}`,
	HeaderFiller: "*",
}

var DefaultTable = Table{
	Style:  "default",
	Box:    StyleBoxASCII,
	Prefix: `{{ .Host }}`,

	Options: &TableOptions{
		DrawBorder:      core.Ptr(false),
		SeparateColumns: core.Ptr(true),
		SeparateHeader:  core.Ptr(true),
		SeparateRows:    core.Ptr(false),
		SeparateFooter:  core.Ptr(false),
	},

	Border: &BorderColors{
		Header: &CellColors{
			Fg:   core.Ptr(""),
			Bg:   core.Ptr(""),
			Attr: core.Ptr("faint"),
		},

		Row: &CellColors{
			Fg:   core.Ptr(""),
			Bg:   core.Ptr(""),
			Attr: core.Ptr("faint"),
		},

		RowAlternate: &CellColors{
			Fg:   core.Ptr(""),
			Bg:   core.Ptr(""),
			Attr: core.Ptr("faint"),
		},

		Footer: &CellColors{
			Fg:   core.Ptr(""),
			Bg:   core.Ptr(""),
			Attr: core.Ptr("faint"),
		},
	},

	Title: &CellColors{
		Fg:    core.Ptr(""),
		Bg:    core.Ptr(""),
		Align: core.Ptr(""),
		Attr:  core.Ptr("bold"),
	},

	Header: &CellColors{
		Fg:     core.Ptr(""),
		Bg:     core.Ptr(""),
		Align:  core.Ptr(""),
		Attr:   core.Ptr("bold"),
		Format: core.Ptr("default"),
	},

	Row: &CellColors{
		Fg:     core.Ptr(""),
		Bg:     core.Ptr(""),
		Align:  core.Ptr(""),
		Attr:   core.Ptr("normal"),
		Format: core.Ptr("default"),
	},

	Footer: &CellColors{
		Fg:     core.Ptr(""),
		Bg:     core.Ptr(""),
		Align:  core.Ptr(""),
		Attr:   core.Ptr("normal"),
		Format: core.Ptr("default"),
	},
}

// Populates ThemeList
func (c *ConfigYAML) ParseThemesYAML() ([]Theme, []ResourceErrors[Theme]) {
	var themes []Theme
	count := len(c.Themes.Content)

	themeErrors := []ResourceErrors[Theme]{}
	j := -1
	for i := 0; i < count; i += 2 {
		j += 1

		theme := &Theme{
			Name:        c.Themes.Content[i].Value,
			context:     c.Path,
			contextLine: c.Themes.Content[i].Line,
		}
		re := ResourceErrors[Theme]{Resource: theme, Errors: []error{}}
		themeErrors = append(themeErrors, re)

		err := c.Themes.Content[i+1].Decode(theme)
		if err != nil {
			for _, yerr := range err.(*yaml.TypeError).Errors {
				themeErrors[j].Errors = append(themeErrors[j].Errors, errors.New(yerr))
			}
			continue
		}

		themes = append(themes, *theme)
	}

	// Loop through themes and set default values
	for i := range themes {
		// TEXT
		if themes[i].Text.PrefixColors == nil {
			themes[i].Text.PrefixColors = DefaultText.PrefixColors
		}

		// if themes[i].Text.Prefix == "" {
		// 	themes[i].Text.Prefix = DefaultText.Prefix
		// }

		// TABLE
		if themes[i].Table.Style == "connected-light" {
			themes[i].Table.Box = StyleBoxLight
		} else {
			themes[i].Table.Style = "ascii"
			themes[i].Table.Box = StyleBoxASCII
		}

		if themes[i].Table.Prefix == "" {
			themes[i].Table.Prefix = DefaultTable.Prefix
		}

		if themes[i].Table.Options == nil {
			themes[i].Table.Options = DefaultTable.Options
		} else {
			if themes[i].Table.Options.DrawBorder == nil {
				themes[i].Table.Options.DrawBorder = DefaultTable.Options.DrawBorder
			}

			if themes[i].Table.Options.SeparateColumns == nil {
				themes[i].Table.Options.SeparateColumns = DefaultTable.Options.SeparateColumns
			}

			if themes[i].Table.Options.SeparateHeader == nil {
				themes[i].Table.Options.SeparateHeader = DefaultTable.Options.SeparateHeader
			}

			if themes[i].Table.Options.SeparateRows == nil {
				themes[i].Table.Options.SeparateRows = DefaultTable.Options.SeparateRows
			}

			if themes[i].Table.Options.SeparateFooter == nil {
				themes[i].Table.Options.SeparateFooter = DefaultTable.Options.SeparateFooter
			}
		}

		if themes[i].Table.Border == nil {
			themes[i].Table.Border = DefaultTable.Border
		} else {
			// Header
			if themes[i].Table.Border.Header == nil {
				themes[i].Table.Border.Header = DefaultTable.Border.Header
			} else {
				if themes[i].Table.Border.Header.Fg == nil {
					themes[i].Table.Border.Header.Fg = DefaultTable.Border.Header.Fg
				}
				if themes[i].Table.Border.Header.Bg == nil {
					themes[i].Table.Border.Header.Bg = DefaultTable.Border.Header.Bg
				}
				if themes[i].Table.Border.Header.Attr == nil {
					themes[i].Table.Border.Header.Attr = DefaultTable.Border.Header.Attr
				}
			}

			// Row
			if themes[i].Table.Border.Row == nil {
				themes[i].Table.Border.Row = DefaultTable.Border.Row
			} else {
				if themes[i].Table.Border.Row.Fg == nil {
					themes[i].Table.Border.Row.Fg = DefaultTable.Border.Row.Fg
				}
				if themes[i].Table.Border.Row.Bg == nil {
					themes[i].Table.Border.Row.Bg = DefaultTable.Border.Row.Bg
				}
				if themes[i].Table.Border.Row.Attr == nil {
					themes[i].Table.Border.Row.Attr = DefaultTable.Border.Row.Attr
				}
			}

			// RowAlternate
			if themes[i].Table.Border.RowAlternate == nil {
				themes[i].Table.Border.RowAlternate = DefaultTable.Border.RowAlternate
			} else {
				if themes[i].Table.Border.RowAlternate.Fg == nil {
					themes[i].Table.Border.RowAlternate.Fg = DefaultTable.Border.RowAlternate.Fg
				}
				if themes[i].Table.Border.RowAlternate.Bg == nil {
					themes[i].Table.Border.RowAlternate.Bg = DefaultTable.Border.RowAlternate.Bg
				}
				if themes[i].Table.Border.RowAlternate.Attr == nil {
					themes[i].Table.Border.RowAlternate.Attr = DefaultTable.Border.RowAlternate.Attr
				}
			}

			// Footer
			if themes[i].Table.Border.Footer == nil {
				themes[i].Table.Border.Footer = DefaultTable.Border.Footer
			} else {
				if themes[i].Table.Border.Footer.Fg == nil {
					themes[i].Table.Border.Footer.Fg = DefaultTable.Border.Footer.Fg
				}
				if themes[i].Table.Border.Footer.Bg == nil {
					themes[i].Table.Border.Footer.Bg = DefaultTable.Border.Footer.Bg
				}
				if themes[i].Table.Border.Footer.Attr == nil {
					themes[i].Table.Border.Footer.Attr = DefaultTable.Border.Footer.Attr
				}
			}
		}

		// Title
		if themes[i].Table.Title == nil {
			themes[i].Table.Title = DefaultTable.Title
		} else {
			// Header
			if themes[i].Table.Title == nil {
				themes[i].Table.Title = DefaultTable.Title
			} else {
				if themes[i].Table.Title.Fg == nil {
					themes[i].Table.Title.Fg = DefaultTable.Title.Fg
				}
				if themes[i].Table.Title.Bg == nil {
					themes[i].Table.Title.Bg = DefaultTable.Title.Bg
				}
				if themes[i].Table.Title.Align == nil {
					themes[i].Table.Title.Align = DefaultTable.Title.Align
				}
				if themes[i].Table.Title.Attr == nil {
					themes[i].Table.Title.Attr = DefaultTable.Title.Attr
				}
				if themes[i].Table.Title.Format == nil {
					themes[i].Table.Title.Format = DefaultTable.Title.Format
				}
			}
		}

		// Header
		if themes[i].Table.Header == nil {
			themes[i].Table.Header = DefaultTable.Header
		} else {
			// Header
			if themes[i].Table.Header == nil {
				themes[i].Table.Header = DefaultTable.Header
			} else {
				if themes[i].Table.Header.Fg == nil {
					themes[i].Table.Header.Fg = DefaultTable.Header.Fg
				}
				if themes[i].Table.Header.Bg == nil {
					themes[i].Table.Header.Bg = DefaultTable.Header.Bg
				}
				if themes[i].Table.Header.Align == nil {
					themes[i].Table.Header.Align = DefaultTable.Header.Align
				}
				if themes[i].Table.Header.Attr == nil {
					themes[i].Table.Header.Attr = DefaultTable.Header.Attr
				}
				if themes[i].Table.Header.Format == nil {
					themes[i].Table.Header.Format = DefaultTable.Header.Format
				}
			}
		}

		// Row
		if themes[i].Table.Row == nil {
			themes[i].Table.Row = DefaultTable.Row
		} else {
			// Row
			if themes[i].Table.Row == nil {
				themes[i].Table.Row = DefaultTable.Row
			} else {
				if themes[i].Table.Row.Fg == nil {
					themes[i].Table.Row.Fg = DefaultTable.Row.Fg
				}
				if themes[i].Table.Row.Bg == nil {
					themes[i].Table.Row.Bg = DefaultTable.Row.Bg
				}
				if themes[i].Table.Row.Align == nil {
					themes[i].Table.Row.Align = DefaultTable.Row.Align
				}
				if themes[i].Table.Row.Attr == nil {
					themes[i].Table.Row.Attr = DefaultTable.Row.Attr
				}
				if themes[i].Table.Row.Format == nil {
					themes[i].Table.Row.Format = DefaultTable.Row.Format
				}
			}
		}

		// Footer
		if themes[i].Table.Footer == nil {
			themes[i].Table.Footer = DefaultTable.Footer
		} else {
			// Footer
			if themes[i].Table.Footer == nil {
				themes[i].Table.Footer = DefaultTable.Footer
			} else {
				if themes[i].Table.Footer.Fg == nil {
					themes[i].Table.Footer.Fg = DefaultTable.Footer.Fg
				}
				if themes[i].Table.Footer.Bg == nil {
					themes[i].Table.Footer.Bg = DefaultTable.Footer.Bg
				}
				if themes[i].Table.Footer.Align == nil {
					themes[i].Table.Footer.Align = DefaultTable.Footer.Align
				}
				if themes[i].Table.Footer.Attr == nil {
					themes[i].Table.Footer.Attr = DefaultTable.Footer.Attr
				}
				if themes[i].Table.Footer.Format == nil {
					themes[i].Table.Footer.Format = DefaultTable.Footer.Format
				}
			}
		}
	}

	return themes, themeErrors
}

func (c *Config) GetTheme(name string) (*Theme, error) {
	for _, theme := range c.Themes {
		if name == theme.Name {
			return &theme, nil
		}
	}

	return nil, &core.ThemeNotFound{Name: name}
}

func (c *Config) GetThemeNames() []string {
	names := []string{}
	for _, theme := range c.Themes {
		names = append(names, theme.Name)
	}

	return names
}
