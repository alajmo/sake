package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
)

type AlreadySakeDirectory struct {
	Dir string
}

func (c *AlreadySakeDirectory) Error() string {
	return fmt.Sprintf("`%s` is already a sake directory\n", c.Dir)
}

type ConfigEnvFailed struct {
	Name string
	Err  string
}

func (c *ConfigEnvFailed) Error() string {
	return fmt.Sprintf("failed to evaluate env `%s` \n  %s", c.Name, c.Err)
}

type PasswordEvalFailed struct {
	Err string
}

func (c *PasswordEvalFailed) Error() string {
	return fmt.Sprintf("failed to evaluate password %s", c.Err)
}

type TagNotFound struct {
	Tags []string
}

func (c *TagNotFound) Error() string {
	tags := "`" + strings.Join(c.Tags, "`, `") + "`"
	return fmt.Sprintf("cannot find tags %s", tags)
}

type ServerNotFound struct {
	Name []string
}

func (c *ServerNotFound) Error() string {
	servers := "`" + strings.Join(c.Name, "`, `") + "`"
	return fmt.Sprintf("cannot find servers %s", servers)
}

type TaskNotFound struct {
	IDs []string
}

func (c *TaskNotFound) Error() string {
	tasks := "`" + strings.Join(c.IDs, "`, `") + "`"
	return fmt.Sprintf("cannot find tasks %s", tasks)
}

type TaskMultipleDef struct {
	Name string
}

func (c *TaskMultipleDef) Error() string {
	return fmt.Sprintf("can only define one of the following for task `%s`: cmd, task, tasks", c.Name)
}

type ThemeNotFound struct {
	Name string
}

func (c *ThemeNotFound) Error() string {
	return fmt.Sprintf("cannot find theme `%s`", c.Name)
}

type SpecNotFound struct {
	Name string
}

func (c *SpecNotFound) Error() string {
	return fmt.Sprintf("cannot find spec `%s`", c.Name)
}

type TargetNotFound struct {
	Name string
}

func (c *TargetNotFound) Error() string {
	return fmt.Sprintf("cannot find target `%s`", c.Name)
}

type ConfigNotFound struct {
	Names []string
}

func (f *ConfigNotFound) Error() string {
	return fmt.Sprintf("cannot find any configuration file %v in current directory or any of the parent directories", f.Names)
}

type ConfigErr struct {
	Msg string
}

func (f *ConfigErr) Error() string {
	return f.Msg
}

type FileError struct {
	Err string
}

func (f *FileError) Error() string {
	return f.Err
}

type NoRemoteServerToAttach struct{}

func (c *NoRemoteServerToAttach) Error() string {
	return fmt.Sprintf("no remote server to ssh into")
}

type NoEditorEnv struct{}

func (c *NoEditorEnv) Error() string {
	return fmt.Sprintf("no environment variable `EDITOR` found")
}

func CheckIfError(err error) {
	if err != nil {
		switch err.(type) {
		case *ConfigErr:
			// Errors are already mapped with `error:` prefix
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)
		default:
			fmt.Fprintf(os.Stderr, "%s: %v\n", text.FgRed.Sprintf("error"), err)
			os.Exit(1)
		}
	}
}
