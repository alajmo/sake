package dao

import (
	"errors"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/sake/core"
)

type Target struct {
	Name    string   `yaml:"name"`
	All     bool     `yaml:"all"`
	Servers []string `yaml:"servers"`
	Tags    []string `yaml:"tags"`

	context     string // config path
	contextLine int    // defined at
}

func (t *Target) GetContext() string {
	return t.context
}

func (t *Target) GetContextLine() int {
	return t.contextLine
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

		targets = append(targets, *target)
	}

	return targets, targetErrors
}

func (c Config) GetTarget(name string) (*Target, error) {
	for _, target := range c.Targets {
		if name == target.Name {
			return &target, nil
		}
	}

	return nil, &core.TargetNotFound{Name: name}
}

func (c Config) GetTargetNames() []string {
	names := []string{}
	for _, target := range c.Targets {
		names = append(names, target.Name)
	}

	return names
}
