package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/print"
)

func initCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "init [flags]",
		Short: "Initialize sake in the current directory",
		Long: "Initialize sake in the current directory.",
		Example: `  # Basic example
  sake init`,

		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			servers, err := dao.InitSake(args)
			core.CheckIfError(err)
			PrintServerInit(servers)
		},
		DisableAutoGenTag: true,
	}

	return &cmd
}

func PrintServerInit(servers []dao.Server) {
	theme := dao.Theme{
		Table: dao.DefaultTable,
	}

	options := print.PrintTableOptions{
		Theme:                theme,
		OmitEmpty:            true,
		Output:               "table",
		SuppressEmptyColumns: false,
	}

	data := dao.TableOutput{
		Headers: []string{"server", "host"},
		Rows:    []dao.Row{},
	}

	for _, server := range servers {
		data.Rows = append(data.Rows, dao.Row{Columns: []string{server.Name, server.Host}})
	}

	fmt.Println("\nFollowing servers were added to sake.yaml")
	print.PrintTable("", data.Rows, options, data.Headers, []string{})
}
