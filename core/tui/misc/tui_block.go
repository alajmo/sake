package misc

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// Block theme colors (mani-style)
var (
	BlockKeyColor       = "#5f87d7" // Blue for keys
	BlockSeparatorColor = "#5f87d7" // Blue for separators
	BlockValueColor     = ""        // Default for values
	BlockTrueColor      = "#00af5f" // Green for true
	BlockFalseColor     = "#d75f5f" // Red for false
)

// FormatKeyValue formats a key-value pair with tview color tags
func FormatKeyValue(padding bool, prefix, key, separator, value string, valueIsTrue *bool) string {
	var valueColor string
	if valueIsTrue != nil {
		if *valueIsTrue {
			valueColor = BlockTrueColor
		} else {
			valueColor = BlockFalseColor
		}
	} else {
		valueColor = BlockValueColor
	}

	keyStr := fmt.Sprintf("[%s:-:-]%s%s[-:-:-]", BlockKeyColor, prefix, key)
	sepStr := fmt.Sprintf("[%s:-:-]%s[-:-:-]", BlockSeparatorColor, separator)

	var valueStr string
	if valueColor != "" {
		valueStr = fmt.Sprintf("[%s:-:-]%s[-:-:-]", valueColor, value)
	} else {
		valueStr = value
	}

	str := fmt.Sprintf("%s%s %s\n", keyStr, sepStr, valueStr)
	if padding {
		return fmt.Sprintf("%4s%s", " ", str)
	}
	return str
}

// FormatCmd formats a command with indentation
func FormatCmd(cmd string) string {
	output := ""
	scanner := bufio.NewScanner(strings.NewReader(cmd))
	for scanner.Scan() {
		output += fmt.Sprintf("%4s%s\n", " ", scanner.Text())
	}
	return output
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}

// SubTaskInfo holds info about a sub-task for display
type SubTaskInfo struct {
	Name  string
	Desc  string
	Cmd   string
	Task  string // reference to another task
	IsRef bool   // true if this is a task reference
}

// FormatTaskBlock formats a task description in mani-style block format
func FormatTaskBlock(name, desc, cmd string, local bool, tty bool, attach bool, workDir string, shell string, envs []string, tags []string, subTasks []SubTaskInfo, spec, target, theme string) string {
	output := "\n"

	output += FormatKeyValue(false, "", "name", ":", name, nil)

	if desc != "" {
		output += FormatKeyValue(false, "", "description", ":", desc, nil)
	}

	output += FormatKeyValue(false, "", "local", ":", strconv.FormatBool(local), BoolPtr(local))
	output += FormatKeyValue(false, "", "tty", ":", strconv.FormatBool(tty), BoolPtr(tty))
	output += FormatKeyValue(false, "", "attach", ":", strconv.FormatBool(attach), BoolPtr(attach))

	if workDir != "" {
		output += FormatKeyValue(false, "", "work_dir", ":", workDir, nil)
	}

	if shell != "" {
		output += FormatKeyValue(false, "", "shell", ":", shell, nil)
	}

	if spec != "" {
		output += FormatKeyValue(false, "", "spec", ":", spec, nil)
	}

	if target != "" {
		output += FormatKeyValue(false, "", "target", ":", target, nil)
	}

	if theme != "" {
		output += FormatKeyValue(false, "", "theme", ":", theme, nil)
	}

	if len(tags) > 0 {
		output += FormatKeyValue(false, "", "tags", ":", strings.Join(tags, ", "), nil)
	}

	if len(envs) > 0 {
		output += FormatKeyValue(false, "", "env", ":", "", nil)
		for _, env := range envs {
			parts := strings.SplitN(strings.TrimSuffix(env, "\n"), "=", 2)
			if len(parts) == 2 {
				output += FormatKeyValue(true, "", parts[0], ":", parts[1], nil)
			}
		}
	}

	// Show cmd if it's a simple command task
	if cmd != "" {
		output += FormatKeyValue(false, "", "cmd", ":", "", nil)
		output += FormatCmd(cmd)
	}

	// Show sub-tasks/task references
	if len(subTasks) > 0 {
		output += FormatKeyValue(false, "", "tasks", ":", "", nil)
		for _, st := range subTasks {
			if st.IsRef && st.Task != "" {
				// Task reference
				if st.Name != "" {
					output += FormatKeyValue(true, "- ", st.Name, ":", fmt.Sprintf("task: %s", st.Task), nil)
				} else {
					output += FormatKeyValue(true, "- ", "task", ":", st.Task, nil)
				}
			} else if st.Cmd != "" {
				// Inline command
				if st.Name != "" {
					if st.Desc != "" {
						output += FormatKeyValue(true, "- ", st.Name, ":", st.Desc, nil)
					} else {
						output += FormatKeyValue(true, "- ", st.Name, ":", "(cmd)", nil)
					}
				} else {
					output += FormatKeyValue(true, "- ", "cmd", "", "", nil)
				}
			}
		}
	}

	output += "\n"
	return output
}

// FormatServerBlock formats a server description in mani-style block format
func FormatServerBlock(name, desc, host, user string, port uint16, local bool, tags []string, bastions []string, identityFile string, workDir string) string {
	output := "\n"

	output += FormatKeyValue(false, "", "name", ":", name, nil)

	if desc != "" {
		output += FormatKeyValue(false, "", "description", ":", desc, nil)
	}

	output += FormatKeyValue(false, "", "local", ":", strconv.FormatBool(local), BoolPtr(local))

	if !local {
		output += FormatKeyValue(false, "", "host", ":", host, nil)

		if user != "" {
			output += FormatKeyValue(false, "", "user", ":", user, nil)
		}

		output += FormatKeyValue(false, "", "port", ":", strconv.FormatUint(uint64(port), 10), nil)

		if identityFile != "" {
			output += FormatKeyValue(false, "", "identity_file", ":", identityFile, nil)
		}

		if len(bastions) > 0 {
			output += FormatKeyValue(false, "", "bastions", ":", strings.Join(bastions, ", "), nil)
		}
	}

	if workDir != "" {
		output += FormatKeyValue(false, "", "work_dir", ":", workDir, nil)
	}

	if len(tags) > 0 {
		output += FormatKeyValue(false, "", "tags", ":", strings.Join(tags, ", "), nil)
	}

	output += "\n"
	return output
}
