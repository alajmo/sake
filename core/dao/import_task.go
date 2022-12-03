package dao

import (
	"fmt"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/text"
)

type TaskNode struct {
	ID       string
	TaskRefs []TaskRef
	Visiting bool
}

type TaskLink struct {
	A TaskNode
	B TaskNode
}

// Perform a depth-first search for tasks
// The following nomenclature is used:
//
// tasks: <-- root context
//
//	b: <-- root task
//		tasks: <-- child context
//			- task: a <-- child task
//			  env:
//			    foo: bar
func dfsTask(task *Task, tn *TaskNode, tm map[string]*TaskNode, cycles *[]TaskLink, cr *ConfigResources) {
	tn.Visiting = true

	// For each task ref, create a node
	for i := range tn.TaskRefs {
		var tnn TaskNode
		v, exists := tm[tn.TaskRefs[i].Task]
		if exists {
			tnn = *v
		} else {
			tnn = TaskNode{
				ID:       tn.TaskRefs[i].Task,
				Visiting: false,
			}
			tm[tnn.ID] = &tnn
		}

		// Found cyclic dependency
		if tnn.Visiting {
			c := TaskLink{
				A: *tn,
				B: tnn,
			}

			*cycles = append(*cycles, c)
			break
		}

		if tn.TaskRefs[i].Cmd != "" {
			// name: <name> <-- task
			// tasks:
			//   - cmd: <cmd> <-- tn.TaskRefs[i].Cmd

			local := task.Local
			if tn.TaskRefs[i].Local != nil {
				local = *tn.TaskRefs[i].Local
			}

			tty := task.TTY
			if tn.TaskRefs[i].TTY != nil {
				tty = *tn.TaskRefs[i].TTY
			}

			ignoreErrors := task.Spec.IgnoreErrors
			if tn.TaskRefs[i].IgnoreErrors != nil {
				ignoreErrors = *tn.TaskRefs[i].IgnoreErrors
			}

			envs := MergeEnvs(tn.TaskRefs[i].Envs, task.Envs)

			workDir := SelectFirstNonEmpty(tn.TaskRefs[i].WorkDir, task.WorkDir)
			shell := SelectFirstNonEmpty(tn.TaskRefs[i].Shell, task.Shell)

			childTask := TaskCmd{
				ID:           tn.TaskRefs[i].Task,
				Name:         tn.TaskRefs[i].Name,
				Desc:         tn.TaskRefs[i].Desc,
				Register:     tn.TaskRefs[i].Register,
				RootDir:      filepath.Dir(task.context),
				WorkDir:      workDir,
				Shell:        shell,
				Cmd:          tn.TaskRefs[i].Cmd,
				Envs:         envs,
				Local:        local,
				TTY:          tty,
				IgnoreErrors: ignoreErrors,
			}
			task.Tasks = append(task.Tasks, childTask)
		} else {
			// Reference command
			// tasks:
			//   a:
			//     ...
			//   b: <-- tn
			//     tasks:
			//       - task: a <-- tn.TaskRefs[i].Task

			childTask, err := cr.GetTask(tn.TaskRefs[i].Task)

			if err != nil {
				taskError := ResourceErrors[Task]{Resource: task, Errors: []error{err}}
				cr.TaskErrors = append(cr.TaskErrors, taskError)
				continue
			}

			if childTask.Cmd != "" {
				// tasks:
				//   a:
				//     cmd: <cmd>
				//   b:
				//     tasks:
				//       - task: a

				name := childTask.Name
				if tn.TaskRefs[i].Name != "" {
					name = tn.TaskRefs[i].Name
				}

				local := childTask.Local
				if tn.TaskRefs[i].Local != nil {
					local = *tn.TaskRefs[i].Local
				}

				tty := childTask.TTY
				if tn.TaskRefs[i].TTY != nil {
					tty = *tn.TaskRefs[i].TTY
				}

				ignoreErrors := childTask.Spec.IgnoreErrors
				if tn.TaskRefs[i].IgnoreErrors != nil {
					ignoreErrors = *tn.TaskRefs[i].IgnoreErrors
				}

				envs := MergeEnvs(tn.TaskRefs[i].Envs, task.Envs, childTask.Envs)

				workDir := SelectFirstNonEmpty(tn.TaskRefs[i].WorkDir, task.WorkDir, childTask.WorkDir)
				shell := SelectFirstNonEmpty(tn.TaskRefs[i].Shell, task.Shell, childTask.Shell)

				// TODO: Should task.Register be set here?
				t := TaskCmd{
					ID:           childTask.ID,
					Name:         name,
					Desc:         childTask.Desc,
					RootDir:      filepath.Dir(task.context),
					WorkDir:      workDir,
					Shell:        shell,
					Cmd:          childTask.Cmd,
					Register:     tn.TaskRefs[i].Register,
					Envs:         envs,
					Local:        local,
					TTY:          tty,
					IgnoreErrors: ignoreErrors,
				}
				task.Tasks = append(task.Tasks, t)
			} else {
				// tasks:
				//   a:
				//     tasks:
				//       - task: d
				//   b:
				//     tasks:
				//       - task: a

				// Append new task refs to the tnn node and traverse those task refs
				// Make sure it's a copy and not reference since we may traverse
				// the same task in the same context and we don't want env variables
				// to be populated from previous traversals.
				tnn.TaskRefs = []TaskRef{}
				for j, k := range childTask.TaskRefs {
					// Environment variable references:
					// a: <-- referenced node
					//   env: <-- childTask.Envs, takes last precedence
					//     hello: world
					//   tasks:
					//     - task: a
					//       env: <-- tnn.TaskRefs[j].Envs, takes second precedence
					//         bar: bar
					// b: <-- current node / traversing node
					//   env: <-- this env will be passed at the end so we don't have to include it now
					//     xyz: xyz
					//   tasks:
					//     - task: a
					//       env: <-- tn.TaskRefs[i].Envs, takes first precedence
					//         foo: foo

					// TODO: May need to add IgnoreErrors here
					tnn.TaskRefs = append(tnn.TaskRefs, k)
					tnn.TaskRefs[j].Envs = MergeEnvs(tn.TaskRefs[i].Envs, tnn.TaskRefs[j].Envs, childTask.Envs)
					tnn.TaskRefs[j].WorkDir = SelectFirstNonEmpty(tn.TaskRefs[i].WorkDir, tnn.TaskRefs[j].WorkDir, childTask.WorkDir)
					tnn.TaskRefs[j].Shell = SelectFirstNonEmpty(tn.TaskRefs[i].Shell, tnn.TaskRefs[j].Shell, childTask.Shell)
				}

				dfsTask(task, &tnn, tm, cycles, cr)
			}
		}
	}

	tn.Visiting = false
}

type FoundCyclicTaskDependency struct {
	Cycles []TaskLink
}

func (c *FoundCyclicTaskDependency) Error() string {
	var msg string

	var errPrefix = text.FgRed.Sprintf("error")
	var ptrPrefix = text.FgBlue.Sprintf("-->")
	msg = fmt.Sprintf("%s: %s\n", errPrefix, "found direct or indirect circular dependency")
	for i := range c.Cycles {
		msg += fmt.Sprintf("  %s %s\n      %s\n", ptrPrefix, c.Cycles[i].A.ID, c.Cycles[i].B.ID)
	}

	return msg
}
