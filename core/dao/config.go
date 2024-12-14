package dao

import (
	"fmt"
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

	DEFAULT_SHELL = "bash -c"

	DEFAULT_TIMEOUT = uint(20)

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
		Regex:   "",
		Invert:  false,
		Limit:   0,
		LimitP:  0,
	}

	DEFAULT_SPEC = Spec{
		Name:              "default",
		Desc:              "the default spec",
		Describe:          false,
		ListHosts:         false,
		Order:             "inventory",
		Silent:            false,
		Hidden:            false,
		Strategy:          "linear",
		Batch:             0,
		BatchP:            0,
		Forks:             10000,
		Output:            "text",
		MaxFailPercentage: 0,
		AnyErrorsFatal:    true,
		IgnoreErrors:      false,
		IgnoreUnreachable: false,
		OmitEmptyRows:     false,
		OmitEmptyColumns:  false,
		Report:            []string{"recap"},
		Verbose:           false,
		Confirm:           false,
		Step:              false,
		Print:             "all",
	}
)

type Config struct {
	SSHConfigFile     *string
	DefaultTimeout    uint
	DisableVerifyHost bool
	KnownHostsFile    string
	Shell             string
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
	DefaultTimeout    *uint     `yaml:"default_timeout"`
	KnownHostsFile    *string   `yaml:"known_hosts_file"`
	Shell             string    `yaml:"shell"`
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
func ReadConfig(configFilepath string, userConfigPath string, sshConfigFile string, noColor bool) (Config, error) {
	CheckUserNoColor(noColor)
	var configPath string

	userConfigFile, err := getUserConfigFile(userConfigPath)
	if err != nil {
		return Config{}, err
	}

	sshConfigPath, err := getSSHConfigPath(sshConfigFile)
	if err != nil {
		return Config{}, err
	}

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

	dat, err := os.ReadFile(configPath)
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
	config.SSHConfigFile = sshConfigPath
	config.CheckConfigNoColor()

	if configErr != nil {
		return config, configErr
	}

	return config, nil
}

func getUserConfigFile(userConfigPath string) (*string, error) {
	// Flag
	if userConfigPath != "" {
		if _, err := os.Stat(userConfigPath); err != nil {
			return nil, fmt.Errorf("user config not found: %w", err)
		}
		return &userConfigPath, nil
	}

	// Env
	val, present := os.LookupEnv("SAKE_USER_CONFIG")
	if present {
		if _, err := os.Stat(val); err != nil {
			return nil, fmt.Errorf("user config not found: %w", err)
		}
		return &val, nil
	}

	// Default
	defaultUserConfigDir, _ := os.UserConfigDir()

	defaultUserConfigPath := filepath.Join(defaultUserConfigDir, "sake", "config.yaml")
	if _, err := os.Stat(defaultUserConfigPath); err == nil {
		return &defaultUserConfigPath, nil
	}

	defaultUserConfigPath = filepath.Join(defaultUserConfigDir, "sake", "config.yml")
	if _, err := os.Stat(defaultUserConfigPath); err == nil {
		return &defaultUserConfigPath, nil
	}

	return nil, nil
}

func getSSHConfigPath(sshConfigPath string) (*string, error) {
	// Flag
	if sshConfigPath != "" {
		if _, err := os.Stat(sshConfigPath); err == nil {
			return &sshConfigPath, nil
		} else {
			return &sshConfigPath, err
		}
	}

	// Env
	val, present := os.LookupEnv("SAKE_SSH_CONFIG")
	if present {
		return &val, nil
	}

	// User SSH config
	if home, err := os.UserHomeDir(); err == nil {
		userSSHConfigFile := filepath.Join(home, ".ssh", "config")
		if _, err := os.Stat(userSSHConfigFile); err == nil {
			return &userSSHConfigFile, nil
		}
	}

	// Global SSH config
	globalSSHConfig := "/etc/ssh/ssh_config"
	if _, err := os.Stat(globalSSHConfig); err == nil {
		return &globalSSHConfig, nil
	}

	return nil, nil
}

// Open sake config in editor
func (c *Config) EditConfig() error {
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
		case "nvim":
			args = []string{fmt.Sprintf("+%v", lineNr), path}
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
func (c *Config) EditTask(name string) error {
	configPath := c.Path
	if name != "" {
		task, err := c.GetTask(name)
		if err != nil {
			return err
		}
		configPath = task.context
	}

	dat, err := os.ReadFile(configPath)
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
func (c *Config) EditServer(name string) error {
	var group string
	configPath := c.Path
	if name != "" {
		server, err := c.GetServerByGroup(name)
		if err != nil {
			return err
		}
		group = server.Group
		configPath = server.context
	}

	dat, err := os.ReadFile(configPath)
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
			if server.Value == group {
				lineNr = server.Line
				break
			}
		}
	}

	return openEditor(configPath, lineNr)
}

// Open sake config in editor and optionally go to line matching the target name
func (c *Config) EditTarget(name string) error {
	configPath := c.Path
	if name != "" {
		target, err := c.GetTarget(name)
		if err != nil {
			return err
		}
		configPath = target.context
	}

	dat, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	type ConfigTmp struct {
		Targets yaml.Node
	}

	var configTmp ConfigTmp
	err = yaml.Unmarshal(dat, &configTmp)
	if err != nil {
		return err
	}

	lineNr := 0
	if name == "" {
		lineNr = configTmp.Targets.Line - 1
	} else {
		for _, target := range configTmp.Targets.Content {
			if target.Value == name {
				lineNr = target.Line
				break
			}
		}
	}

	return openEditor(configPath, lineNr)
}

// Open sake config in editor and optionally go to line matching the spec name
func (c *Config) EditSpec(name string) error {
	configPath := c.Path
	if name != "" {
		spec, err := c.GetSpec(name)
		if err != nil {
			return err
		}
		configPath = spec.context
	}

	dat, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	type ConfigTmp struct {
		Specs yaml.Node
	}

	var configTmp ConfigTmp
	err = yaml.Unmarshal(dat, &configTmp)
	if err != nil {
		return err
	}

	lineNr := 0
	if name == "" {
		lineNr = configTmp.Specs.Line - 1
	} else {
		for _, spec := range configTmp.Specs.Content {
			if spec.Value == name {
				lineNr = spec.Line
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

func (c *Config) ParseInventory(userArgs []string) error {
	var servers []Server
	var shell = DEFAULT_SHELL
	if c.Shell != "" {
		shell = c.Shell
	}
	shell = core.FormatShell(shell)

	for _, s := range c.Servers {
		if s.Inventory != "" {
			hosts, err := core.EvaluateInventory(shell, s.context, s.Inventory, s.Envs, userArgs)
			if err != nil {
				return err
			}

			if len(hosts) == 0 {
				return fmt.Errorf("inventory %s returned 0 hosts", s.Name)
			}

			for i, host := range hosts {
				server, err := CreateInventoryServers(host, i, s, userArgs)
				if err != nil {
					return err
				}

				servers = append(servers, server)
			}

		} else {
			servers = append(servers, s)
		}
	}

	c.Servers = servers

	return nil
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
