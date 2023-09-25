package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/run"
)

func sshCmd(config *dao.Config, configErr *error) *cobra.Command {
	var runFlags core.RunFlags

	cmd := cobra.Command{
		Use:   "ssh <server> [flags]",
		Short: "ssh to server",
		Long:  `ssh to server.`,

		Example: `  # ssh to server
  sake ssh <server>`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			ssh(args, config, &runFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			return config.GetRemoteServerNameAndDesc(), cobra.ShellCompDirectiveNoFileComp
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().SortFlags = false

	cmd.Flags().StringVarP(&runFlags.IdentityFile, "identity-file", "i", "", "set identity file for all servers")
	cmd.Flags().StringVar(&runFlags.Password, "password", "", "set ssh password for all servers")

	return &cmd
}

func ssh(args []string, config *dao.Config, runFlags *core.RunFlags) {
	server, err := config.GetServer(args[0])
	core.CheckIfError(err)
	servers := []dao.Server{*server}

	errConnect, err := run.ParseServers(config.SSHConfigFile, &servers, runFlags, "inventory")
	if len(errConnect) > 0 {
		core.Exit(&errConnect[0])
	}
	core.CheckIfError(err)

	err = run.SSHToServer(servers[0], config.DisableVerifyHost, config.KnownHostsFile)
	core.CheckIfError(err)
}
