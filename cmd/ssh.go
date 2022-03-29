package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/run"
)

func sshCmd(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Use:   "ssh <server> [flags]",
		Short: "ssh to server",
		Long:  `ssh to server.`,

		Example: `  # ssh to server
  sake ssh <server>`,
		Args:                  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			ssh(args, config)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			return config.GetRemoteServerNameAndDesc(), cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	return &cmd
}

func ssh(args []string, config *dao.Config) {
	server, err := config.GetServer(args[0])
	core.CheckIfError(err)

	err = run.SSHToServer(server.Host, server.User, server.Port, config.DisableVerifyHost, config.KnownHostsFile)
	core.CheckIfError(err)
}
