package dao

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alajmo/sake/core"
	"github.com/jedib0t/go-pretty/v6/text"
	"gopkg.in/yaml.v3"
)

type Import struct {
	Path string

	context     string
	contextLine int
}

func (i *Import) GetContext() string {
	return i.context
}

func (i *Import) GetContextLine() int {
	return i.contextLine
}

// Used for config imports
type ConfigResources struct {
	DisableVerifyHost *bool
	DefaultTimeout    *uint
	KnownHostsFile    *string
	Shell             string
	Imports           []Import
	Themes            []Theme
	Specs             []Spec
	Targets           []Target
	Tasks             []Task
	Servers           []Server
	Envs              []string

	ConfigErrors []ResourceErrors[ConfigYAML]
	ImportErrors []ResourceErrors[Import]
	ThemeErrors  []ResourceErrors[Theme]
	SpecErrors   []ResourceErrors[Spec]
	TargetErrors []ResourceErrors[Target]
	TaskErrors   []ResourceErrors[Task]
	ServerErrors []ResourceErrors[Server]
}

type Node struct {
	Path     string
	Imports  []Import
	Visiting bool
	Visited  bool
}

type NodeLink struct {
	A Node
	B Node
}

type FoundCyclicDependency struct {
	Cycles []NodeLink
}

func (c *FoundCyclicDependency) Error() string {
	var msg string

	var errPrefix = text.FgRed.Sprintf("error")
	var ptrPrefix = text.FgBlue.Sprintf("-->")
	msg = fmt.Sprintf("%s: %s\n", errPrefix, "Found direct or indirect circular dependency")
	for i := range c.Cycles {
		msg += fmt.Sprintf("  %s %s\n      %s\n", ptrPrefix, c.Cycles[i].A.Path, c.Cycles[i].B.Path)
	}

	return msg
}

