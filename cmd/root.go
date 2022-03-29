package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core/dao"
)

const (
	appName      = "sake"
	shortAppDesc = "sake is a CLI tool that enables you to run commands on servers via ssh"
	longAppDesc  = `sake is a CLI tool that enables you to run commands on servers via ssh.

Think of it like make, you define servers and tasks in a declarative configuration file and then run the tasks on the servers.
`
)

var (
	config         dao.Config
	configErr      error
	configFilepath string
	userConfigPath string
	noColor        bool
	buildMode      = ""
	version        = "dev"
	commit         = "none"
	date           = "n/a"
	rootCmd        = &cobra.Command{
		Use:     appName,
		Short:   shortAppDesc,
		Long:    longAppDesc,
		Version: version,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// When user input's wrong command or flag
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&configFilepath, "config", "c", "", "specify config")
	rootCmd.PersistentFlags().StringVarP(&userConfigPath, "user-config", "u", "", "specify user config")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color")

	rootCmd.AddCommand(
		completionCmd(),
		genCmd(),
		initCmd(),
		listCmd(&config, &configErr),
		describeCmd(&config, &configErr),
		editCmd(&config, &configErr),
		execCmd(&config, &configErr),
		runCmd(&config, &configErr),
		sshCmd(&config, &configErr),
	)

	rootCmd.SetVersionTemplate(fmt.Sprintf("Version: %-10s\nCommit: %-10s\nDate: %-10s\n", version, commit, date))

	if buildMode == "man" {
		rootCmd.AddCommand(genDocsCmd(longAppDesc))
	}
}

func initConfig() {
	config, configErr = dao.ReadConfig(configFilepath, userConfigPath, noColor)
}
