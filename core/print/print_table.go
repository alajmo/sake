package print

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/text"

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
	taskHeaders []string,
) {
	t := CreateTable(options, defaultHeaders, taskHeaders)

	// Headers
	var headers []any
	for _, h := range defaultHeaders {
		headers = append(headers, h)
	}
	for _, h := range taskHeaders {
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
		fmt.Println()
		fmt.Println(text.Bold.Sprintf(title))
	}

	RenderTable(t, options.Output)
}