// parseConfig is the starting point of reading and importing sake config files.
// The following summary represents the logic:
// 1. Initialize the root config and optionally, if it exists, add the user config (XDG_HOME_CONFIG/sake/config.yaml) to the imports list of the root config
// 2. Perform a depth-first search of imports and collect all specs, targets, themes, tasks, servers and store in intermediate struct ConfigResources
//   - Nested tasks for tasks are saved as TaskRefYAML
//   - Spec, Theme, Target are saved as references here as well, if they are specified
//     3. If the default theme, spec, and target objects are not overwritten, then create them
//     3.1. Create default Theme collection
//     3.2. Create default Spec collection
//     3.3. Create default Target collection
//     4. Perform a depth-first search for task references and save them as T
//     5. We check duplicate server hosts in the config collection
//
// Given config imports, use a Depth-first-search algorithm to recursively
// check for resources (tasks, servers, dirs, themes, specs, targets).
// A struct is passed around that is populated with resources from each config.
// In case a cyclic dependency is found (a -> b and b -> a), we return early
// with an error containing the cyclic dependency found.
func (c *ConfigYAML) parseConfig() (Config, error) {
	// Main config
	cr := ConfigResources{}

	cr.Envs = []string{
		fmt.Sprintf("SAKE_DIR=%s", c.Dir),
		fmt.Sprintf("SAKE_PATH=%s", c.Path),
	}

	if !IsNullNode(c.Import) {
		err := CheckIsSequenceNode(c.Import)
		if err != nil {
			cfg := *c
			cfg.contextLine = c.Import.Line
			configError := ResourceErrors[ConfigYAML]{
				Resource: &cfg,
				Errors:   []error{err},
			}
			cr.ConfigErrors = append(cr.ConfigErrors, configError)
		} else {
			imports, importErrors := c.ParseImportsYAML()
			cr.ImportErrors = append(cr.ImportErrors, importErrors...)

			cr.Imports = append(cr.Imports, imports...)
		}
	}

	c.loadResources(&cr)

	// Check if there's a user config file ($XDG_HOME_CONFIG/sake/config.yaml)
	if c.UserConfigFile != nil {
		cr.Imports = append(cr.Imports, Import{Path: *c.UserConfigFile, context: c.Path, contextLine: -1})
	}

	// Import sake configs and check cyclic dependencies
	n := Node{
		Path:    c.Path,
		Imports: cr.Imports,
	}
	m := make(map[string]*Node)
	m[n.Path] = &n
	importCycles := []NodeLink{}
	dfsImport(&n, m, &importCycles, &cr)

	// Create default config if not exists
	_, err := cr.GetTheme(DEFAULT_THEME.Name)
	if err != nil {
		cr.Themes = append(cr.Themes, DEFAULT_THEME)
	}

	// Create default spec if not exists
	_, err = cr.GetSpec(DEFAULT_SPEC.Name)
	if err != nil {
		// TODO: Fill in all default values for spec
		cr.Specs = append(cr.Specs, DEFAULT_SPEC)
	}

	// Create default spec if not exists
	_, err = cr.GetTarget(DEFAULT_TARGET.Name)
	if err != nil {
		cr.Targets = append(cr.Targets, DEFAULT_TARGET)
	}

	// Process tasks:
	//  - Expand references (targets, specs, themes, tasks)
	//  - Check for cyclic dependencies for tasks
	taskCycles := []TaskLink{}
	for i := range cr.Tasks {
		if cr.Tasks[i].ThemeRef != "" {
			theme, err := cr.GetTheme(cr.Tasks[i].ThemeRef)
			if err != nil {
				cr.TaskErrors[i].Errors = append(cr.TaskErrors[i].Errors, err)
			} else {
				cr.Tasks[i].Theme = *theme
			}
		}

		if cr.Tasks[i].SpecRef != "" {
			spec, err := cr.GetSpec(cr.Tasks[i].SpecRef)
			if err != nil {
				cr.TaskErrors[i].Errors = append(cr.TaskErrors[i].Errors, err)
			} else {
				cr.Tasks[i].Spec = *spec
			}
		}

		if cr.Tasks[i].TargetRef != "" {
			target, err := cr.GetTarget(cr.Tasks[i].TargetRef)
			if err != nil {
				cr.TaskErrors[i].Errors = append(cr.TaskErrors[i].Errors, err)
			} else {
				cr.Tasks[i].Target = *target
			}
		}

		if cr.Tasks[i].Cmd != "" {
			taskCmd := TaskCmd{
				ID:      cr.Tasks[i].ID,
				Name:    cr.Tasks[i].Name,
				Desc:    cr.Tasks[i].Desc,
				RootDir: filepath.Dir(cr.Tasks[i].context),
				WorkDir: cr.Tasks[i].WorkDir,
				Cmd:     cr.Tasks[i].Cmd,
				Local:   cr.Tasks[i].Local,
				Shell:   cr.Tasks[i].Shell,
				TTY:     cr.Tasks[i].TTY,
				Envs:    cr.Tasks[i].Envs,
			}
			cr.Tasks[i].Tasks = append(cr.Tasks[i].Tasks, taskCmd)
		} else {
			tn := TaskNode{
				ID:       cr.Tasks[i].ID,
				TaskRefs: cr.Tasks[i].TaskRefs,
				Visiting: false,
			}
			tm := make(map[string]*TaskNode)
			tm[tn.ID] = &tn
			dfsTask(&cr.Tasks[i], &tn, tm, &taskCycles, &cr)
		}
	}

	// Create config
	var config = Config{
		Tasks:   cr.Tasks,
		Servers: cr.Servers,
		Themes:  cr.Themes,
		Specs:   cr.Specs,
		Targets: cr.Targets,
		Envs:    cr.Envs,
		Path:    c.Path,
	}

	if cr.DisableVerifyHost == nil {
		config.DisableVerifyHost = false
	} else {
		config.DisableVerifyHost = *cr.DisableVerifyHost
	}

	if cr.DefaultTimeout == nil {
		config.DefaultTimeout = DEFAULT_TIMEOUT
	} else {
		config.DefaultTimeout = *cr.DefaultTimeout
	}

	if cr.Shell != "" {
		config.Shell = cr.Shell
	}

	if cr.KnownHostsFile == nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return config, err
		}

		knownHostsFile := filepath.Join(home, "/.ssh/known_hosts")
		config.KnownHostsFile = knownHostsFile
	} else {
		config.KnownHostsFile = *cr.KnownHostsFile
	}

	// Check duplicate imports
	importErr := checkDuplicateImports(cr.Imports)

	duplicateObjects := checkDuplicateObjects(config)

	// Concat errors
	errString := concatErrors(importErr, duplicateObjects, cr, &importCycles, &taskCycles)

	if errString != "" {
		return config, &core.ConfigErr{Msg: errString}
	}

	return config, nil
}

