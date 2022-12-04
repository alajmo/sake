package dao

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
	"gopkg.in/yaml.v3"

	"github.com/alajmo/sake/core"
)

// Resource Errors

type ResourceErrors[T any] struct {
	Resource *T
	Errors   []error
}

type Resource interface {
	GetContext() string
	GetContextLine() int
}

// FormatErrors concats one or many errors attached to it for a single resource.
func FormatErrors(re Resource, errs []error) error {
	var msg = ""
	partsRe := regexp.MustCompile(`line (\d*): (.*)`)

	context := re.GetContext()

	var errPrefix = text.FgRed.Sprintf("error")
	var ptrPrefix = text.FgBlue.Sprintf("-->")
	for _, err := range errs {
		match := partsRe.FindStringSubmatch(err.Error())
		// In-case matching fails, return unformatted error
		if len(match) != 3 {
			contextLine := re.GetContextLine()

			if contextLine == -1 {
				msg = fmt.Sprintf("%s%s: %s\n  %s %s\n\n", msg, errPrefix, err, ptrPrefix, context)
			} else {
				msg = fmt.Sprintf("%s%s: %s\n  %s %s:%d\n\n", msg, errPrefix, err, ptrPrefix, context, contextLine)
			}
		} else {
			msg = fmt.Sprintf("%s%s: %s\n  %s %s:%s\n\n", msg, errPrefix, match[2], ptrPrefix, context, match[1])
		}
	}

	if msg != "" {
		return &core.ConfigErr{Msg: msg}
	}

	return nil
}

// ENV

func ParseNodeEnv(node yaml.Node) []string {
	var envs []string
	count := len(node.Content)

	for i := 0; i < count; i += 2 {
		env := fmt.Sprintf("%v=%v", node.Content[i].Value, node.Content[i+1].Value)
		envs = append(envs, env)
	}

	return envs
}

func EvaluateEnv(envList []string) ([]string, error) {
	var envs []string

	for _, arg := range envList {
		kv := strings.SplitN(arg, "=", 2)

		if strings.HasPrefix(kv[1], "$(") && strings.HasSuffix(kv[1], ")") {
			kv[1] = strings.TrimPrefix(kv[1], "$(")
			kv[1] = strings.TrimSuffix(kv[1], ")")

			cmd := exec.Command("bash", "-c", kv[1])
			cmd.Env = os.Environ()
			out, err := cmd.CombinedOutput()
			if err != nil {
				return envs, &core.ConfigEnvFailed{Name: kv[0], Err: string(out)}
			}

			envs = append(envs, fmt.Sprintf("%v=%v", kv[0], string(out)))
		} else {
			envs = append(envs, fmt.Sprintf("%v=%v", kv[0], kv[1]))
		}
	}

	return envs, nil
}

// MergeEnvs merges environment variables.
// Priority is from highest to lowest (1st env takes precedence over the last entry).
func MergeEnvs(envs ...[]string) []string {
	var mergedEnvs []string
	args := make(map[string]bool)

	for _, part := range envs {
		for _, elem := range part {
			if elem == "" {
				continue
			}

			elem = strings.TrimSuffix(elem, "\n")

			kv := strings.SplitN(elem, "=", 2)
			_, ok := args[kv[0]]

			if !ok {
				mergedEnvs = append(mergedEnvs, elem)
				args[kv[0]] = true
			}
		}
	}

	return mergedEnvs
}

// SelectFirstNonEmpty selects first non-empty string.
func SelectFirstNonEmpty(values ...string) string {
	for _, w := range values {
		if w != "" {
			return w
		}
	}

	return ""
}

func IsNullNode(node yaml.Node) bool {
	return node.Kind == 0
}

func CheckIsMappingNode(node yaml.Node) error {
	if node.Kind != 4 {
		return errors.New("expected mapping")
	}

	return nil
}

func CheckIsSequenceNode(node yaml.Node) error {
	if node.Kind != 2 {
		return errors.New("expected sequence")
	}

	return nil
}

func CheckIsScalarNode(node yaml.Node) error {
	if node.Kind != 8 {
		return errors.New("expected scalar")
	}

	return nil
}

func EvaluatePassword(password string) (string, error) {
	if strings.HasPrefix(password, "$(") && strings.HasSuffix(password, ")") {
		password = strings.TrimPrefix(password, "$(")
		password = strings.TrimSuffix(password, ")")

		cmd := exec.Command("bash", "-c", password)
		cmd.Env = os.Environ()
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", &core.PasswordEvalFailed{Err: string(out)}
		}

		password = strings.TrimSpace(string(out))
		return password, nil
	} else {
		return password, nil
	}
}
