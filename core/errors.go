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

type InventoryEvalFailed struct {
	Err string
}

func (c *InventoryEvalFailed) Error() string {
	return fmt.Sprintf("failed to run inventory command %s", c.Err)
}

type TagNotFound struct {
	Tags []string
}

func (c *TagNotFound) Error() string {
	tags := "`" + strings.Join(c.Tags, "`, `") + "`"
	return fmt.Sprintf("cannot find tags %s", tags)
}

type TargetsNotFound struct {
	Targets []string
}

func (c *TargetsNotFound) Error() string {
	targets := "`" + strings.Join(c.Targets, "`, `") + "`"
	return fmt.Sprintf("cannot find targets %s", targets)
}

type SpecsNotFound struct {
	Specs []string
}

func (c *SpecsNotFound) Error() string {
	specs := "`" + strings.Join(c.Specs, "`, `") + "`"
	return fmt.Sprintf("cannot find specs %s", specs)
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

type LimitMultipleDef struct {
	Name string
}

func (c *LimitMultipleDef) Error() string {
	if c.Name != "" {
		return fmt.Sprintf("can only define one of the following for target `%s`: limit, limit_p", c.Name)
	}

	return "can only define one of the following for target: limit, limit_p"
}

type MultipleFailSet struct {
	Name string
}

func (c *MultipleFailSet) Error() string {
	if c.Name != "" {
		return fmt.Sprintf("can only define either `any_errors_fatal` or `max_fail_percentage` but not both spec `%s`", c.Name)
	}

	return "can only define either `any_errors_fatal` or `max_fail_percentage` but not both spec"
}

type BatchMultipleDef struct {
	Name string
}

func (c *BatchMultipleDef) Error() string {
	if c.Name != "" {
		return fmt.Sprintf("can only define one of the following for spec `%s`: batch, batch_p", c.Name)
	}

	return "can only define one of the following for spec: batch, batch_p"
}

type ZeroNotAllowed struct {
	Name string
}

func (c *ZeroNotAllowed) Error() string {
	return fmt.Sprintf("invalid value for %s, cannot be 0", c.Name)
}

type InvalidPercentInput struct {
	Name string
}

func (c *InvalidPercentInput) Error() string {
	return fmt.Sprintf("percentage can only be between 0 and 100 for property `%s`", c.Name)
}

type InvalidPercentInput2 struct {
	Name string
}

func (c *InvalidPercentInput2) Error() string {
	return fmt.Sprintf("percentage can only be between 1 and 100 for property `%s`", c.Name)
}

type RegisterInvalidName struct {
	Value string
}

func (c *RegisterInvalidName) Error() string {
	return fmt.Sprintf("invalid register variable name '%s', only alphanumeric characters and underscore are allowed and variable cannot start with a digit", c.Value)
}

type ServerMultipleDef struct {
	Name string
}

func (c *ServerMultipleDef) Error() string {
	return fmt.Sprintf("can only define one of the following for server `%s`: host, hosts", c.Name)
}

type ServerBastionMultipleDef struct {
	Name string
}

func (c *ServerBastionMultipleDef) Error() string {
	return fmt.Sprintf("can only define one of the following for server `%s`: bastion, bastions", c.Name)
}

type TaskRefMultipleDef struct {
	Name string
}

func (c *TaskRefMultipleDef) Error() string {
	return fmt.Sprintf("found `task` and `cmd` definition for sub tasks in task `%s`", c.Name)
}

type NoTaskRefDefined struct {
	Name string
}

func (c *NoTaskRefDefined) Error() string {
	return fmt.Sprintf("found no `task` or `cmd` definition for sub-task in task `%s`", c.Name)
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
	return "no remote server to ssh into"
}

type NoEditorEnv struct{}

func (c *NoEditorEnv) Error() string {
	return "no environment variable `EDITOR` found"
}

// If there's a misconfiguration with golang templates (prefix/header for instance in text.go)
type TemplateParseError struct {
	Msg string
}

func (f *TemplateParseError) Error() string {
	return fmt.Sprintf("failed to parse %s", f.Msg)
}

// If there's a misconfiguration somewhere, not associated with server errors
type ExecError struct {
	Err      error
	ExitCode int
}

func (e *ExecError) Error() string {
	return ""
}

func CheckIfError(err error) {
	if err != nil {
		Exit(err)
	}
}

func Exit(err error) {
	switch err := err.(type) {
	case *ConfigErr:
		// Errors are already mapped with `error:` prefix
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	case *ExecError:
		// Don't print anything when there's a ExecError: server execution failed
		os.Exit(err.ExitCode)
	default:
		fmt.Fprintf(os.Stderr, "%s: %v\n", text.FgRed.Sprintf("error"), err)
		os.Exit(1)
	}
}