func concatErrors(importErr string, duplicateObjects string, cr ConfigResources, importCycles *[]NodeLink, taskCycles *[]TaskLink) string {
	errString := importErr
	errString += duplicateObjects

	if len(*importCycles) > 0 {
		err := &FoundCyclicDependency{Cycles: *importCycles}
		errString = fmt.Sprintf("%s%s\n", errString, err.Error())
	}

	if len(*taskCycles) > 0 {
		err := &FoundCyclicTaskDependency{Cycles: *taskCycles}
		errString = fmt.Sprintf("%s%s\n", errString, err.Error())
	}

	for _, cfg := range cr.ConfigErrors {
		if len(cfg.Errors) > 0 {
			errString = fmt.Sprintf("%s%s", errString, FormatErrors(cfg.Resource, cfg.Errors))
		}
	}

	for _, imp := range cr.ImportErrors {
		if len(imp.Errors) > 0 {
			errString = fmt.Sprintf("%s%s", errString, FormatErrors(imp.Resource, imp.Errors))
		}
	}

	for _, server := range cr.ServerErrors {
		if len(server.Errors) > 0 {
			errString = fmt.Sprintf("%s%s", errString, FormatErrors(server.Resource, server.Errors))
		}
	}

	for _, theme := range cr.ThemeErrors {
		if len(theme.Errors) > 0 {
			errString = fmt.Sprintf("%s%s", errString, FormatErrors(theme.Resource, theme.Errors))
		}
	}

	for _, spec := range cr.SpecErrors {
		if len(spec.Errors) > 0 {
			errString = fmt.Sprintf("%s%s", errString, FormatErrors(spec.Resource, spec.Errors))
		}
	}

	for _, target := range cr.TargetErrors {
		if len(target.Errors) > 0 {
			errString = fmt.Sprintf("%s%s", errString, FormatErrors(target.Resource, target.Errors))
		}
	}

	for _, task := range cr.TaskErrors {
		if len(task.Errors) > 0 {
			errString = fmt.Sprintf("%s%s", errString, FormatErrors(task.Resource, task.Errors))
		}
	}

	return errString
}

func parseConfigFile(path string) (ConfigYAML, error) {
	var configYAML ConfigYAML

	absPath, err := filepath.Abs(path)
	if err != nil {
		configYAML.Path = path
		configYAML.Dir = filepath.Dir(absPath)

		return configYAML, &core.FileError{Err: err.Error()}
	}

	configYAML.Path = absPath
	configYAML.Dir = filepath.Dir(absPath)

	dat, err := os.ReadFile(absPath)
	if err != nil {
		return configYAML, &core.FileError{Err: err.Error()}
	}

	err = yaml.Unmarshal(dat, &configYAML)
	if err != nil {
		return configYAML, err
	}

	return configYAML, nil
}

