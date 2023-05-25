package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/alajmo/sake/core/dao"
)

const (
	appName      = "sake"
	shortAppDesc = "sake is a task runner for local and remote hosts"
	longAppDesc  = `sake is a task runner for local and remote hosts.

You define servers and tasks in a sake.yaml config file and then run the tasks on the servers.
`
)

var (
	config         dao.Config
	configErr      error
	configPath     string
	userConfigPath string
	sshConfigPath  string
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
	if runtime.GOOS == "windows" {
		dao.DEFAULT_SHELL = "pwsh -NoProfile -command"
	}

	if runtime.GOOS == "darwin" {
		dao.DEFAULT_SHELL = "zsh -c"
	}

	cobra.OnInitialize(initConfig)

	cobra.EnableCommandSorting = false

	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "specify config")
	rootCmd.PersistentFlags().StringVarP(&userConfigPath, "user-config", "u", "", "specify user config")
	rootCmd.PersistentFlags().StringVar(&sshConfigPath, "ssh-config", "", "specify ssh config")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color")

	rootCmd.AddCommand(
		initCmd(),
		listCmd(&config, &configErr),
		describeCmd(&config, &configErr),
		runCmd(&config, &configErr),
		execCmd(&config, &configErr),
		sshCmd(&config, &configErr),
		editCmd(&config, &configErr),
		checkCmd(&configErr),
		completionCmd(),
		genCmd(),
	)

	rootCmd.SetVersionTemplate(fmt.Sprintf("Version: %-10s\nCommit: %-10s\nDate: %-10s\n", version, commit, date))

	if buildMode == "man" {
		rootCmd.AddCommand(genDocsCmd(longAppDesc))
	}
}

func initConfig() {
	config, configErr = dao.ReadConfig(configPath, userConfigPath, sshConfigPath, noColor)
}
