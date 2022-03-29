package dao

// TODO: This file needs refactoring

import (
	"errors"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"gopkg.in/yaml.v3"

	"github.com/alajmo/sake/core"
)

type TableOptions struct {
	DrawBorder      *bool `yaml:"draw_border"`
	SeparateColumns *bool `yaml:"separate_columns"`
	SeparateHeader  *bool `yaml:"separate_header"`
	SeparateRows    *bool `yaml:"separate_rows"`
	SeparateFooter  *bool `yaml:"separate_footer"`
}

type TableFormat struct {
	Header *string `yaml:"header"`
	Row    *string `yaml:"row"`
}

type ColorOptions struct {
	Fg    *string `yaml:"fg"`
	Bg    *string `yaml:"bg"`
	Align *string `yaml:"align"`
	Attr  *string `yaml:"attr"`
}

type BorderColors struct {
	Header       *ColorOptions `yaml:"header"`
	Row          *ColorOptions `yaml:"row"`
	RowAlternate *ColorOptions `yaml:"row_alt"`
	Footer       *ColorOptions `yaml:"footer"`
}

type CellColors struct {
	Server *ColorOptions `yaml:"server"`
	Tag    *ColorOptions `yaml:"tag"`
	Desc   *ColorOptions `yaml:"desc"`
	Host   *ColorOptions `yaml:"host"`
	User   *ColorOptions `yaml:"user"`
	Port   *ColorOptions `yaml:"port"`
	Local  *ColorOptions `yaml:"local"`
	Task   *ColorOptions `yaml:"task"`
	Output *ColorOptions `yaml:"output"`
}

type TableColor struct {
	Border *BorderColors `yaml:"border"`
	Header *CellColors   `yaml:"header"`
	Row    *CellColors   `yaml:"row"`
}

type Table struct {
	// Stylable via YAML
	Name    string        `yaml:"name"`
	Style   string        `yaml:"style"`
	Color   *TableColor   `yaml:"color"`
	Format  *TableFormat  `yaml:"format"`
	Options *TableOptions `yaml:"options"`

	// Not stylable via YAML
	Box table.BoxStyle `yaml:"-"`
}

type Text struct {
	Prefix       bool     `yaml:"prefix"`
	PrefixColors []string `yaml:"prefix_colors"`
	Header       bool     `yaml:"header"`
	HeaderChar   string   `yaml:"header_char"`
	HeaderPrefix string   `yaml:"header_prefix"`
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
	Prefix:       true,
	PrefixColors: []string{"green", "blue", "red", "yellow", "magenta", "cyan"},
	Header:       true,
	HeaderPrefix: "TASK",
	HeaderChar:   "*",
}

