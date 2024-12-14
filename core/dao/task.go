package dao

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/sake/core"
)

var REGISTER_REGEX = regexp.MustCompile("^[a-zA-Z_]+[a-zA-Z0-9_]*$")

// This is the struct that is added to the Task.Tasks in import_task.go
type TaskCmd struct {
	ID           string
	Name         string
	Desc         string
	WorkDir      string
	Shell        string
	RootDir      string
	Register     string
	Cmd          string
	Local        bool
	TTY          bool
	IgnoreErrors bool
	Envs         []string
}

// This is the struct that is added to the Task.TaskRefs
type TaskRef struct {
	Name         string
	Desc         string
	Cmd          string
	WorkDir      string
	Shell        string
	Register     string
	Task         string
	Local        *bool
	TTY          *bool
	IgnoreErrors *bool
	Envs         []string
}

type Task struct {
	ID      string
	Name    string
	Desc    string
	TTY     bool
	Local   bool
	Attach  bool
	WorkDir string
	Shell   string
	Envs    []string
	Cmd     string
	Tasks   []TaskCmd
	Spec    Spec
	Target  Target
	Theme   Theme

	TaskRefs  []TaskRef
	SpecRef   string
	TargetRef string
	ThemeRef  string

	context     string // config path
	contextLine int    // defined at
}

// Unmarshaled from YAML
type TaskYAML struct {
	Name    string        `yaml:"name"`
	Desc    string        `yaml:"desc"`
	Local   bool          `yaml:"local"`
	TTY     bool          `yaml:"tty"`
	Attach  bool          `yaml:"attach"`
	WorkDir string        `yaml:"work_dir"`
	Shell   string        `yaml:"shell"`
	Cmd     string        `yaml:"cmd"`
	Task    string        `yaml:"task"`
	Tasks   []TaskRefYAML `yaml:"tasks"`
	Env     yaml.Node     `yaml:"env"`
	Spec    yaml.Node     `yaml:"spec"`
	Target  yaml.Node     `yaml:"target"`
	Theme   yaml.Node     `yaml:"theme"`
}

// Unmarshaled from YAML
type TaskRefYAML struct {
	Name         string    `yaml:"name"`
	Desc         string    `yaml:"desc"`
	WorkDir      string    `yaml:"work_dir"`
	Shell        string    `yaml:"shell"`
	Cmd          string    `yaml:"cmd"`
	Task         string    `yaml:"task"`
	Register     string    `yaml:"register"`
	Local        *bool     `yaml:"local"`
	IgnoreErrors *bool     `yaml:"ignore_errors"`
	TTY          *bool     `yaml:"tty"`
	Env          yaml.Node `yaml:"env"`
}

func (t Task) GetValue(key string, _ int) string {
	lkey := strings.ToLower(key)
	switch lkey {
	case "name", "task":
		return t.Name
	case "desc", "description":
		return t.Desc
	case "local":
		return strconv.FormatBool(t.Local)
	case "tty":
		return strconv.FormatBool(t.TTY)
	case "attach":
		return strconv.FormatBool(t.Attach)
	case "work_dir":
		return t.WorkDir
	case "shell":
		return t.Shell
	case "spec":
		return t.Spec.Name
	case "target":
		return t.Target.Name
	case "theme":
		return t.Theme.Name
	default:
		return ""
	}
}

func (t *Task) GetContext() string {
	return t.context
}

func (t *Task) GetContextLine() int {
	return t.contextLine
}

