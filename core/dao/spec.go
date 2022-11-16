package dao

import (
	// "errors"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/sake/core"
)

type Spec struct {
	Name              string `yaml:"_"`
	Desc              *string `yaml:"desc"`
	Strategy          string `yaml:"strategy"`
	Batch             uint32 `yaml:"batch"`
	BatchP			  uint8  `yaml:"batch_p"`
	Forks             uint32 `yaml:"forks"`
	Output            string `yaml:"output"`
	MaxFailPercentage uint8  `yaml:"max_fail_percentage"`
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

func (s Spec) GetValue(key string, _ int) string {
	lkey := strings.ToLower(key)
	switch lkey {
	case "name", "spec":
		return s.Name
	case "desc", "Desc":
		if s.Desc != nil {
			return *s.Desc
		} else {
			return ""
		}
	case "strategy":
		return s.Strategy
	case "forks":
		return strconv.Itoa(int(s.Forks))
	case "batch":
		return strconv.Itoa(int(s.Batch))
	case "batch_p":
		return strconv.Itoa(int(s.BatchP))
	case "output":
		return s.Output
	case "max_fail_percentage":
		return strconv.Itoa(int(s.MaxFailPercentage))
	case "any_errors_fatal":
		return strconv.FormatBool(s.AnyErrorsFatal)
	case "ignore_errors":
		return strconv.FormatBool(s.IgnoreErrors)
	case "ignore_unreachable":
		return strconv.FormatBool(s.IgnoreUnreachable)
	case "omit_empty":
		return strconv.FormatBool(s.OmitEmpty)
	default:
		return ""
	}
}

// ParseSpecsYAML parses the specs dictionary and returns it as a list.
func (c *ConfigYAML) ParseSpecsYAML() ([]Spec, []ResourceErrors[Spec]) {
	var specs []Spec
	count := len(c.Specs.Content)

	specErrors := []ResourceErrors[Spec]{}
	j := -1
	for i := 0; i < count; i += 2 {
		j += 1
		spec, serr := c.DecodeSpec(c.Specs.Content[i].Value, *c.Specs.Content[i+1])
		re := ResourceErrors[Spec]{Resource: spec, Errors: []error{}}
		specErrors = append(specErrors, re)

		for _, e := range serr {
			specErrors[j].Errors = append(specErrors[j].Errors, e)
		}
		specs = append(specs, *spec)
	}

	return specs, specErrors
}

func (c *ConfigYAML) DecodeSpec(name string, specYAML yaml.Node) (*Spec, []error) {
	spec := &Spec{
		Name: name,
		context:     c.Path,
		contextLine: specYAML.Line,
	}

	specErrors := []error{}
	err := specYAML.Decode(spec)
	if err != nil {
		specErrors = append(specErrors, err)
	}

	if spec.AnyErrorsFatal && spec.MaxFailPercentage > 0 {
		specErrors = append(specErrors, &core.MultipleFailSet{Name: name})
	}

	if spec.BatchP > 0 && spec.Batch > 0 {
		specErrors = append(specErrors, &core.BatchMultipleDef{Name: name})
	}

	if spec.BatchP > 100 {
		specErrors = append(specErrors, &core.InvalidPercentInput{Name: "batch_p"})
	}

	return spec, specErrors
}

func (c *Config) GetSpec(name string) (*Spec, error) {
	for _, spec := range c.Specs {
		if name == spec.Name {
			return &spec, nil
		}
	}

	return nil, &core.SpecNotFound{Name: name}
}

func (c *Config) GetSpecNames() []string {
	names := []string{}
	for _, spec := range c.Specs {
		if spec.Desc != nil {
			names = append(names, fmt.Sprintf("%s\t%s", spec.Name, *spec.Desc))
		} else {
			names = append(names, spec.Name)
		}
	}

	return names
}

func (c *Config) GetSpecsByName(names []string) ([]Spec, error) {
	if len(names) == 0 {
		return c.Specs, nil
	}

	foundSpecs := make(map[string]bool)
	for _, t := range names {
		foundSpecs[t] = false
	}

	var filteredSpecs []Spec
	for _, id := range names {
		if foundSpecs[id] {
			continue
		}

		for _, spec := range c.Specs {
			if id == spec.Name {
				foundSpecs[spec.Name] = true
				filteredSpecs = append(filteredSpecs, spec)
			}
		}
	}

	nonExistingSpecs := []string{}
	for k, v := range foundSpecs {
		if !v {
			nonExistingSpecs = append(nonExistingSpecs, k)
		}
	}

	if len(nonExistingSpecs) > 0 {
		return []Spec{}, &core.SpecsNotFound{Specs: nonExistingSpecs}
	}

	return filteredSpecs, nil
}
