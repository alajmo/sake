package print

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
)

// TODO: Support csv,html,json,markdown
func PrintReport(
	theme *dao.Theme,
	reportData dao.ReportData,
	spec dao.Spec,
) error {
	reportTheme := dao.DEFAULT_THEME
	reportTheme.Table.Options.DrawBorder = core.Ptr(false)
	reportTheme.Table.Options.SeparateColumns = core.Ptr(false)
	reportTheme.Table.Options.SeparateRows = core.Ptr(false)
	reportTheme.Table.Options.SeparateHeader = core.Ptr(true)
	reportTheme.Table.Options.SeparateFooter = core.Ptr(false)
	options := PrintTableOptions{
		Theme:            reportTheme,
		Output:           "table",
		OmitEmptyRows:    false,
		OmitEmptyColumns: false,
	}

	summaryTheme := dao.DEFAULT_THEME
	summaryTheme.Table.Options.DrawBorder = core.Ptr(false)
	summaryTheme.Table.Options.SeparateColumns = core.Ptr(false)
	summaryTheme.Table.Options.SeparateRows = core.Ptr(false)
	summaryTheme.Table.Options.SeparateHeader = core.Ptr(true)
	summaryTheme.Table.Options.SeparateFooter = core.Ptr(false)
	summaryOptions := PrintTableOptions{
		Theme:            summaryTheme,
		Output:           "table",
		OmitEmptyRows:    false,
		OmitEmptyColumns: false,
	}

	if core.StringInSlice("all", spec.Report) {
		printRecapHeader("RETURN CODES ", theme.Text.HeaderFiller)
		err := PrintExitReport(&reportTheme, options, reportData)
		if err != nil {
			return err
		}
		printRecapHeader("TASK STATUS ", theme.Text.HeaderFiller)
		err = PrintTaskReport(&reportTheme, options, reportData)
		if err != nil {
			return err
		}
		printRecapHeader("TIME ", theme.Text.HeaderFiller)
		err = PrintProfileReport(&reportTheme, options, reportData)
		if err != nil {
			return err
		}
		printRecapHeader("RECAP ", theme.Text.HeaderFiller)
		err = PrintSummaryReport(&summaryTheme, summaryOptions, reportData)
		if err != nil {
			return err
		}

		return nil
	}

	for _, v := range spec.Report {
		switch v {
		case "recap":
			printRecapHeader("RECAP ", theme.Text.HeaderFiller)
			err := PrintSummaryReport(&summaryTheme, summaryOptions, reportData)
			if err != nil {
				return err
			}
		case "rc":
			printRecapHeader("RETURN CODES ", theme.Text.HeaderFiller)
			err := PrintExitReport(&reportTheme, options, reportData)
			if err != nil {
				return err
			}
		case "task":
			printRecapHeader("TASK STATUS ", theme.Text.HeaderFiller)
			err := PrintTaskReport(&reportTheme, options, reportData)
			if err != nil {
				return err
			}
		case "time":
			printRecapHeader("TIME ", theme.Text.HeaderFiller)
			err := PrintProfileReport(&reportTheme, options, reportData)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

/*
	               Return Code
	Server      Output-0  Output-1  Output-2

------------------------------------------

	list-0      1
	list-1      0         1
	list-2      1
	list-3      0         0         0
	list-4      0         0         0
*/
func PrintExitReport(
	theme *dao.Theme,
	options PrintTableOptions,
	reportData dao.ReportData,
) error {
	theme.Table.Options.SeparateFooter = core.Ptr(false)
	var data dao.TableOutput
	for i := range reportData.Tasks {
		name := getStatusName(reportData.Tasks[i].Name, reportData.Tasks[i].Status)
		data.Rows = append(data.Rows, dao.Row{Columns: []string{name}})

		for _, t := range reportData.Tasks[i].Rows {
			if t.Status == dao.Skipped || t.Status == dao.Unreachable {
				data.Rows[i].Columns = append(data.Rows[i].Columns, "")
			} else {
				v := strconv.Itoa(t.ReturnCode)
				if t.ReturnCode > 0 {
					v = FailedPrint.Sprint(v)
				} else {
					v = OkPrint.Sprint(v)
				}
				data.Rows[i].Columns = append(data.Rows[i].Columns, v)
			}
		}
	}

	err := PrintTable(data.Rows, options, reportData.Headers, []string{}, true, false)
	if err != nil {
		return err
	}

	return nil
}

/*
	Server  Output-0  Output-1  Output-2

--------------------------------------

	list-0  0.05 s    0.00 s    0.00 s
	list-1  1.05 s    0.01 s    0.00 s
	list-2  0.01 s    0.00 s    0.00 s
	list-3  1.01 s    0.00 s    0.00 s
	list-4  1.01 s    0.00 s    0.00 s
*/
func PrintProfileReport(
	theme *dao.Theme,
	options PrintTableOptions,
	reportData dao.ReportData,
) error {
	theme.Table.Options.SeparateFooter = core.Ptr(true)
	var data dao.TableOutput
	data.Headers = append(reportData.Headers, "Total")

	var taskDuration []time.Duration
	for i := 1; i < len(reportData.Headers); i++ {
		taskDuration = append(taskDuration, time.Duration(0))
	}

	for i := range reportData.Tasks {
		name := getStatusName(reportData.Tasks[i].Name, reportData.Tasks[i].Status)
		data.Rows = append(data.Rows, dao.Row{Columns: []string{name}})

		var sDuration time.Duration
		for k, t := range reportData.Tasks[i].Rows {
			if t.Status == dao.Skipped || t.Status == dao.Unreachable {
				data.Rows[i].Columns = append(data.Rows[i].Columns, "")
			} else {
				seconds := NormalPrint.Sprintf("%.2f s", t.Duration.Seconds())
				data.Rows[i].Columns = append(data.Rows[i].Columns, seconds)
				sDuration += t.Duration
				taskDuration[k] += t.Duration
			}
		}
		tSeconds := NormalPrint.Sprintf("%.2f s", sDuration.Seconds())
		data.Rows[i].Columns = append(data.Rows[i].Columns, tSeconds)
	}

	// Don't calculate total if only 1 server
	if len(reportData.Tasks) > 1 {
		footerName := getStatusName("Total", reportData.Status)
		data.Footers = append(data.Footers, footerName)
		var tot time.Duration
		for _, t := range taskDuration {
			v := NormalPrint.Sprintf("%.2f s", t.Seconds())
			data.Footers = append(data.Footers, v)
			tot += t
		}
		data.Footers = append(data.Footers, NormalPrint.Sprintf("%.2f s", tot.Seconds()))
	}

	err := PrintTable(data.Rows, options, data.Headers, data.Footers, true, false)
	if err != nil {
		return err
	}

	return nil
}

/*
	                  Task
	Server      Output-0  Output-1  Output-2

------------------------------------------

	list-0      ignored   ok        ok
	list-1      ok        ignored   ok
	list-2      ignored   ignored   ok
	list-3      ok        ok        ok
	list-4      ok        ok        ok
*/
func PrintTaskReport(
	theme *dao.Theme,
	options PrintTableOptions,
	reportData dao.ReportData,
) error {
	theme.Table.Options.SeparateFooter = core.Ptr(false)
	var data dao.TableOutput
	for i := range reportData.Tasks {
		name := getStatusName(reportData.Tasks[i].Name, reportData.Tasks[i].Status)
		data.Rows = append(data.Rows, dao.Row{Columns: []string{name}})

		for _, t := range reportData.Tasks[i].Rows {
			var v string
			switch t.Status {
			case dao.Ok:
				v = OkPrint.Sprint(t.Status.String())
			case dao.Skipped:
				v = SkippedPrint.Sprint(t.Status.String())
			case dao.Ignored:
				v = IgnoredPrint.Sprint(t.Status.String())
			case dao.Failed:
				v = FailedPrint.Sprint(t.Status.String())
			case dao.Unreachable:
				v = UnreachablePrint.Sprint(t.Status.String())
			}

			data.Rows[i].Columns = append(data.Rows[i].Columns, v)
		}
	}

	err := PrintTable(data.Rows, options, reportData.Headers, []string{}, true, false)
	if err != nil {
		return err
	}

	return nil
}

/*
	Summary

Server      Ok    Ignored    Failed    Skipped
-------------------------------------------------
list-0      ok=2  ignored=1  failed=0  skipped=0
list-1      ok=2  ignored=1  failed=0  skipped=0
list-2      ok=1  ignored=2  failed=0  skipped=0
list-3      ok=3  ignored=0  failed=0  skipped=0
list-4      ok=3  ignored=0  failed=0  skipped=0
*/
func PrintSummaryReport(
	theme *dao.Theme,
	options PrintTableOptions,
	reportData dao.ReportData,
) error {
	theme.Table.Options.SeparateHeader = core.Ptr(false)
	theme.Table.Options.SeparateFooter = core.Ptr(false)

	var data dao.TableOutput
	data.Headers = []string{"", "", "", "", "", ""}
	var taskStatuses = []dao.TaskStatus{
		dao.Ok,
		dao.Unreachable,
		dao.Ignored,
		dao.Failed,
		dao.Skipped,
	}

	for i := range reportData.Tasks {
		data.Rows = append(data.Rows, dao.Row{})
		name := getStatusName(reportData.Tasks[i].Name, reportData.Tasks[i].Status)
		data.Rows[i].Columns = append(data.Rows[i].Columns, name)
		for _, s := range taskStatuses {
			val := getTotalStatus(s, reportData.Tasks[i].Status)
			data.Rows[i].Columns = append(data.Rows[i].Columns, val)
		}
	}

	// Don't calculate total if only 1 server
	if len(reportData.Tasks) > 1 {
		theme.Table.Options.SeparateFooter = core.Ptr(true)
		if reportData.Status[dao.Failed] == 0 && reportData.Status[dao.Unreachable] == 0 {
			tot := OkPrint.Sprintf("%s", "Total")
			data.Footers = append(data.Footers, tot)
		} else if reportData.Status[dao.Unreachable] > 0 {
			data.Footers = append(data.Footers, FailedPrint.Sprintf("%s", "Total"))
		} else if reportData.Status[dao.Ok] == 0 {
			data.Footers = append(data.Footers, SkippedPrint.Sprintf("%s", "Total"))
		} else {
			data.Footers = append(data.Footers, FailedPrint.Sprintf("%s", "Total"))
		}
		for _, s := range taskStatuses {
			val := getTotalStatus(s, reportData.Status)
			data.Footers = append(data.Footers, val)
		}
	}

	err := PrintTable(data.Rows, options, data.Headers, data.Footers, false, true)
	if err != nil {
		return err
	}

	return nil
}

func printRecapHeader(h string, filler string) {
	hh := text.Bold.Sprint(h)
	width, _, _ := term.GetSize(0)
	headerLength := len(core.Strip(hh))
	if width > 0 {
		header := fmt.Sprintf("\n%s%s", hh, strings.Repeat(filler, width-headerLength-1))
		fmt.Println(header)
	}
}

func getStatusName(name string, status map[dao.TaskStatus]int) string {
	var out string
	if status[dao.Failed] > 0 || status[dao.Unreachable] > 0 {
		out = FailedPrint.Sprintf("%s\t", name)
	} else if status[dao.Ok] == 0 && status[dao.Skipped] > 0 {
		out = SkippedPrint.Sprintf("%s\t", name)
	} else {
		out = OkPrint.Sprintf("%s\t", name)
	}
	return out
}

// func getTotalStatus(s dao.TaskStatus, reportData dao.ReportData) string {
func getTotalStatus(s dao.TaskStatus, status map[dao.TaskStatus]int) string {
	var val string
	vv := int(status[s])
	v := strconv.Itoa(vv)

	if vv > 0 {
		switch s {
		case dao.Ok:
			val = OkPrint.Sprintf("%s=%s", s, v)
		case dao.Skipped:
			val = SkippedPrint.Sprintf("%s=%s", s, v)
		case dao.Ignored:
			val = IgnoredPrint.Sprintf("%s=%s", s, v)
		case dao.Failed:
			val = FailedPrint.Sprintf("%s=%s", s, v)
		case dao.Unreachable:
			val = FailedPrint.Sprintf("%s=%s", s, v)
		}
	} else {
		val = ZeroPrint.Sprintf("%s=%s", s, v)
	}

	return val
}