// ParseTasksYAML parses the task dictionary and returns it as a list.
// This function also sets task references.
// Valid formats (only one is allowed):
//
//	cmd: |
//	  echo pong
//
//	task: ping
//
//	tasks:
//	  - task: ping
//	  - task: ping
//	  - cmd: echo pong
func (c *ConfigYAML) ParseTasksYAML() ([]Task, []ResourceErrors[Task]) {
	var tasks []Task
	count := len(c.Tasks.Content)

	taskErrors := []ResourceErrors[Task]{}
	j := -1
	for i := 0; i < count; i += 2 {
		j += 1
		task := &Task{
			ID:          c.Tasks.Content[i].Value,
			context:     c.Path,
			contextLine: c.Tasks.Content[i].Line,
		}
		re := ResourceErrors[Task]{Resource: task, Errors: []error{}}
		taskErrors = append(taskErrors, re)
		taskYAML := &TaskYAML{}

		if c.Tasks.Content[i+1].Kind == 8 {
			// Shorthand definition:
			// ping: echo 123
			taskYAML.Name = c.Tasks.Content[i].Value
			taskYAML.Cmd = c.Tasks.Content[i+1].Value
		} else {
			// Full definition:
			// ping:
			//   cmd: echo 123
			err := c.Tasks.Content[i+1].Decode(taskYAML)

			if err != nil {
				for _, yerr := range err.(*yaml.TypeError).Errors {
					taskErrors[j].Errors = append(taskErrors[j].Errors, errors.New(yerr))
				}
			}

			// Check that only 1 one 3 is defined (cmd, task, tasks)
			numDefined := 0
			if taskYAML.Cmd != "" {
				numDefined += 1
			}
			if taskYAML.Task != "" {
				numDefined += 1
			}
			if len(taskYAML.Tasks) > 0 {
				numDefined += 1
			}
			if numDefined > 1 {
				taskErrors[j].Errors = append(taskErrors[j].Errors, &core.TaskMultipleDef{Name: c.Tasks.Content[i].Value})
			}

			if numDefined > 1 || err != nil {
				continue
			}
		}

		if taskYAML.Name != "" {
			task.Name = taskYAML.Name
		} else {
			task.Name = c.Tasks.Content[i].Value
		}
		task.Desc = taskYAML.Desc
		task.TTY = taskYAML.TTY
		task.Local = taskYAML.Local
		task.WorkDir = taskYAML.WorkDir
		task.Shell = taskYAML.Shell
		task.Attach = taskYAML.Attach

		if !IsNullNode(taskYAML.Env) {
			err := CheckIsMappingNode(taskYAML.Env)
			if err != nil {
				taskErrors[j].Errors = append(taskErrors[j].Errors, err)
			} else {
				task.Envs = append(task.Envs, ParseNodeEnv(taskYAML.Env)...)
			}
		}

		task.Tasks = []TaskCmd{}
		task.TaskRefs = []TaskRef{}

		// Spec
		if len(taskYAML.Spec.Content) > 0 {
			// Inline Spec
			spec, specErrors := c.DecodeSpec("", taskYAML.Spec)
			taskErrors[j].Errors = append(taskErrors[j].Errors, specErrors...)
			task.Spec = *spec
		} else if taskYAML.Spec.Value != "" {
			// Spec reference
			task.SpecRef = taskYAML.Spec.Value
		} else {
			task.SpecRef = DEFAULT_SPEC.Name
		}

		// Target
		if len(taskYAML.Target.Content) > 0 {
			// Inline Target
			target, targetErrors := c.DecodeTarget("", taskYAML.Target)
			taskErrors[j].Errors = append(taskErrors[j].Errors, targetErrors...)
			task.Target = *target
		} else if taskYAML.Target.Value != "" {
			// Target reference
			task.TargetRef = taskYAML.Target.Value
		} else {
			task.TargetRef = DEFAULT_TARGET.Name
		}

		// Theme
		if len(taskYAML.Theme.Content) > 0 {
			// Inline Theme
			theme := &Theme{}
			err := taskYAML.Theme.Decode(theme)
			if err != nil {
				for _, yerr := range err.(*yaml.TypeError).Errors {
					taskErrors[j].Errors = append(taskErrors[j].Errors, errors.New(yerr))
				}
			} else {
				task.Theme = *theme
			}
		} else if taskYAML.Theme.Value != "" {
			// Theme reference
			task.ThemeRef = taskYAML.Theme.Value
		} else {
			task.ThemeRef = DEFAULT_THEME.Name
		}

		// Set task cmd/reference
		if taskYAML.Task != "" {
			// Task Reference
			tr := TaskRef{
				Task: taskYAML.Task,
			}

			task.TaskRefs = append(task.TaskRefs, tr)
		} else if len(taskYAML.Tasks) > 0 {
			// Tasks References
			for k := range taskYAML.Tasks {
				tr := TaskRef{
					Name:         taskYAML.Tasks[k].Name,
					Desc:         taskYAML.Tasks[k].Desc,
					WorkDir:      taskYAML.Tasks[k].WorkDir,
					Shell:        taskYAML.Tasks[k].Shell,
					Local:        taskYAML.Tasks[k].Local,
					TTY:          taskYAML.Tasks[k].TTY,
					IgnoreErrors: taskYAML.Tasks[k].IgnoreErrors,
					Envs:         ParseNodeEnv(taskYAML.Tasks[k].Env),
				}

				if taskYAML.Tasks[k].Register != "" {
					match := REGISTER_REGEX.MatchString(taskYAML.Tasks[k].Register)
					if match {
						tr.Register = taskYAML.Tasks[k].Register
					} else {
						taskErrors[j].Errors = append(taskErrors[j].Errors, &core.RegisterInvalidName{Value: taskYAML.Tasks[k].Register})
						continue
					}
				}

				// TODO: What about this?
				// Find servers matching the flag
				// var servers []Server
				// for _, server := range c.Servers {
				// 	match := pattern.MatchString(server.Host)
				// 	if match {
				// 		servers = append(servers, server)
				// 	}
				// }

				// Check that only cmd or task is defined
				if taskYAML.Tasks[k].Cmd != "" && taskYAML.Tasks[k].Task != "" {
					taskErrors[j].Errors = append(taskErrors[j].Errors, &core.TaskRefMultipleDef{Name: c.Tasks.Content[i].Value})
					continue
				} else if taskYAML.Tasks[k].Cmd != "" {
					tr.Cmd = taskYAML.Tasks[k].Cmd
				} else if taskYAML.Tasks[k].Task != "" {
					tr.Task = taskYAML.Tasks[k].Task
				} else {
					taskErrors[j].Errors = append(taskErrors[j].Errors, &core.NoTaskRefDefined{Name: c.Tasks.Content[i].Value})
					continue
				}

				task.TaskRefs = append(task.TaskRefs, tr)
			}
		} else if taskYAML.Cmd != "" {
			// Command
			task.Cmd = taskYAML.Cmd
		}

		tasks = append(tasks, *task)
	}

	return tasks, taskErrors
}

