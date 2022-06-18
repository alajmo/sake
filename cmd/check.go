package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
)

func checkCmd(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Use:   "check",
		Short: "Validate config",
		Long: `Validate config.`,
		Example: `  # Validate config
  sake check`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if *configErr != nil {
				fmt.Printf("Found configuration errors:\n\n")
				core.Exit(*configErr)
			}

			fmt.Println("Config Valid")
		},
		DisableAutoGenTag: true,
	}

	return &cmd
}
