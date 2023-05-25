// This source will generate
//   - core/sake.1
//   - docs/command-reference.md
//
// and is not included in the final build.

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
)

func genDocsCmd(longAppDesc string) *cobra.Command {
	cmd := cobra.Command{
		Use:   "gen-docs",
		Short: "Generate man and markdown pages",
		Run: func(cmd *cobra.Command, args []string) {
			err := core.CreateManPage(
				longAppDesc,
				version,
				date,
				rootCmd,
				checkCmd(&configErr),
				runCmd(&config, &configErr),
				execCmd(&config, &configErr),
				initCmd(),
				editCmd(&config, &configErr),
				listCmd(&config, &configErr),
				describeCmd(&config, &configErr),
				sshCmd(&config, &configErr),
				genCmd(),
			)
			core.CheckIfError(err)
		},

		DisableAutoGenTag: true,
	}

	return &cmd
}