var DefaultTable = Table{
	Style: "default",
	Box:   StyleBoxASCII,

	Format: &TableFormat{
		Header: core.Ptr("title"),
		Row:    core.Ptr(""),
	},

	Options: &TableOptions{
		DrawBorder:      core.Ptr(false),
		SeparateColumns: core.Ptr(true),
		SeparateHeader:  core.Ptr(true),
		SeparateRows:    core.Ptr(false),
		SeparateFooter:  core.Ptr(false),
	},

	Color: &TableColor{
		Border: &BorderColors{
			Header: &ColorOptions{
				Fg:   core.Ptr(""),
				Bg:   core.Ptr(""),
				Attr: core.Ptr("faint"),
			},

			Row: &ColorOptions{
				Fg:   core.Ptr(""),
				Bg:   core.Ptr(""),
				Attr: core.Ptr("faint"),
			},

			RowAlternate: &ColorOptions{
				Fg:   core.Ptr(""),
				Bg:   core.Ptr(""),
				Attr: core.Ptr("faint"),
			},

			Footer: &ColorOptions{
				Fg:   core.Ptr(""),
				Bg:   core.Ptr(""),
				Attr: core.Ptr("faint"),
			},
		},

		Header: &CellColors{
			Server: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			Tag: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			Desc: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			Host: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			User: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			Local: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			Port: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			Task: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			Output: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
		},

		Row: &CellColors{
			Server: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			Tag: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			Desc: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			Host: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			User: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			Local: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			Port: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			Task: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			Output: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
		},
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

		// TABLE
		if themes[i].Table.Style == "connected-light" {
			themes[i].Table.Box = StyleBoxLight
		} else {
			themes[i].Table.Style = "ascii"
			themes[i].Table.Box = StyleBoxASCII
		}

		// Format
		if themes[i].Table.Format == nil {
			themes[i].Table.Format = DefaultTable.Format
		} else {
			if themes[i].Table.Format.Header == nil {
				themes[i].Table.Format.Header = DefaultTable.Format.Header
			}

			if themes[i].Table.Format.Row == nil {
				themes[i].Table.Format.Row = DefaultTable.Format.Row
			}
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

		// Colors
		if themes[i].Table.Color == nil {
			themes[i].Table.Color = DefaultTable.Color
		} else {
			// Border
			if themes[i].Table.Color.Border == nil {
				themes[i].Table.Color.Border = DefaultTable.Color.Border
			} else {
				// Header
				if themes[i].Table.Color.Border.Header == nil {
					themes[i].Table.Color.Border.Header = DefaultTable.Color.Border.Header
				} else {
					if themes[i].Table.Color.Border.Header.Fg == nil {
						themes[i].Table.Color.Border.Header.Fg = DefaultTable.Color.Border.Header.Fg
					}
					if themes[i].Table.Color.Border.Header.Bg == nil {
						themes[i].Table.Color.Border.Header.Bg = DefaultTable.Color.Border.Header.Bg
					}
					if themes[i].Table.Color.Border.Header.Attr == nil {
						themes[i].Table.Color.Border.Header.Attr = DefaultTable.Color.Border.Header.Attr
					}
				}

				// Row
				if themes[i].Table.Color.Border.Row == nil {
					themes[i].Table.Color.Border.Row = DefaultTable.Color.Border.Row
				} else {
					if themes[i].Table.Color.Border.Row.Fg == nil {
						themes[i].Table.Color.Border.Row.Fg = DefaultTable.Color.Border.Row.Fg
					}
					if themes[i].Table.Color.Border.Row.Bg == nil {
						themes[i].Table.Color.Border.Row.Bg = DefaultTable.Color.Border.Row.Bg
					}
					if themes[i].Table.Color.Border.Row.Attr == nil {
						themes[i].Table.Color.Border.Row.Attr = DefaultTable.Color.Border.Row.Attr
					}
				}

				// RowAlternate
				if themes[i].Table.Color.Border.RowAlternate == nil {
					themes[i].Table.Color.Border.RowAlternate = DefaultTable.Color.Border.RowAlternate
				} else {
					if themes[i].Table.Color.Border.RowAlternate.Fg == nil {
						themes[i].Table.Color.Border.RowAlternate.Fg = DefaultTable.Color.Border.RowAlternate.Fg
					}
					if themes[i].Table.Color.Border.RowAlternate.Bg == nil {
						themes[i].Table.Color.Border.RowAlternate.Bg = DefaultTable.Color.Border.RowAlternate.Bg
					}
					if themes[i].Table.Color.Border.RowAlternate.Attr == nil {
						themes[i].Table.Color.Border.RowAlternate.Attr = DefaultTable.Color.Border.RowAlternate.Attr
					}
				}

				// Footer
				if themes[i].Table.Color.Border.Footer == nil {
					themes[i].Table.Color.Border.Footer = DefaultTable.Color.Border.Footer
				} else {
					if themes[i].Table.Color.Border.Footer.Fg == nil {
						themes[i].Table.Color.Border.Footer.Fg = DefaultTable.Color.Border.Footer.Fg
					}
					if themes[i].Table.Color.Border.Footer.Bg == nil {
						themes[i].Table.Color.Border.Footer.Bg = DefaultTable.Color.Border.Footer.Bg
					}
					if themes[i].Table.Color.Border.Footer.Attr == nil {
						themes[i].Table.Color.Border.Footer.Attr = DefaultTable.Color.Border.Footer.Attr
					}
				}

			}

			// Header
			if themes[i].Table.Color.Header == nil {
				themes[i].Table.Color.Header = DefaultTable.Color.Header
			} else {
				// Server
				if themes[i].Table.Color.Header.Server == nil {
					themes[i].Table.Color.Header.Server = DefaultTable.Color.Header.Server
				} else {
					if themes[i].Table.Color.Header.Server.Fg == nil {
						themes[i].Table.Color.Header.Server.Fg = DefaultTable.Color.Header.Server.Fg
					}
					if themes[i].Table.Color.Header.Server.Bg == nil {
						themes[i].Table.Color.Header.Server.Bg = DefaultTable.Color.Header.Server.Bg
					}
					if themes[i].Table.Color.Header.Server.Align == nil {
						themes[i].Table.Color.Header.Server.Align = DefaultTable.Color.Header.Server.Align
					}
					if themes[i].Table.Color.Header.Server.Attr == nil {
						themes[i].Table.Color.Header.Server.Attr = DefaultTable.Color.Header.Server.Attr
					}
				}

				// Tag
				if themes[i].Table.Color.Header.Tag == nil {
					themes[i].Table.Color.Header.Tag = DefaultTable.Color.Header.Tag
				} else {
					if themes[i].Table.Color.Header.Tag.Fg == nil {
						themes[i].Table.Color.Header.Tag.Fg = DefaultTable.Color.Header.Tag.Fg
					}
					if themes[i].Table.Color.Header.Tag.Bg == nil {
						themes[i].Table.Color.Header.Tag.Bg = DefaultTable.Color.Header.Tag.Bg
					}
					if themes[i].Table.Color.Header.Tag.Align == nil {
						themes[i].Table.Color.Header.Tag.Align = DefaultTable.Color.Header.Tag.Align
					}
					if themes[i].Table.Color.Header.Tag.Attr == nil {
						themes[i].Table.Color.Header.Tag.Attr = DefaultTable.Color.Header.Tag.Attr
					}
				}

				// Desc
				if themes[i].Table.Color.Header.Desc == nil {
					themes[i].Table.Color.Header.Desc = DefaultTable.Color.Header.Desc
				} else {
					if themes[i].Table.Color.Header.Desc.Fg == nil {
						themes[i].Table.Color.Header.Desc.Fg = DefaultTable.Color.Header.Desc.Fg
					}
					if themes[i].Table.Color.Header.Desc.Bg == nil {
						themes[i].Table.Color.Header.Desc.Bg = DefaultTable.Color.Header.Desc.Bg
					}
					if themes[i].Table.Color.Header.Desc.Align == nil {
						themes[i].Table.Color.Header.Desc.Align = DefaultTable.Color.Header.Desc.Align
					}
					if themes[i].Table.Color.Header.Desc.Attr == nil {
						themes[i].Table.Color.Header.Desc.Attr = DefaultTable.Color.Header.Desc.Attr
					}
				}

				// Host
				if themes[i].Table.Color.Header.Host == nil {
					themes[i].Table.Color.Header.Host = DefaultTable.Color.Header.Host
				} else {
					if themes[i].Table.Color.Header.Host.Fg == nil {
						themes[i].Table.Color.Header.Host.Fg = DefaultTable.Color.Header.Host.Fg
					}
					if themes[i].Table.Color.Header.Host.Bg == nil {
						themes[i].Table.Color.Header.Host.Bg = DefaultTable.Color.Header.Host.Bg
					}
					if themes[i].Table.Color.Header.Host.Align == nil {
						themes[i].Table.Color.Header.Host.Align = DefaultTable.Color.Header.Host.Align
					}
					if themes[i].Table.Color.Header.Host.Attr == nil {
						themes[i].Table.Color.Header.Host.Attr = DefaultTable.Color.Header.Host.Attr
					}
				}

				// User
				if themes[i].Table.Color.Header.User == nil {
					themes[i].Table.Color.Header.User = DefaultTable.Color.Header.User
				} else {
					if themes[i].Table.Color.Header.User.Fg == nil {
						themes[i].Table.Color.Header.User.Fg = DefaultTable.Color.Header.User.Fg
					}
					if themes[i].Table.Color.Header.User.Bg == nil {
						themes[i].Table.Color.Header.User.Bg = DefaultTable.Color.Header.User.Bg
					}
					if themes[i].Table.Color.Header.User.Align == nil {
						themes[i].Table.Color.Header.User.Align = DefaultTable.Color.Header.User.Align
					}
					if themes[i].Table.Color.Header.User.Attr == nil {
						themes[i].Table.Color.Header.User.Attr = DefaultTable.Color.Header.User.Attr
					}
				}

				// Port
				if themes[i].Table.Color.Header.Port == nil {
					themes[i].Table.Color.Header.Port = DefaultTable.Color.Header.Port
				} else {
					if themes[i].Table.Color.Header.Port.Fg == nil {
						themes[i].Table.Color.Header.Port.Fg = DefaultTable.Color.Header.Port.Fg
					}
					if themes[i].Table.Color.Header.Port.Bg == nil {
						themes[i].Table.Color.Header.Port.Bg = DefaultTable.Color.Header.Port.Bg
					}
					if themes[i].Table.Color.Header.Port.Align == nil {
						themes[i].Table.Color.Header.Port.Align = DefaultTable.Color.Header.Port.Align
					}
					if themes[i].Table.Color.Header.Port.Attr == nil {
						themes[i].Table.Color.Header.Port.Attr = DefaultTable.Color.Header.Port.Attr
					}
				}

				// Local
				if themes[i].Table.Color.Header.Local == nil {
					themes[i].Table.Color.Header.Local = DefaultTable.Color.Header.Local
				} else {
					if themes[i].Table.Color.Header.Local.Fg == nil {
						themes[i].Table.Color.Header.Local.Fg = DefaultTable.Color.Header.Local.Fg
					}
					if themes[i].Table.Color.Header.Local.Bg == nil {
						themes[i].Table.Color.Header.Local.Bg = DefaultTable.Color.Header.Local.Bg
					}
					if themes[i].Table.Color.Header.Local.Align == nil {
						themes[i].Table.Color.Header.Local.Align = DefaultTable.Color.Header.Local.Align
					}
					if themes[i].Table.Color.Header.Local.Attr == nil {
						themes[i].Table.Color.Header.Local.Attr = DefaultTable.Color.Header.Local.Attr
					}
				}

				// Task
				if themes[i].Table.Color.Header.Task == nil {
					themes[i].Table.Color.Header.Task = DefaultTable.Color.Header.Task
				} else {
					if themes[i].Table.Color.Header.Task.Fg == nil {
						themes[i].Table.Color.Header.Task.Fg = DefaultTable.Color.Header.Task.Fg
					}
					if themes[i].Table.Color.Header.Task.Bg == nil {
						themes[i].Table.Color.Header.Task.Bg = DefaultTable.Color.Header.Task.Bg
					}
					if themes[i].Table.Color.Header.Task.Align == nil {
						themes[i].Table.Color.Header.Task.Align = DefaultTable.Color.Header.Task.Align
					}
					if themes[i].Table.Color.Header.Task.Attr == nil {
						themes[i].Table.Color.Header.Task.Attr = DefaultTable.Color.Header.Task.Attr
					}
				}

				// Output
				if themes[i].Table.Color.Header.Output == nil {
					themes[i].Table.Color.Header.Output = DefaultTable.Color.Header.Output
				} else {
					if themes[i].Table.Color.Header.Output.Fg == nil {
						themes[i].Table.Color.Header.Output.Fg = DefaultTable.Color.Header.Output.Fg
					}
					if themes[i].Table.Color.Header.Output.Bg == nil {
						themes[i].Table.Color.Header.Output.Bg = DefaultTable.Color.Header.Output.Bg
					}
					if themes[i].Table.Color.Header.Output.Align == nil {
						themes[i].Table.Color.Header.Output.Align = DefaultTable.Color.Header.Output.Align
					}
					if themes[i].Table.Color.Header.Output.Attr == nil {
						themes[i].Table.Color.Header.Output.Attr = DefaultTable.Color.Header.Output.Attr
					}
				}
			}

			// Row
			if themes[i].Table.Color.Row == nil {
				themes[i].Table.Color.Row = DefaultTable.Color.Row
			} else {
				// Server
				if themes[i].Table.Color.Row.Server == nil {
					themes[i].Table.Color.Row.Server = DefaultTable.Color.Row.Server
				} else {
					if themes[i].Table.Color.Row.Server.Fg == nil {
						themes[i].Table.Color.Row.Server.Fg = DefaultTable.Color.Row.Server.Fg
					}
					if themes[i].Table.Color.Row.Server.Bg == nil {
						themes[i].Table.Color.Row.Server.Bg = DefaultTable.Color.Row.Server.Bg
					}
					if themes[i].Table.Color.Row.Server.Align == nil {
						themes[i].Table.Color.Row.Server.Align = DefaultTable.Color.Row.Server.Align
					}
					if themes[i].Table.Color.Row.Server.Attr == nil {
						themes[i].Table.Color.Row.Server.Attr = DefaultTable.Color.Row.Server.Attr
					}
				}

				// Tag
				if themes[i].Table.Color.Row.Tag == nil {
					themes[i].Table.Color.Row.Tag = DefaultTable.Color.Row.Tag
				} else {
					if themes[i].Table.Color.Row.Tag.Fg == nil {
						themes[i].Table.Color.Row.Tag.Fg = DefaultTable.Color.Row.Tag.Fg
					}
					if themes[i].Table.Color.Row.Tag.Bg == nil {
						themes[i].Table.Color.Row.Tag.Bg = DefaultTable.Color.Row.Tag.Bg
					}
					if themes[i].Table.Color.Row.Tag.Align == nil {
						themes[i].Table.Color.Row.Tag.Align = DefaultTable.Color.Row.Tag.Align
					}
					if themes[i].Table.Color.Row.Tag.Attr == nil {
						themes[i].Table.Color.Row.Tag.Attr = DefaultTable.Color.Row.Tag.Attr
					}
				}

				// Desc
				if themes[i].Table.Color.Row.Desc == nil {
					themes[i].Table.Color.Row.Desc = DefaultTable.Color.Row.Desc
				} else {
					if themes[i].Table.Color.Row.Desc.Fg == nil {
						themes[i].Table.Color.Row.Desc.Fg = DefaultTable.Color.Row.Desc.Fg
					}
					if themes[i].Table.Color.Row.Desc.Bg == nil {
						themes[i].Table.Color.Row.Desc.Bg = DefaultTable.Color.Row.Desc.Bg
					}
					if themes[i].Table.Color.Row.Desc.Align == nil {
						themes[i].Table.Color.Row.Desc.Align = DefaultTable.Color.Row.Desc.Align
					}
					if themes[i].Table.Color.Row.Desc.Attr == nil {
						themes[i].Table.Color.Row.Desc.Attr = DefaultTable.Color.Row.Desc.Attr
					}
				}

				// Host
				if themes[i].Table.Color.Row.Host == nil {
					themes[i].Table.Color.Row.Host = DefaultTable.Color.Row.Host
				} else {
					if themes[i].Table.Color.Row.Host.Fg == nil {
						themes[i].Table.Color.Row.Host.Fg = DefaultTable.Color.Row.Host.Fg
					}
					if themes[i].Table.Color.Row.Host.Bg == nil {
						themes[i].Table.Color.Row.Host.Bg = DefaultTable.Color.Row.Host.Bg
					}
					if themes[i].Table.Color.Row.Host.Align == nil {
						themes[i].Table.Color.Row.Host.Align = DefaultTable.Color.Row.Host.Align
					}
					if themes[i].Table.Color.Row.Host.Attr == nil {
						themes[i].Table.Color.Row.Host.Attr = DefaultTable.Color.Row.Host.Attr
					}
				}

				// User
				if themes[i].Table.Color.Row.User == nil {
					themes[i].Table.Color.Row.User = DefaultTable.Color.Row.User
				} else {
					if themes[i].Table.Color.Row.User.Fg == nil {
						themes[i].Table.Color.Row.User.Fg = DefaultTable.Color.Row.User.Fg
					}
					if themes[i].Table.Color.Row.User.Bg == nil {
						themes[i].Table.Color.Row.User.Bg = DefaultTable.Color.Row.User.Bg
					}
					if themes[i].Table.Color.Row.User.Align == nil {
						themes[i].Table.Color.Row.User.Align = DefaultTable.Color.Row.User.Align
					}
					if themes[i].Table.Color.Row.User.Attr == nil {
						themes[i].Table.Color.Row.User.Attr = DefaultTable.Color.Row.User.Attr
					}
				}

				// Port
				if themes[i].Table.Color.Row.Port == nil {
					themes[i].Table.Color.Row.Port = DefaultTable.Color.Row.Port
				} else {
					if themes[i].Table.Color.Row.Port.Fg == nil {
						themes[i].Table.Color.Row.Port.Fg = DefaultTable.Color.Row.Port.Fg
					}
					if themes[i].Table.Color.Row.Port.Bg == nil {
						themes[i].Table.Color.Row.Port.Bg = DefaultTable.Color.Row.Port.Bg
					}
					if themes[i].Table.Color.Row.Port.Align == nil {
						themes[i].Table.Color.Row.Port.Align = DefaultTable.Color.Row.Port.Align
					}
					if themes[i].Table.Color.Row.Port.Attr == nil {
						themes[i].Table.Color.Row.Port.Attr = DefaultTable.Color.Row.Port.Attr
					}
				}

				// Local
				if themes[i].Table.Color.Row.Local == nil {
					themes[i].Table.Color.Row.Local = DefaultTable.Color.Row.Local
				} else {
					if themes[i].Table.Color.Row.Local.Fg == nil {
						themes[i].Table.Color.Row.Local.Fg = DefaultTable.Color.Row.Local.Fg
					}
					if themes[i].Table.Color.Row.Local.Bg == nil {
						themes[i].Table.Color.Row.Local.Bg = DefaultTable.Color.Row.Local.Bg
					}
					if themes[i].Table.Color.Row.Local.Align == nil {
						themes[i].Table.Color.Row.Local.Align = DefaultTable.Color.Row.Local.Align
					}
					if themes[i].Table.Color.Row.Local.Attr == nil {
						themes[i].Table.Color.Row.Local.Attr = DefaultTable.Color.Row.Local.Attr
					}
				}

				// Task
				if themes[i].Table.Color.Row.Task == nil {
					themes[i].Table.Color.Row.Task = DefaultTable.Color.Row.Task
				} else {
					if themes[i].Table.Color.Row.Task.Fg == nil {
						themes[i].Table.Color.Row.Task.Fg = DefaultTable.Color.Row.Task.Fg
					}
					if themes[i].Table.Color.Row.Task.Bg == nil {
						themes[i].Table.Color.Row.Task.Bg = DefaultTable.Color.Row.Task.Bg
					}
					if themes[i].Table.Color.Row.Task.Align == nil {
						themes[i].Table.Color.Row.Task.Align = DefaultTable.Color.Row.Task.Align
					}
					if themes[i].Table.Color.Row.Task.Attr == nil {
						themes[i].Table.Color.Row.Task.Attr = DefaultTable.Color.Row.Task.Attr
					}
				}

				// Output
				if themes[i].Table.Color.Row.Output == nil {
					themes[i].Table.Color.Row.Output = DefaultTable.Color.Row.Output
				} else {
					if themes[i].Table.Color.Row.Output.Fg == nil {
						themes[i].Table.Color.Row.Output.Fg = DefaultTable.Color.Row.Output.Fg
					}
					if themes[i].Table.Color.Row.Output.Bg == nil {
						themes[i].Table.Color.Row.Output.Bg = DefaultTable.Color.Row.Output.Bg
					}
					if themes[i].Table.Color.Row.Output.Align == nil {
						themes[i].Table.Color.Row.Output.Align = DefaultTable.Color.Row.Output.Align
					}
					if themes[i].Table.Color.Row.Output.Attr == nil {
						themes[i].Table.Color.Row.Output.Attr = DefaultTable.Color.Row.Output.Attr
					}
				}
			}
		}
	}

	return themes, themeErrors
}

func (c Config) GetTheme(name string) (*Theme, error) {
	for _, theme := range c.Themes {
		if name == theme.Name {
			return &theme, nil
		}
	}

	return nil, &core.ThemeNotFound{Name: name}
}

func (c Config) GetThemeNames() []string {
	names := []string{}
	for _, theme := range c.Themes {
		names = append(names, theme.Name)
	}

	return names
}
