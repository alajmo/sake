package dao

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jedib0t/go-pretty/v6/text"
	"gopkg.in/yaml.v3"

	"github.com/alajmo/sake/core"
)

var (
	ACCEPTABLE_FILE_NAMES = []string{"sake.yaml", "sake.yml", ".sake.yaml", ".sake.yml"}

	DEFAULT_THEME = Theme{
		Name:  "default",
		Table: DefaultTable,
		Text:  DefaultText,
	}

	DEFAULT_TARGET = Target{
		Name:    "default",
		All:     false,
		Servers: []string{},
		Tags:    []string{},
	}

	DEFAULT_SPEC = Spec{
		Name:              "default",
		Output:            "text",
		Parallel:          false,
		AnyErrorsFatal:    false,
		IgnoreUnreachable: false,
		IgnoreErrors:      false,
		OmitEmpty:         false,
	}
)

type Config struct {
	DisableVerifyHost bool
	KnownHostsFile    string
	Envs              []string
	Themes            []Theme
	Specs             []Spec
	Targets           []Target
	Servers           []Server
	Tasks             []Task
	Path              string
}

type ConfigYAML struct {
	// Internal
	Path           string  `yaml:"-"`
	Dir            string  `yaml:"-"`
	UserConfigFile *string `yaml:"-"`

	// Intermediate
	DisableVerifyHost *bool     `yaml:"disable_verify_host"`
	KnownHostsFile    *string   `yaml:"known_hosts_file"`
	Import            yaml.Node `yaml:"import"`
	Env               yaml.Node `yaml:"env"`
	Themes            yaml.Node `yaml:"themes"`
	Specs             yaml.Node `yaml:"specs"`
	Targets           yaml.Node `yaml:"targets"`
	Servers           yaml.Node `yaml:"servers"`
	Tasks             yaml.Node `yaml:"tasks"`

	contextLine int `yaml:"-"`
}

func (c *ConfigYAML) GetContext() string {
	return c.Path
}

func (c *ConfigYAML) GetContextLine() int {
	return c.contextLine
}

// Function to read sake configs.
func ReadConfig(configFilepath string, userConfigPath string, noColor bool) (Config, error) {
	CheckUserNoColor(noColor)
	var configPath string

	userConfigFile := getUserConfigFile(userConfigPath)

	// Try to find config file in current directory and all parents
	if configFilepath != "" {
		filename, err := filepath.Abs(configFilepath)
		if err != nil {
			return Config{}, err
		}

		configPath = filename
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return Config{}, err
		}

		// Check first cwd and all parent directories, then if not found,
		// check if env variable SAKE_CONFIG is set, and if not found
		// return no config found
		filename, err := core.FindFileInParentDirs(wd, ACCEPTABLE_FILE_NAMES)
		if err != nil {
			val, present := os.LookupEnv("SAKE_CONFIG")
			if present {
				filename = val
			} else {
				return Config{}, err
			}
		}

		filename, err = filepath.Abs(filename)
		if err != nil {
			return Config{}, err
		}

		configPath = filename
	}

	dat, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	// Found config, now try to read it
	var configYAML ConfigYAML

	configYAML.Path = configPath
	configYAML.Dir = filepath.Dir(configPath)
	configYAML.UserConfigFile = userConfigFile

	err = yaml.Unmarshal(dat, &configYAML)
	if err != nil {
		re := ResourceErrors[ConfigYAML]{Resource: &configYAML, Errors: []error{err}}
		return Config{}, FormatErrors(re.Resource, re.Errors)
	}

	config, configErr := configYAML.parseConfig()
	config.CheckConfigNoColor()

	if configErr != nil {
		return config, configErr
	}

	return config, nil
}

// Returns the config env list as a string splice in the form [key=value, key1=$(echo 123)]
func (c ConfigYAML) ParseEnvsYAML() []string {
	var envs []string
	count := len(c.Env.Content)
	for i := 0; i < count; i += 2 {
		env := fmt.Sprintf("%v=%v", c.Env.Content[i].Value, c.Env.Content[i+1].Value)
		envs = append(envs, env)
	}

	return envs
}

func getUserConfigFile(userConfigPath string) *string {
	// Flag
	if userConfigPath != "" {
		if _, err := os.Stat(userConfigPath); err == nil {
			return &userConfigPath
		}
	}

	// Env
	val, present := os.LookupEnv("SAKE_USER_CONFIG")
	if present {
		return &val
	}

	// Default
	defaultUserConfigDir, _ := os.UserConfigDir()
	defaultUserConfigPath := filepath.Join(defaultUserConfigDir, "sake", "config.yaml")
	if _, err := os.Stat(defaultUserConfigPath); err == nil {
		return &defaultUserConfigPath
	}

	return nil
}

// Open sake config in editor
func (c Config) EditConfig() error {
	return openEditor(c.Path, -1)
}