func (c *ConfigYAML) loadResources(cr *ConfigResources) {
	if c.Shell != "" {
		cr.Shell = c.Shell
	}

	if c.DefaultTimeout != nil {
		cr.DefaultTimeout = c.DefaultTimeout
	}

	if c.DisableVerifyHost != nil {
		cr.DisableVerifyHost = c.DisableVerifyHost
	}

	if c.KnownHostsFile != nil {
		knownHostsFile := os.ExpandEnv(*c.KnownHostsFile)
		if strings.HasPrefix(knownHostsFile, "~/") {
			// Expand tilde ~
			home, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}

			knownHostsFile = filepath.Join(home, knownHostsFile[2:])
			cr.KnownHostsFile = &knownHostsFile
		} else {
			// Absolute filepath
			if filepath.IsAbs(knownHostsFile) {
				cr.KnownHostsFile = &knownHostsFile
			} else {
				// Relative filepath
				if *c.KnownHostsFile != "" {
					knownHostsFile = filepath.Join(c.Dir, knownHostsFile)
					cr.KnownHostsFile = &knownHostsFile
				} else {
					cr.KnownHostsFile = c.KnownHostsFile
				}
			}
		}
	}

	// Tasks
	if !IsNullNode(c.Tasks) {
		err := CheckIsMappingNode(c.Tasks)
		if err != nil {
			cfg := *c
			cfg.contextLine = c.Tasks.Line
			configError := ResourceErrors[ConfigYAML]{
				Resource: &cfg,
				Errors:   []error{err},
			}
			cr.ConfigErrors = append(cr.ConfigErrors, configError)
		} else {
			tasks, taskErrors := c.ParseTasksYAML()
			cr.TaskErrors = append(cr.TaskErrors, taskErrors...)
			cr.Tasks = append(cr.Tasks, tasks...)
		}
	}

	// Servers
	if !IsNullNode(c.Servers) {
		err := CheckIsMappingNode(c.Servers)
		if err != nil {
			cfg := *c
			cfg.contextLine = c.Servers.Line
			configError := ResourceErrors[ConfigYAML]{
				Resource: &cfg,
				Errors:   []error{err},
			}
			cr.ConfigErrors = append(cr.ConfigErrors, configError)
		} else {
			servers, serverErrors := c.ParseServersYAML()
			cr.Servers = append(cr.Servers, servers...)
			cr.ServerErrors = append(cr.ServerErrors, serverErrors...)
		}
	}

	// Themes
	if !IsNullNode(c.Themes) {
		err := CheckIsMappingNode(c.Themes)
		if err != nil {
			cfg := *c
			cfg.contextLine = c.Themes.Line
			configError := ResourceErrors[ConfigYAML]{
				Resource: &cfg,
				Errors:   []error{err},
			}
			cr.ConfigErrors = append(cr.ConfigErrors, configError)
		} else {
			themes, themeErrors := c.ParseThemesYAML()
			cr.Themes = append(cr.Themes, themes...)
			cr.ThemeErrors = append(cr.ThemeErrors, themeErrors...)
		}
	}

	// Specs
	if !IsNullNode(c.Specs) {
		err := CheckIsMappingNode(c.Specs)
		if err != nil {
			cfg := *c
			cfg.contextLine = c.Specs.Line
			configError := ResourceErrors[ConfigYAML]{
				Resource: &cfg,
				Errors:   []error{err},
			}
			cr.ConfigErrors = append(cr.ConfigErrors, configError)
		} else {
			specs, specErrors := c.ParseSpecsYAML()
			cr.Specs = append(cr.Specs, specs...)
			cr.SpecErrors = append(cr.SpecErrors, specErrors...)
		}
	}

	// Targets
	if !IsNullNode(c.Targets) {
		err := CheckIsMappingNode(c.Targets)
		if err != nil {
			cfg := *c
			cfg.contextLine = c.Targets.Line
			configError := ResourceErrors[ConfigYAML]{
				Resource: &cfg,
				Errors:   []error{err},
			}
			cr.ConfigErrors = append(cr.ConfigErrors, configError)
		} else {
			targets, targetErrors := c.ParseTargetsYAML()
			cr.Targets = append(cr.Targets, targets...)
			cr.TargetErrors = append(cr.TargetErrors, targetErrors...)
		}
	}

	// Envs
	if !IsNullNode(c.Env) {
		err := CheckIsMappingNode(c.Env)
		if err != nil {
			cfg := *c
			cfg.contextLine = c.Env.Line
			configError := ResourceErrors[ConfigYAML]{
				Resource: &cfg,
				Errors:   []error{err},
			}
			cr.ConfigErrors = append(cr.ConfigErrors, configError)
		} else {
			envs := ParseNodeEnv(c.Env)
			cr.Envs = append(cr.Envs, envs...)
		}
	}
}

func (c *ConfigYAML) ParseImportsYAML() ([]Import, []ResourceErrors[Import]) {
	var imports []Import
	count := len(c.Import.Content)

	importErrors := []ResourceErrors[Import]{}
	for i := 0; i < count; i += 1 {
		imp := &Import{
			Path:        c.Import.Content[i].Value,
			context:     c.Path,
			contextLine: c.Import.Content[i].Line,
		}
		re := ResourceErrors[Import]{Resource: imp, Errors: []error{}}
		importErrors = append(importErrors, re)

		err := CheckIsScalarNode(*c.Import.Content[i])
		if err != nil {
			importErrors[i].Errors = append(importErrors[i].Errors, err)
			continue
		}

		imports = append(imports, *imp)
	}

	return imports, importErrors
}