func ParseTaskEnv(cmdEnv []string, userEnv []string, parentEnv []string, configEnv []string) ([]string, error) {
	cmdEnv, err := EvaluateEnv(cmdEnv)
	if err != nil {
		return []string{}, err
	}

	pEnv, err := EvaluateEnv(parentEnv)
	if err != nil {
		return []string{}, err
	}

	envs := MergeEnvs(userEnv, cmdEnv, pEnv, configEnv)

	return envs, nil
}

func (c *Config) GetTaskServers(task *Task, runFlags *core.RunFlags, setRunFlags *core.SetRunFlags) ([]Server, error) {
	var servers []Server
	var err error
	// If any runtime target flags are used, disregard config specified task targets
	if len(runFlags.Servers) > 0 || len(runFlags.Tags) > 0 || runFlags.Regex != "" || setRunFlags.All || setRunFlags.Invert {
		servers, err = c.FilterServers(runFlags.All, runFlags.Servers, runFlags.Tags, runFlags.Regex, runFlags.Invert)
		if err != nil {
			return []Server{}, err
		}
	} else if runFlags.Target != "" {
		target, err := c.GetTarget(runFlags.Target)
		if err != nil {
			return []Server{}, err
		}
		task.Target = *target
		servers, err = c.FilterServers(task.Target.All, task.Target.Servers, task.Target.Tags, task.Target.Regex, runFlags.Invert)
		if err != nil {
			return []Server{}, err
		}
	} else {
		servers, err = c.FilterServers(task.Target.All, task.Target.Servers, task.Target.Tags, task.Target.Regex, runFlags.Invert)
		if err != nil {
			return []Server{}, err
		}
	}

	var limit uint32
	if runFlags.Limit > 0 {
		limit = runFlags.Limit
	} else if task.Target.Limit > 0 {
		limit = task.Target.Limit
	}

	var limitp uint8
	if runFlags.LimitP > 0 {
		limitp = runFlags.LimitP
	} else if task.Target.LimitP > 0 {
		limitp = task.Target.LimitP
	}

	if limit > 0 {
		if limit <= uint32(len(servers)) {
			return servers[0:limit], nil
		}
	} else if limitp > 0 {
		if limitp <= 100 {
			tot := float64(len(servers))
			percentage := float64(limitp) / float64(100)
			limit := math.Floor(percentage * tot)

			if limit > 0 {
				return servers[0:int(limit)], nil
			} else {
				return servers[0:1], nil
			}
		} else {
			return []Server{}, &core.InvalidPercentInput{}
		}
	}

	return servers, nil
}

