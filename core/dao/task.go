package dao

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/sake/core"
)

type TaskCmd struct {
	ID      string
	Name    string
	Desc    string
	WorkDir string
	Cmd     string
	Local   bool
	Envs    []string
}

type TaskRef struct {
	Name    string
	Desc    string
	Cmd     string
	WorkDir string
	Task    string
	Local   *bool
	Envs    []string
}

type TaskRefYAML struct {
	Name    string    `yaml:"name"`
	Desc    string    `yaml:"desc"`
	WorkDir string    `yaml:"work_dir"`
	Cmd     string    `yaml:"cmd"`
	Task    string    `yaml:"task"`
	Local   *bool     `yaml:"local"`
	Env     yaml.Node `yaml:"env"`
}

type Task struct {
	ID      string
	Name    string
	Desc    string
	TTY     bool
	Local   bool
	Attach  bool
	WorkDir string
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

type TaskYAML struct {
	Name    string        `yaml:"name"`
	Desc    string        `yaml:"desc"`
	Local   bool          `yaml:"local"`
	TTY     bool          `yaml:"tty"`
	Attach  bool          `yaml:"attach"`
	WorkDir string        `yaml:"work_dir"`
	Cmd     string        `yaml:"cmd"`
	Task    string        `yaml:"task"`
	Tasks   []TaskRefYAML `yaml:"tasks"`
	Env     yaml.Node     `yaml:"env"`
	Spec    yaml.Node     `yaml:"spec"`
	Target  yaml.Node     `yaml:"target"`
	Theme   yaml.Node     `yaml:"theme"`
}

func (t Task) GetValue(key string, _ int) string {
	switch key {
	case "Name", "name", "Task", "task":
		return t.Name
	case "Desc", "desc", "Description", "description":
		return t.Desc
	case "Command", "command":
		return t.Cmd
	}

	return ""
}

func (t Task) GetDefaultEnvs() []string {
	var defaultEnvs []string
	for _, env := range t.Envs {
		if strings.Contains(env, "SAKE_TASK_") {
			defaultEnvs = append(defaultEnvs, env)
		}
	}

	return defaultEnvs
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
//	 cmd: |
//	   echo pong
//
//	 task: ping
//
//	 tasks:
//	   - task: ping
//	   - task: ping
//	   - cmd: echo pong
//
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

			if err != nil {
				for _, yerr := range err.(*yaml.TypeError).Errors {
					taskErrors[j].Errors = append(taskErrors[j].Errors, errors.New(yerr))
				}
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
		task.Attach = taskYAML.Attach

		defaultEnvs := []string{
			fmt.Sprintf("SAKE_TASK_ID=%s", task.ID),
			fmt.Sprintf("SAKE_TASK_NAME=%s", taskYAML.Name),
			fmt.Sprintf("SAKE_TASK_DESC=%s", taskYAML.Desc),
			fmt.Sprintf("SAKE_TASK_LOCAL=%t", taskYAML.Local),
		}

		task.Envs = append(task.Envs, defaultEnvs...)

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
			// Spec value
			spec := &Spec{}
			err := taskYAML.Spec.Decode(spec)
			if err != nil {
				for _, yerr := range err.(*yaml.TypeError).Errors {
					taskErrors[j].Errors = append(taskErrors[j].Errors, errors.New(yerr))
				}
			} else {
				task.Spec = *spec
			}
		} else if taskYAML.Spec.Value != "" {
			// Spec reference
			task.SpecRef = taskYAML.Spec.Value
		} else {
			task.SpecRef = DEFAULT_SPEC.Name
		}

		// Target
		if len(taskYAML.Target.Content) > 0 {
			// Target value
			target := &Target{}
			err := taskYAML.Target.Decode(target)
			if err != nil {
				for _, yerr := range err.(*yaml.TypeError).Errors {
					taskErrors[j].Errors = append(taskErrors[j].Errors, errors.New(yerr))
				}
			} else {
				task.Target = *target
			}
		} else if taskYAML.Target.Value != "" {
			// Target reference
			task.TargetRef = taskYAML.Target.Value
		} else {
			task.TargetRef = DEFAULT_TARGET.Name
		}

		// Theme
		if len(taskYAML.Theme.Content) > 0 {
			// Theme value
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
			for i := range taskYAML.Tasks {
				tr := TaskRef{
					Name:    taskYAML.Tasks[i].Name,
					Desc:    taskYAML.Tasks[i].Desc,
					Cmd:     taskYAML.Tasks[i].Cmd,
					WorkDir: taskYAML.Tasks[i].WorkDir,
					Local:   taskYAML.Tasks[i].Local,
					Task:    taskYAML.Tasks[i].Task,
					Envs:    ParseNodeEnv(taskYAML.Tasks[i].Env),
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

func (c Config) GetTaskServers(task *Task, runFlags *core.RunFlags) ([]Server, error) {
	var servers []Server
	var err error
	// If any runtime target flags are used, disregard config specified task targets
	if len(runFlags.Servers) > 0 || len(runFlags.Tags) > 0 || runFlags.All {
		servers, err = c.FilterServers(runFlags.All, runFlags.Servers, runFlags.Tags)
	} else {
		servers, err = c.FilterServers(task.Target.All, task.Target.Servers, task.Target.Tags)
	}

	if err != nil {
		return []Server{}, err
	}

	return servers, nil
}

func (c Config) GetTasksByIDs(ids []string) ([]Task, error) {
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

func (c Config) GetTaskNames() []string {
	taskNames := []string{}
	for _, task := range c.Tasks {
		taskNames = append(taskNames, task.Name)
	}

	return taskNames
}

func (c Config) GetTaskIDAndDesc() []string {
	taskNames := []string{}
	for _, task := range c.Tasks {
		taskNames = append(taskNames, fmt.Sprintf("%s\t%s", task.ID, task.Desc))
	}

	return taskNames
}

func (c Config) GetTask(id string) (*Task, error) {
	for _, task := range c.Tasks {
		if id == task.ID {
			return &task, nil
		}
	}

	return nil, &core.TaskNotFound{IDs: []string{id}}
}
