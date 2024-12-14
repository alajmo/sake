package print

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
)

type PrintTableOptions struct {
	Title            string
	Output           string
	Theme            dao.Theme
	Resource         string
	OmitEmptyRows    bool
	OmitEmptyColumns bool
}

func PrintTable[T dao.Items](
	data []T,
	options PrintTableOptions,
	headers []string,
	footers []string,
	padTop bool,
	padBottom bool,
) error {

	switch options.Output {
	case "table", "table-1":
		return table1(data, options, headers, footers, padTop, padBottom)
	case "table-2":
		return table2(data, options, headers, padTop, padBottom)
	case "table-3":
		return table3(data, options, headers, padTop, padBottom)
	case "table-4":
		return table4(data, options, headers, padTop, padBottom)
	case "csv":
		return printCSV(data, options, headers)
	case "json":
		return printJSON(data, headers)
	default:
		return table1(data, options, headers, footers, padTop, padBottom)
	}
}

/*
1 table, task names in 1st row

	Server | Host               | Hostname     | OS           | Kernel

--------+--------------------+--------------+--------------+--------

	ip6-1  | 2001:3984:3989::10 | 31cb8139dffd | Ubuntu 22.04 | 5.18.0

--------+--------------------+--------------+--------------+--------

	ip6-2  | 2001:3984:3989::11 | 54666c1891fb | Ubuntu 22.04 | 5.18.0
*/
func table1[T dao.Items](
	data []T,
	options PrintTableOptions,
	headers []string,
	footers []string,
	padTop bool,
	padBottom bool,
) error {
	t := CreateTable(options, headers)

	// Headers
	var tableHeaders table.Row
	for _, h := range headers {
		tableHeaders = append(tableHeaders, h)
	}
	t.AppendHeader(tableHeaders)

	// Rows
	for _, item := range data {
		var row []any
		for i, h := range headers {
			value := item.GetValue(fmt.Sprintf("%v", h), i)
			row = append(row, value)
		}

		if options.OmitEmptyRows {
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

	var tableFooter table.Row
	for _, h := range footers {
		tableFooter = append(tableFooter, h)
	}
	t.AppendFooter(tableFooter)

	if options.Title != "" {
		t.SetTitle(options.Title)
	}

	RenderTable(t, options.Output, padTop, padBottom)

	return nil
}

/*
1 table, task names in 1st column

	Task     | Ip6-1              | Ip6-2

----------+--------------------+--------------------

	Host     | 2001:3984:3989::10 | 2001:3984:3989::11

----------+--------------------+--------------------

	Hostname | 31cb8139dffd       | 54666c1891fb

----------+--------------------+--------------------

	OS       | Ubuntu 22.04       | Ubuntu 22.04

----------+--------------------+--------------------

	Kernel   | 5.18.0             | 5.18.0
*/
func table2[T dao.Items](
	data []T,
	options PrintTableOptions,
	headers []string,
	padTop bool,
	padBottom bool,
) error {
	tableHeaders := table.Row{options.Resource}
	rh := []string{options.Resource}
	for _, h := range data {
		value := h.GetValue(fmt.Sprintf("%v", h), 0)
		rh = append(rh, value)
		tableHeaders = append(tableHeaders, value)
	}
	t := CreateTable(options, rh)

	t.AppendHeader(tableHeaders)
	for i, header := range headers[1:] {
		var row table.Row
		row = append(row, header)
		for _, h := range data {
			value := h.GetValue(fmt.Sprintf("%v", h), i+1)
			row = append(row, value)
		}

		if options.OmitEmptyRows {
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

	RenderTable(t, options.Output, padTop, padBottom)

	return nil
}

/*
1 table per server, task names in 1st row

	                         ip6-1

	 Host               | Hostname     | OS           | Kernel
	--------------------+--------------+--------------+--------
	 2001:3984:3989::10 | 31cb8139dffd | Ubuntu 22.04 | 5.18.0



	                         ip6-2

	 Host               | Hostname     | OS           | Kernel
	--------------------+--------------+--------------+--------
	 2001:3984:3989::11 | 54666c1891fb | Ubuntu 22.04 | 5.18.0
*/
func table3[T dao.Items](
	data []T,
	options PrintTableOptions,
	headers []string,
	padTop bool,
	padBottom bool,
) error {
	var tableHeaders table.Row
	for _, h := range headers[1:] {
		tableHeaders = append(tableHeaders, h)
	}

	for _, s := range data {
		t := CreateTable(options, headers)

		t.AppendHeader(tableHeaders)

		title := fmt.Sprintf("\n%s\n", s.GetValue(fmt.Sprintf("%v", s), 0))
		t.SetTitle(title)

		var row table.Row
		for i, h := range headers[1:] {
			value := s.GetValue(fmt.Sprintf("%v", h), i+1)
			row = append(row, value)
		}
		t.AppendRow(row)

		if options.OmitEmptyRows {
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

		RenderTable(t, options.Output, padTop, padBottom)
	}

	return nil
}

/*
1 table per server, task names in 1st column

	Task     | Ip6-1

----------+--------------------

	Host     | 2001:3984:3989::10

----------+--------------------

	Hostname | 31cb8139dffd

----------+--------------------

	OS       | Ubuntu 22.04

----------+--------------------

	Kernel   | 5.18.0

	Task     | Ip6-2

----------+--------------------

	Host     | 2001:3984:3989::11

----------+--------------------

	Hostname | 54666c1891fb

----------+--------------------

	OS       | Ubuntu 22.04

----------+--------------------

	Kernel   | 5.18.0
*/
func table4[T dao.Items](
	data []T,
	options PrintTableOptions,
	headers []string,
	padTop bool,
	padBottom bool,
) error {
	for _, s := range data {
		val := s.GetValue(fmt.Sprintf("%v", s), 0)
		t := CreateTable(options, []string{options.Resource, val})

		tableHeaders := table.Row{options.Resource, val}
		t.AppendHeader(tableHeaders)

		for i, h := range headers[1:] {
			var row []any
			value := s.GetValue(fmt.Sprintf("%v", h), i+1)
			row = append(row, h)
			row = append(row, value)

			if options.OmitEmptyRows {
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

		RenderTable(t, options.Output, padTop, padBottom)
	}

	return nil
}

/*
server,host
ip6-1,2001:3984:3989::10
ip6-2,2001:3984:3989::11
*/
func printCSV[T dao.Items](data []T, options PrintTableOptions, headers []string) error {
	t := CreateTable(options, headers)

	// Headers
	var tableHeaders table.Row
	for _, h := range headers {
		tableHeaders = append(tableHeaders, h)
	}
	t.AppendHeader(tableHeaders)

	// Rows
	for _, item := range data {
		var row []any
		for i, h := range headers {
			value := item.GetValue(fmt.Sprintf("%v", h), i)
			switch h {
			case "tags":
				t := strings.Split(value, "\n")
				v := strings.Join(t, ",")
				row = append(row, v)
			default:
				row = append(row, value)
			}
		}

		if options.OmitEmptyRows {
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

	if options.Title != "" {
		t.SetTitle(options.Title)
	}

	RenderTable(t, options.Output, true, true)

	return nil
}

/*
[

	{
	  "ping": "pong"
	  "server": "list-0",
	},
	{
	  "ping": "pong"
	  "server": "list-1",
	}

]
*/
func printJSON[T dao.Items](data []T, headers []string) error {
	var out []map[string]any
	for _, v := range data {
		m := make(map[string]any)
		for j, k := range headers {
			value := v.GetValue(k, j)
			switch k {
			case "servers":
				s := core.SplitString(value, "\n")
				m[k] = s
			case "tags":
				t := core.SplitString(value, "\n")
				m[k] = t
			default:
				m[k] = value
			}
		}
		out = append(out, m)
	}

	a, err := json.Marshal(out)
	if err != nil {
		return err
	}

	fmt.Println(string(a))

	return nil
}