func (c *Config) GetTasksByIDs(ids []string) ([]Task, error) {
	if len(ids) == 0 {
		return c.Tasks, nil
	}

	foundTasks := make(map[string]bool)
	for _, t := range ids {
		foundTasks[t] = false
	}

	var filteredTasks []Task
	for _, id := range ids {
		if foundTasks[id] {
			continue
		}

		for _, task := range c.Tasks {
			if id == task.ID {
				foundTasks[task.ID] = true
				filteredTasks = append(filteredTasks, task)
			}
		}
	}

	nonExistingTasks := []string{}
	for k, v := range foundTasks {
		if !v {
			nonExistingTasks = append(nonExistingTasks, k)
		}
	}

	if len(nonExistingTasks) > 0 {
		return []Task{}, &core.TaskNotFound{IDs: nonExistingTasks}
	}

	return filteredTasks, nil
}

func (c *Config) GetTaskIDAndDesc() []string {
	taskNames := []string{}
	for _, task := range c.Tasks {
		if task.Spec.Hidden {
			continue
		}

		if task.Desc != "" {
			taskNames = append(taskNames, fmt.Sprintf("%s\t%s", task.ID, task.Desc))
		} else if task.ID != task.Name {
			taskNames = append(taskNames, fmt.Sprintf("%s\t%s", task.ID, task.Name))
		} else {
			taskNames = append(taskNames, task.ID)
		}
	}

	return taskNames
}

func (c *Config) GetTask(id string) (*Task, error) {
	for _, task := range c.Tasks {
		if id == task.ID {
			return &task, nil
		}
	}

	return nil, &core.TaskNotFound{IDs: []string{id}}
}

type TaskStatus int64

const (
	Skipped TaskStatus = iota
	Ok
	Failed
	Ignored
	Unreachable
)

type Report struct {
	ReturnCode int
	Duration   time.Duration
	Status     TaskStatus
}

type ReportRow struct {
	Name   string
	Status map[TaskStatus]int
	Rows   []Report
}

type ReportData struct {
	Headers []string
	Tasks   []ReportRow
	Status  map[TaskStatus]int
}

func (s TaskStatus) String() string {
	switch s {
	case Ok:
		return "ok"
	case Failed:
		return "failed"
	case Skipped:
		return "skipped"
	case Ignored:
		return "ignored"
	case Unreachable:
		return "unreachable"
	}

	return ""
}