func dfsImport(n *Node, m map[string]*Node, cycles *[]NodeLink, cr *ConfigResources) {
	n.Visiting = true

	for i := range n.Imports {
		p, err := core.GetAbsolutePath(filepath.Dir(n.Path), n.Imports[i].Path, "")
		if err != nil {
			importError := ResourceErrors[Import]{Resource: &n.Imports[i], Errors: core.StringsToErrors(err.(*yaml.TypeError).Errors)}
			cr.ImportErrors = append(cr.ImportErrors, importError)
			continue
		}

		// Skip visited nodes
		var nc Node
		v, exists := m[p]
		if exists {
			nc = *v
		} else {
			nc = Node{Path: p}
			m[nc.Path] = &nc
		}

		if nc.Visited {
			continue
		}

		// Found cyclic dependency
		if nc.Visiting {
			c := NodeLink{
				A: *n,
				B: nc,
			}

			*cycles = append(*cycles, c)
			break
		}

		// Import raw configYAML
		configYAML, err := parseConfigFile(nc.Path)

		// Error belongs to config file trying to import the new config
		if err != nil {
			switch err.(type) {
			case *core.FileError:
				importError := ResourceErrors[Import]{Resource: &n.Imports[i], Errors: []error{err}}
				cr.ImportErrors = append(cr.ImportErrors, importError)
			default:
				configError := ResourceErrors[ConfigYAML]{Resource: &configYAML, Errors: []error{err}}
				cr.ConfigErrors = append(cr.ConfigErrors, configError)
			}

			continue
		}

		// Error belongs to the newly imported config file
		if !IsNullNode(configYAML.Import) {
			err := CheckIsSequenceNode(configYAML.Import)
			if err != nil {
				configError := ResourceErrors[ConfigYAML]{Resource: &configYAML, Errors: []error{err}}
				cr.ConfigErrors = append(cr.ConfigErrors, configError)
				continue
			} else {
				imports, importErrors := configYAML.ParseImportsYAML()
				cr.ImportErrors = append(cr.ImportErrors, importErrors...)
				nc.Imports = imports
			}
		}

		// Load resources from the config
		configYAML.loadResources(cr)

		dfsImport(&nc, m, cycles, cr)
	}

	n.Visiting = false
	n.Visited = true
}

type FoundDuplicateImports struct {
	imports []string
}

func (c *FoundDuplicateImports) Error() string {
	var msg string

	var errPrefix = text.FgRed.Sprintf("error")
	var ptrPrefix = text.FgBlue.Sprintf("-->")
	msg = fmt.Sprintf("%s: %s\n  %s", errPrefix, "found duplicate imports", ptrPrefix)
	msg += fmt.Sprintf(" %s\n", c.imports[0])
	for i, s := range c.imports[1:] {
		if i < len(c.imports[1:])-1 {
			msg += fmt.Sprintf("      %s\n", s)
		} else {
			msg += fmt.Sprintf("      %s", s)
		}
	}

	return msg
}

func checkDuplicateImports(imports []Import) string {
	paths := []string{}
	for _, p := range imports {
		paths = append(paths, p.Path)
	}

	duplicates := []string{}
	visited := make(map[string]bool, 0)
	for _, p := range paths {
		_, exists := visited[p]
		if exists && !core.StringInSlice(p, duplicates) {
			duplicates = append(duplicates, p)
		} else {
			visited[p] = true
		}
	}

	var errString string
	if len(duplicates) > 0 {
		err := &FoundDuplicateImports{imports: duplicates}
		errString = fmt.Sprintf("%s%s\n", errString, err.Error())
	}

	return errString
}

type FoundDuplicateObjects struct {
	Name   string
	Type   string
	Values []string
}

