package dao

import (
	"errors"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/sake/core"
)

type Spec struct {
	Name              string `yaml:"_"`
	Output            string `yaml:"output"`
	Parallel          bool   `yaml:"parallel"`
	AnyErrorsFatal    bool   `yaml:"any_errors_fatal"`
	IgnoreErrors      bool   `yaml:"ignore_errors"`
	IgnoreUnreachable bool   `yaml:"ignore_unreachable"`
	OmitEmpty         bool   `yaml:"omit_empty"`

	context     string // config path
	contextLine int    // defined at
}

func (s *Spec) GetContext() string {
	return s.context
}

func (s *Spec) GetContextLine() int {
	return s.contextLine
}

// ParseSpecsYAML parses the specs dictionary and returns it as a list.
func (c *ConfigYAML) ParseSpecsYAML() ([]Spec, []ResourceErrors[Spec]) {
	var specs []Spec
	count := len(c.Specs.Content)

	specErrors := []ResourceErrors[Spec]{}
	j := -1
	for i := 0; i < count; i += 2 {
		j += 1
		spec := &Spec{
			Name:        c.Specs.Content[i].Value,
			context:     c.Path,
			contextLine: c.Specs.Content[i].Line,
		}
		re := ResourceErrors[Spec]{Resource: spec, Errors: []error{}}
		specErrors = append(specErrors, re)

		err := c.Specs.Content[i+1].Decode(spec)
		if err != nil {
			for _, yerr := range err.(*yaml.TypeError).Errors {
				specErrors[j].Errors = append(specErrors[j].Errors, errors.New(yerr))
			}
			continue
		}

		specs = append(specs, *spec)
	}

	return specs, specErrors
}

func (c Config) GetSpec(name string) (*Spec, error) {
	for _, spec := range c.Specs {
		if name == spec.Name {
			return &spec, nil
		}
	}

	return nil, &core.SpecNotFound{Name: name}
}

func (c Config) GetSpecNames() []string {
	names := []string{}
	for _, spec := range c.Specs {
		names = append(names, spec.Name)
	}

	return names
}