func openEditor(path string, lineNr int) error {
	editor, found := os.LookupEnv("EDITOR")
	if !found {
		return &core.NoEditorEnv{}
	}

	var args []string

	if lineNr > 0 {
		switch editor {
		case "vim":
			args = []string{fmt.Sprintf("+%v", lineNr), path}
		case "vi":
			args = []string{fmt.Sprintf("+%v", lineNr), path}
		case "emacs":
			args = []string{fmt.Sprintf("+%v", lineNr), path}
		case "nano":
			args = []string{fmt.Sprintf("+%v", lineNr), path}
		case "code": // visual studio code
			args = []string{"--goto", fmt.Sprintf("%s:%v", path, lineNr)}
		case "idea": // Intellij
			args = []string{"--line", fmt.Sprintf("%v", lineNr), path}
		case "subl": // Sublime
			args = []string{fmt.Sprintf("%s:%v", path, lineNr)}
		case "atom":
			args = []string{fmt.Sprintf("%s:%v", path, lineNr)}
		case "notepad-plus-plus":
			args = []string{"-n", fmt.Sprintf("%v", lineNr), path}
		default:
			args = []string{path}
		}
	} else {
		args = []string{path}
	}

	editorBin, err := exec.LookPath(editor)
	if err != nil {
		return err
	}

	cmd := exec.Command(editorBin, args...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Open sake config in editor and optionally go to line matching the task name
func (c Config) EditTask(name string) error {
	configPath := c.Path
	if name != "" {
		task, err := c.GetTask(name)
		if err != nil {
			return err
		}
		configPath = task.context
	}

	dat, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	type ConfigTmp struct {
		Tasks yaml.Node
	}

	var configTmp ConfigTmp
	err = yaml.Unmarshal(dat, &configTmp)
	if err != nil {
		return err
	}

	lineNr := 0
	if name == "" {
		lineNr = configTmp.Tasks.Line - 1
	} else {
		for _, task := range configTmp.Tasks.Content {
			if task.Value == name {
				lineNr = task.Line
				break
			}
		}
	}

	return openEditor(configPath, lineNr)
}

// Open sake config in editor and optionally go to line matching the server name
func (c Config) EditServer(name string) error {
	configPath := c.Path
	if name != "" {
		server, err := c.GetServer(name)
		if err != nil {
			return err
		}
		configPath = server.context
	}

	dat, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	type ConfigTmp struct {
		Servers yaml.Node
	}

	var configTmp ConfigTmp
	err = yaml.Unmarshal(dat, &configTmp)
	if err != nil {
		return err
	}

	lineNr := 0
	if name == "" {
		lineNr = configTmp.Servers.Line - 1
	} else {
		for _, server := range configTmp.Servers.Content {
			if server.Value == name {
				lineNr = server.Line
				break
			}
		}
	}

	return openEditor(configPath, lineNr)
}

func InitSake(args []string) ([]Server, error) {
	// Choose to initialize sake in a different directory
	// 1. absolute or
	// 2. relative or
	// 3. working directory
	var configDir string
	if len(args) > 0 && filepath.IsAbs(args[0]) {
		// absolute path
		configDir = args[0]
	} else if len(args) > 0 {
		// relative path
		wd, err := os.Getwd()
		if err != nil {
			return []Server{}, err
		}
		configDir = filepath.Join(wd, args[0])
	} else {
		// working directory
		wd, err := os.Getwd()
		if err != nil {
			return []Server{}, err
		}
		configDir = wd
	}

	err := os.MkdirAll(configDir, os.ModePerm)
	if err != nil {
		return []Server{}, err
	}

	configPath := filepath.Join(configDir, "sake.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return []Server{}, &core.AlreadySakeDirectory{Dir: configDir}
	}

	rootName := "localhost"
	rootHost := "0.0.0.0"
	rootServer := Server{Name: rootName, Host: rootHost}
	servers := []Server{rootServer}
	funcMap := template.FuncMap{
		"serverItem": func(name string, host string) string {
			var txt = name + ":"
			txt = txt + "\n    host: " + host
			txt = txt + "\n    local: true"
			return txt
		},
	}

	tmpl, err := template.New("init").Funcs(funcMap).Parse(`servers:
  {{- range .}}
  {{ (serverItem .Name .Host) }}
  {{ end }}
tasks:
  ping:
    desc: Pong
    cmd: echo "pong"
`,
	)
	if err != nil {
		return []Server{}, err
	}

	// Create sake.yaml
	f, err := os.Create(configPath)
	if err != nil {
		return []Server{}, err
	}

	err = tmpl.Execute(f, servers)
	if err != nil {
		return []Server{}, err
	}

	f.Close()

	fmt.Println("\nInitialized sake in", configDir)
	fmt.Println("- Created sake.yaml")

	return servers, nil
}

func CheckUserNoColor(noColorFlag bool) {
	_, present := os.LookupEnv("NO_COLOR")
	if noColorFlag || present {
		text.DisableColors()
	}
}

func (c *Config) CheckConfigNoColor() {
	for _, env := range c.Envs {
		name := strings.Split(env, "=")[0]
		if name == "NO_COLOR" {
			text.DisableColors()
		}
	}
}