func (c *FoundDuplicateObjects) Error() string {
	var msg string

	var errPrefix = text.FgRed.Sprintf("error")
	var ptrPrefix = text.FgBlue.Sprintf("-->")
	msg = fmt.Sprintf("%s: %s %s %s\n  %s", errPrefix, "found duplicate", c.Type, c.Name, ptrPrefix)
	msg += fmt.Sprintf(" %s\n", c.Values[0])
	for i, s := range c.Values[1:] {
		if i < len(c.Values[1:])-1 {
			msg += fmt.Sprintf("      %s\n", s)
		} else {
			msg += fmt.Sprintf("      %s", s)
		}
	}

	return msg
}

func checkDuplicateObjects(config Config) string {
	// Task
	taskIDS := []string{}
	visitedTasks := make(map[string]bool, 0)
	tasks := make(map[string][]string, 0)
	for _, t := range config.Tasks {
		tasks[t.ID] = append(tasks[t.ID], t.context)
		_, exists := visitedTasks[t.ID]
		if !exists {
			taskIDS = append(taskIDS, t.ID)
			visitedTasks[t.ID] = true
		}
	}

	var errString string
	for _, id := range taskIDS {
		if len(tasks[id]) > 1 {
			err := &FoundDuplicateObjects{Name: id, Type: "task", Values: tasks[id]}
			errString = fmt.Sprintf("%s%s\n\n", errString, err.Error())
		}
	}

	// Spec
	specIDS := []string{}
	visitedSpecs := make(map[string]bool, 0)
	specs := make(map[string][]string, 0)
	for _, s := range config.Specs {
		specs[s.Name] = append(specs[s.Name], s.context)
		_, exists := visitedSpecs[s.Name]
		if !exists {
			specIDS = append(specIDS, s.Name)
			visitedSpecs[s.Name] = true
		}
	}

	for _, id := range specIDS {
		if len(specs[id]) > 1 {
			err := &FoundDuplicateObjects{Name: id, Type: "spec", Values: specs[id]}
			errString = fmt.Sprintf("%s%s\n\n", errString, err.Error())
		}
	}

	// Target
	targetIDS := []string{}
	visitedTargets := make(map[string]bool, 0)
	targets := make(map[string][]string, 0)
	for _, t := range config.Targets {
		targets[t.Name] = append(targets[t.Name], t.context)
		_, exists := visitedTargets[t.Name]
		if !exists {
			targetIDS = append(targetIDS, t.Name)
			visitedTargets[t.Name] = true
		}
	}

	for _, id := range targetIDS {
		if len(targets[id]) > 1 {
			err := &FoundDuplicateObjects{Name: id, Type: "target", Values: targets[id]}
			errString = fmt.Sprintf("%s%s\n\n", errString, err.Error())
		}
	}

	// Theme
	themeIDS := []string{}
	visitedThemes := make(map[string]bool, 0)
	themes := make(map[string][]string, 0)
	for _, t := range config.Themes {
		themes[t.Name] = append(themes[t.Name], t.context)
		_, exists := visitedThemes[t.Name]
		if !exists {
			themeIDS = append(themeIDS, t.Name)
			visitedThemes[t.Name] = true
		}
	}

	for _, id := range themeIDS {
		if len(themes[id]) > 1 {
			err := &FoundDuplicateObjects{Name: id, Type: "theme", Values: themes[id]}
			errString = fmt.Sprintf("%s%s\n\n", errString, err.Error())
		}
	}

	return errString
}

// Used for config imports
type TaskResources struct {
	Tasks      []Task
	TaskErrors []ResourceErrors[Task]
}

func (c *ConfigResources) GetTask(id string) (*Task, error) {
	for _, task := range c.Tasks {
		if id == task.ID {
			return &task, nil
		}
	}

	return nil, &core.TaskNotFound{IDs: []string{id}}
}

func (c *ConfigResources) GetTheme(name string) (*Theme, error) {
	for _, theme := range c.Themes {
		if name == theme.Name {
			return &theme, nil
		}
	}

	return nil, &core.ThemeNotFound{Name: name}
}

func (c *ConfigResources) GetSpec(name string) (*Spec, error) {
	for _, spec := range c.Specs {
		if name == spec.Name {
			return &spec, nil
		}
	}

	return nil, &core.SpecNotFound{Name: name}
}

func (c *ConfigResources) GetTarget(name string) (*Target, error) {
	for _, target := range c.Targets {
		if name == target.Name {
			return &target, nil
		}
	}

	return nil, &core.TargetNotFound{Name: name}
}
