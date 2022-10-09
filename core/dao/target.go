package dao

import (
	"errors"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/sake/core"
)

type Target struct {
	Name    string   `yaml:"name"`
	All     bool     `yaml:"all"`
	Servers []string `yaml:"servers"`
	Tags    []string `yaml:"tags"`
	Regex   string   `yaml:"regex"`
	Invert  bool     `yaml:"invert"`
	Limit   uint32   `yaml:"limit"`
	LimitP  uint8    `yaml:"limit_p"`

	context     string // config path
	contextLine int    // defined at
}

func (t *Target) GetContext() string {
	return t.context
}

func (t *Target) GetContextLine() int {
	return t.contextLine
}

func (t Target) GetValue(key string, _ int) string {
	lkey := strings.ToLower(key)
	switch lkey {
	case "name", "target":
		return t.Name
	case "all":
		return strconv.FormatBool(t.All)
	case "servers":
		return strings.Join(t.Servers, "\n")
	case "tags":
		return strings.Join(t.Tags, "\n")
	case "regex":
		return t.Regex
	case "invert":
		return strconv.FormatBool(t.Invert)
	case "limit":
		return strconv.Itoa(int(t.Limit))
	case "limit_p":
		return strconv.Itoa(int(t.LimitP))
	default:
		return ""
	}
}

// ParseTargetsYAML parses the target dictionary and returns it as a list.
func (c *ConfigYAML) ParseTargetsYAML() ([]Target, []ResourceErrors[Target]) {
	var targets []Target
	count := len(c.Targets.Content)

	targetErrors := []ResourceErrors[Target]{}
	j := -1
	for i := 0; i < count; i += 2 {
		j += 1
		target := &Target{
			context:     c.Path,
			contextLine: c.Targets.Content[i].Line,
		}
		re := ResourceErrors[Target]{Resource: target, Errors: []error{}}
		targetErrors = append(targetErrors, re)

		err := c.Targets.Content[i+1].Decode(target)
		if err != nil {
			for _, yerr := range err.(*yaml.TypeError).Errors {
				targetErrors[j].Errors = append(targetErrors[j].Errors, errors.New(yerr))
			}
			continue
		}

		target.Name = c.Targets.Content[i].Value

		if target.LimitP > 100 {
			targetErrors[j].Errors = append(targetErrors[j].Errors, &core.InvalidPercentInput{})
		}

		targets = append(targets, *target)
	}

	return targets, targetErrors
}

func (c *Config) GetTarget(name string) (*Target, error) {
	for _, target := range c.Targets {
		if name == target.Name {
			return &target, nil
		}
	}

	return nil, &core.TargetNotFound{Name: name}
}

func (c *Config) GetTargetNames() []string {
	names := []string{}
	for _, target := range c.Targets {
		names = append(names, target.Name)
	}

	return names
}

func (c *Config) GetTargetsByName(names []string) ([]Target, error) {
	if len(names) == 0 {
		return c.Targets, nil
	}

	foundTargets := make(map[string]bool)
	for _, t := range names {
		foundTargets[t] = false
	}

	var filteredTargets []Target
	for _, id := range names {
		if foundTargets[id] {
			continue
		}

		for _, target := range c.Targets {
			if id == target.Name {
				foundTargets[target.Name] = true
				filteredTargets = append(filteredTargets, target)
			}
		}
	}

	nonExistingTargets := []string{}
	for k, v := range foundTargets {
		if !v {
			nonExistingTargets = append(nonExistingTargets, k)
		}
	}

	if len(nonExistingTargets) > 0 {
		return []Target{}, &core.TargetsNotFound{Targets: nonExistingTargets}
	}

	return filteredTargets, nil
}
