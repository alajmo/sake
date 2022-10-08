package print

import (
	"fmt"

	"github.com/alajmo/sake/core/dao"
)

type Items interface {
	GetValue(string, int) string
}

type PrintTableOptions struct {
	Output               string
	Theme                dao.Theme
	OmitEmpty            bool
	SuppressEmptyColumns bool
}

func PrintTable[T Items](
	title string,

	data []T,
	options PrintTableOptions,
	defaultHeaders []string,
	restHeaders []string,
) {
	switch options.Output {
	case "table-2":
		table2(data, options, restHeaders, title)
	case "table-3":
		table3(data, options, defaultHeaders, restHeaders, title)
	case "table-4":
		table4(data, options, defaultHeaders, restHeaders, title)
	default:
		table1(data, options, defaultHeaders, restHeaders, title)
	}
}

func table1[T Items](data []T, options PrintTableOptions, defaultHeaders []string, restHeaders []string, title string) {
	t := CreateTable(options, defaultHeaders, restHeaders)

	// Headers
	var headers []any
	for _, h := range defaultHeaders {
		headers = append(headers, h)
	}
	for _, h := range restHeaders {
		headers = append(headers, h)
	}

	t.AppendHeader(headers)

	// Rows
	for _, item := range data {
		var row []any
		for i, h := range headers {
			value := item.GetValue(fmt.Sprintf("%v", h), i)
			row = append(row, value)
		}

		if options.OmitEmpty {
			empty := true
			for _, v := range row[1:] {
				if v != "" {
					empty = false
				}
			}

			if empty {
				continue
			}
		}

		t.AppendRow(row)
	}

	if title != "" {
		t.SetTitle((title))
	}

	RenderTable(t, options.Output)
}

func table2[T Items](data []T, options PrintTableOptions, restHeaders []string, title string) {
	var headers []any
	var dh []string
	var rh []string
	dh = []string{"task"}
	for _, h := range dh {
		headers = append(headers, h)
	}
	rh = []string{}
	for _, h := range data {
		value := h.GetValue(fmt.Sprintf("%v", h), 0)
		rh = append(rh, value)
		headers = append(headers, value)
	}
	t := CreateTable(options, dh, rh)

	t.AppendHeader(headers)

	for i, task := range restHeaders {
		var row []any
		row = append(row, task)
		for _, h := range data {
			value := h.GetValue(fmt.Sprintf("%v", h), i+1)
			row = append(row, value)
		}

		if options.OmitEmpty {
			empty := true
			for _, v := range row[1:] {
				if v != "" {
					empty = false
				}
			}
			if empty {
				continue
			}
		}

		t.AppendRow(row)
	}

	RenderTable(t, options.Output)
}

func table3[T Items](data []T, options PrintTableOptions, defaultHeaders []string, restHeaders []string, title string) {
	var headers []any
	for _, h := range restHeaders {
		headers = append(headers, h)
	}

	for _, s := range data {
		t := CreateTable(options, []string{}, restHeaders)

		t.AppendHeader(headers)

		title = fmt.Sprintf("\n%s\n", s.GetValue(fmt.Sprintf("%v", s), 0))
		t.SetTitle((title))

		var row []any
		for i, h := range restHeaders {
			value := s.GetValue(fmt.Sprintf("%v", h), i+1)
			row = append(row, value)
		}
		t.AppendRow(row)

		if options.OmitEmpty {
			empty := true
			for _, v := range row {
				if v != "" {
					empty = false
				}
			}
			if empty {
				continue
			}
		}

		RenderTable(t, options.Output)
	}
}

func table4[T Items](data []T, options PrintTableOptions, defaultHeaders []string, restHeaders []string, title string) {
	for _, s := range data {
		headers := []any{"task", s.GetValue(fmt.Sprintf("%v", s), 0)}

		t := CreateTable(options, []string{"task"}, restHeaders)

		t.AppendHeader(headers)

		for i, h := range restHeaders {
			var row []any
			value := s.GetValue(fmt.Sprintf("%v", h), i+1)
			row = append(row, h)
			row = append(row, value)

			if options.OmitEmpty {
				empty := true
				for _, v := range row[1:] {
					if v != "" {
						empty = false
					}
				}
				if empty {
					continue
				}
			}

			t.AppendRow(row)
		}

		RenderTable(t, options.Output)
	}
}
