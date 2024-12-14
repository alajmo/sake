package dao

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/sake/core"
)

type Server struct {
	Name         string
	Desc         string
	Host         string
	Inventory    string
	Bastions     []Bastion
	User         string
	Port         uint16
	Local        bool
	Tags         []string
	Envs         []string
	Shell        string
	WorkDir      string
	IdentityFile *string
	Password     *string

	// Internal
	Group   string
	PubFile *string

	RootDir     string // config dir
	context     string // config path
	contextLine int    // defined at
}

type Bastion struct {
	Host string
	User string
	Port uint16
}

func (b Bastion) GetPrint() string {
	return fmt.Sprintf("%s@%s:%d", b.User, b.Host, b.Port)
}

type ServerYAML struct {
	Name         string    `yaml:"-"`
	Desc         string    `yaml:"desc"`
	Host         string    `yaml:"host"`
	Hosts        yaml.Node `yaml:"hosts"`
	Inventory    string    `yaml:"inventory"`
	Bastion      string    `yaml:"bastion"`
	Bastions     []string  `yaml:"bastions"`
	User         string    `yaml:"user"`
	Port         uint16    `yaml:"port"`
	Local        bool      `yaml:"local"`
	Tags         []string  `yaml:"tags"`
	Env          yaml.Node `yaml:"env"`
	Shell        string    `yaml:"shell"`
	WorkDir      string    `yaml:"work_dir"`
	IdentityFile *string   `yaml:"identity_file"`
	Password     *string   `yaml:"password"`
}

func (s Server) GetValue(key string, _ int) string {
	lkey := strings.ToLower(key)
	switch lkey {
	case "name", "server":
		return s.Name
	case "desc", "description":
		return s.Desc
	case "host":
		return s.Host
	case "bastion":
		return getBastionHosts(s.Bastions, "\n")
	case "user":
		return s.User
	case "port":
		return strconv.Itoa(int(s.Port))
	case "local":
		return strconv.FormatBool(s.Local)
	case "shell":
		return s.Shell
	case "work_dir":
		return s.WorkDir
	case "identity_file":
		if s.IdentityFile != nil {
			return path.Base(*s.IdentityFile)
		} else {
			return ""
		}
	case "tags":
		return strings.Join(s.Tags, ",")
	}

	return ""
}

func (s *Server) GetContext() string {
	return s.context
}

func (s *Server) GetContextLine() int {
	return s.contextLine
}

func (s *Server) GetNonDefaultEnvs() []string {
	var envs []string
	for _, env := range s.Envs {
		if !strings.Contains(env, "S_") {
			envs = append(envs, env)
		}
	}

	return envs
}

// ParseServersYAML parses the servers dictionary and returns it as a list.
func (c *ConfigYAML) ParseServersYAML() ([]Server, []ResourceErrors[Server]) {
	var servers []Server
	count := len(c.Servers.Content)

	serverErrors := []ResourceErrors[Server]{}
	j := -1
	for i := 0; i < count; i += 2 {
		j += 1
		server := &Server{
			context:     c.Path,
			contextLine: c.Servers.Content[i].Line,
		}
		serverYAML := &ServerYAML{}
		re := ResourceErrors[Server]{Resource: server, Errors: []error{}}
		serverErrors = append(serverErrors, re)

		err := c.Servers.Content[i+1].Decode(serverYAML)
		if err != nil {
			for _, yerr := range err.(*yaml.TypeError).Errors {
				serverErrors[j].Errors = append(serverErrors[j].Errors, errors.New(yerr))
			}
			continue
		}

		serverYAML.Name = c.Servers.Content[i].Value

		if serverYAML.User == "" {
			user, err := user.Current()

			if err != nil {
				serverErrors[j].Errors = append(serverErrors[j].Errors, err)
				continue
			}

			serverYAML.User = user.Username
		}

		if serverYAML.Port == 0 {
			serverYAML.Port = 22
		}

		var envs []string
		if !IsNullNode(serverYAML.Env) {
			err := CheckIsMappingNode(serverYAML.Env)
			if err != nil {
				serverErrors[j].Errors = append(serverErrors[j].Errors, err)
			} else {
				envs, err = EvaluateEnv(ParseNodeEnv(serverYAML.Env))
				if err != nil {
					for _, yerr := range err.(*yaml.TypeError).Errors {
						serverErrors[j].Errors = append(serverErrors[j].Errors, errors.New(yerr))
					}
					continue
				}
			}
		}

		bastionDef, err := getServerBastionDefinition(serverYAML)
		bastions := []Bastion{}
		if err != nil {
			serverErrors[j].Errors = append(serverErrors[j].Errors, err)
			continue
		}
		switch bastionDef {
		case "bastion":
			bUser, bHost, bPort, err := core.ParseHostName(serverYAML.Bastion, serverYAML.User, serverYAML.Port)
			if err != nil {
				serverErrors[j].Errors = append(serverErrors[j].Errors, err)
				continue
			}

			bastions = append(bastions, Bastion{
				User: bUser,
				Host: bHost,
				Port: bPort,
			})
		case "bastions":
			for _, bastionStr := range serverYAML.Bastions {
				bUser, bHost, bPort, err := core.ParseHostName(bastionStr, serverYAML.User, serverYAML.Port)
				if err != nil {
					serverErrors[j].Errors = append(serverErrors[j].Errors, err)
					continue
				}

				bastions = append(bastions, Bastion{
					User: bUser,
					Host: bHost,
					Port: bPort,
				})

			}
		}

		defaultEnvs := []string{}
		if serverYAML.IdentityFile != nil {
			defaultEnvs = append(defaultEnvs, fmt.Sprintf("S_IDENTITY=%s", *serverYAML.IdentityFile))
		}
		if len(serverYAML.Tags) > 0 {
			defaultEnvs = append(defaultEnvs, fmt.Sprintf("S_TAGS=%s", strings.Join(serverYAML.Tags, ",")))
		}
		if len(bastions) > 0 {
			defaultEnvs = append(defaultEnvs, fmt.Sprintf("S_BASTION=%s", getBastionHosts(bastions, ",")))
		}

		var identityFile *string
		// Same for all servers
		if serverYAML.IdentityFile != nil {
			iFile := os.ExpandEnv(*serverYAML.IdentityFile)
			identityFile = &iFile

			if strings.HasPrefix(*serverYAML.IdentityFile, "~/") {
				// Expand tilde ~
				home, err := os.UserHomeDir()
				if err != nil {
					panic(err)
				}

				iFile := filepath.Join(home, (*serverYAML.IdentityFile)[2:])
				identityFile = &iFile
			} else {
				// Absolute filepath
				if filepath.IsAbs(*serverYAML.IdentityFile) {
					iFile = *serverYAML.IdentityFile
				} else {
					// Relative filepath
					f := filepath.Join(c.Dir, *serverYAML.IdentityFile)
					iFile = f
				}
			}
		}

		// Same for all servers
		var password *string
		if serverYAML.Password != nil {
			password = serverYAML.Password
		}

		if identityFile == nil && password == nil {
			idFile := core.GetFirstExistingFile("~/.ssh/id_rsa", "~/.ssh/id_ecdsa", "~/.ssh/id_dsa")
			if idFile != "" {
				identityFile = &idFile
			}
		}

		var pubKeyFile *string
		if identityFile != nil {
			if _, err := os.Stat(*identityFile); errors.Is(err, os.ErrNotExist) {
				serverErrors[j].Errors = append(serverErrors[j].Errors, err)
				continue
			}
			if _, err := os.Stat(*identityFile + ".pub"); !errors.Is(err, os.ErrNotExist) {
				str := *identityFile + ".pub"
				pubKeyFile = &str
			}
		}

		hostDef, err := getServerHostDefinition(serverYAML)
		if err != nil {
			serverErrors[j].Errors = append(serverErrors[j].Errors, err)
			continue
		}
		switch hostDef {
		case "host":
			// host: test@192.168.0.1:22
			user, host, port, err := core.ParseHostName(serverYAML.Host, serverYAML.User, serverYAML.Port)
			if err != nil {
				serverErrors[j].Errors = append(serverErrors[j].Errors, err)
				continue
			}

			serverEnvs := append(defaultEnvs, []string{
				fmt.Sprintf("S_NAME=%s", c.Servers.Content[i].Value),
				fmt.Sprintf("S_HOST=%s", host),
				fmt.Sprintf("S_USER=%s", user),
				fmt.Sprintf("S_PORT=%d", port),
			}...)
			serverEnvs = append(serverEnvs, envs...)

			hServer := &Server{
				Name:         c.Servers.Content[i].Value,
				Group:        c.Servers.Content[i].Value,
				Desc:         serverYAML.Desc,
				Host:         host,
				User:         user,
				Port:         port,
				Local:        serverYAML.Local,
				Tags:         serverYAML.Tags,
				Shell:        serverYAML.Shell,
				WorkDir:      serverYAML.WorkDir,
				Envs:         serverEnvs,
				Bastions:     bastions,
				IdentityFile: identityFile,
				PubFile:      pubKeyFile,
				Password:     password,

				RootDir:     filepath.Dir(c.Path),
				context:     c.Path,
				contextLine: c.Servers.Content[i].Line,
			}

			servers = append(servers, *hServer)
		case "hosts":
			// list of hosts

			for k, s := range serverYAML.Hosts.Content {
				user, host, port, err := core.ParseHostName(s.Value, serverYAML.User, serverYAML.Port)
				if err != nil {
					serverErrors[j].Errors = append(serverErrors[j].Errors, err)
					continue
				}

				serverEnvs := append(defaultEnvs, []string{
					fmt.Sprintf("S_NAME=%s-%d", c.Servers.Content[i].Value, k),
					fmt.Sprintf("S_HOST=%s", host),
					fmt.Sprintf("S_USER=%s", user),
					fmt.Sprintf("S_PORT=%d", port),
				}...)
				serverEnvs = append(serverEnvs, envs...)

				hServer := &Server{
					Name:         fmt.Sprintf("%s-%d", c.Servers.Content[i].Value, k),
					Group:        c.Servers.Content[i].Value,
					Desc:         serverYAML.Desc,
					Host:         host,
					User:         user,
					Port:         port,
					Local:        serverYAML.Local,
					Tags:         serverYAML.Tags,
					Shell:        serverYAML.Shell,
					WorkDir:      serverYAML.WorkDir,
					Envs:         serverEnvs,
					Bastions:     bastions,
					IdentityFile: identityFile,
					PubFile:      pubKeyFile,
					Password:     password,

					RootDir:     filepath.Dir(c.Path),
					context:     c.Path,
					contextLine: c.Servers.Content[i].Line,
				}

				servers = append(servers, *hServer)
			}
		case "hosts-string":
			// hosts: 192.168.[0:3].1
			hosts, err := core.EvaluateRange(serverYAML.Hosts.Value)
			if err != nil {
				serverErrors[j].Errors = append(serverErrors[j].Errors, err)
				continue
			}

			for k, s := range hosts {
				user, host, port, err := core.ParseHostName(s, serverYAML.User, serverYAML.Port)
				if err != nil {
					serverErrors[j].Errors = append(serverErrors[j].Errors, err)
					continue
				}

				serverEnvs := append(defaultEnvs, []string{
					fmt.Sprintf("S_NAME=%s-%d", c.Servers.Content[i].Value, k),
					fmt.Sprintf("S_HOST=%s", host),
					fmt.Sprintf("S_USER=%s", user),
					fmt.Sprintf("S_PORT=%d", port),
				}...)
				serverEnvs = append(serverEnvs, envs...)

				hServer := &Server{
					Name:         fmt.Sprintf("%s-%d", c.Servers.Content[i].Value, k),
					Group:        c.Servers.Content[i].Value,
					Desc:         serverYAML.Desc,
					Host:         host,
					User:         user,
					Port:         port,
					Local:        serverYAML.Local,
					Tags:         serverYAML.Tags,
					Shell:        serverYAML.Shell,
					WorkDir:      serverYAML.WorkDir,
					Envs:         serverEnvs,
					Bastions:     bastions,
					IdentityFile: identityFile,
					PubFile:      pubKeyFile,
					Password:     password,

					RootDir:     filepath.Dir(c.Path),
					context:     c.Path,
					contextLine: c.Servers.Content[i].Line,
				}

				servers = append(servers, *hServer)
			}
		case "inventory":
			// User provides command to evaluate to a list of hosts
			serverEnvs := append(defaultEnvs, envs...)
			hServer := &Server{
				Name:         c.Servers.Content[i].Value,
				Group:        c.Servers.Content[i].Value,
				Desc:         serverYAML.Desc,
				Inventory:    serverYAML.Inventory,
				User:         serverYAML.User,
				Port:         serverYAML.Port,
				Local:        serverYAML.Local,
				Tags:         serverYAML.Tags,
				Shell:        serverYAML.Shell,
				WorkDir:      serverYAML.WorkDir,
				Envs:         serverEnvs,
				Bastions:     bastions,
				IdentityFile: identityFile,
				PubFile:      pubKeyFile,
				Password:     password,

				RootDir:     filepath.Dir(c.Path),
				context:     c.Path,
				contextLine: c.Servers.Content[i].Line,
			}

			servers = append(servers, *hServer)
		}
	}

	return servers, serverErrors
}

func getServerBastionDefinition(serverYAML *ServerYAML) (string, error) {
	bastionDef := ""
	numDefined := 0
	if serverYAML.Bastion != "" {
		bastionDef = "bastion"
		numDefined += 1
	}
	if len(serverYAML.Bastions) > 0 {
		numDefined += 1
		bastionDef = "bastions"
	}

	if numDefined > 1 {
		return "", &core.ServerBastionMultipleDef{Name: serverYAML.Name}
	}

	return bastionDef, nil
}

func getServerHostDefinition(serverYAML *ServerYAML) (string, error) {
	hostDef := ""
	numDefined := 0
	if serverYAML.Host != "" {
		hostDef = "host"
		numDefined += 1
	}
	if serverYAML.Inventory != "" {
		hostDef = "inventory"
		numDefined += 1
	}
	if serverYAML.Hosts.Kind == 2 && len(serverYAML.Hosts.Content) > 0 {
		// list of servers
		numDefined += 1
		hostDef = "hosts"
	}
	if serverYAML.Hosts.Kind == 8 && serverYAML.Hosts.Value != "" {
		// string to be evaluated and will result in list of hosts
		numDefined += 1
		hostDef = "hosts-string"
	}

	if numDefined > 1 {
		return "", &core.ServerMultipleDef{Name: serverYAML.Name}
	}

	return hostDef, nil
}

// FilterServers returns servers matching filters, it does a union select.
func (c *Config) FilterServers(
	allServersFlag bool,
	serversFlag []string,
	tagsFlag []string,
	regexFlag string,
	invertFlag bool,
) ([]Server, error) {
	var finalServers []Server

	var allServers []Server
	if allServersFlag {
		allServers = c.Servers
	}

	var err error
	var servers []Server
	if len(serversFlag) > 0 {
		servers, err = c.GetServersByName(serversFlag)
		if err != nil {
			return []Server{}, err
		}
	}

	var tagServers []Server
	if len(tagsFlag) > 0 {
		tagServers, err = c.GetServersByTags(tagsFlag)
		if err != nil {
			return []Server{}, err
		}
	}

	var regexServers []Server
	if regexFlag != "" {
		regexServers, err = c.GetServersByRegex(regexFlag)
		if err != nil {
			return []Server{}, err
		}
	}

	finalServers = GetIntersectionServers(allServers, servers, tagServers, regexServers)

	if invertFlag {
		finalServers = GetInvertedServers(c.Servers, finalServers)
	}

	return finalServers, nil
}

func (c *Config) GetServer(name string) (*Server, error) {
	for _, server := range c.Servers {
		if name == server.Name {
			return &server, nil
		}
	}

	return nil, &core.ServerNotFound{Name: []string{name}}
}

func (c *Config) GetServerByGroup(group string) (*Server, error) {
	for _, server := range c.Servers {
		if group == server.Group {
			return &server, nil
		}
	}

	return nil, &core.ServerNotFound{Name: []string{group}}
}

func (c *Config) GetServersByName(groupNames []string) ([]Server, error) {
	serverGroup := make(map[string]Server, len(c.Servers))
	groupServers := make(map[string][]Server, len(c.Servers))
	for _, s := range c.Servers {
		serverGroup[s.Name] = s
		_, found := groupServers[s.Group]
		if found {
			groupServers[s.Group] = append(groupServers[s.Group], s)
		} else {
			groupServers[s.Group] = []Server{s}
		}
	}

	var matchedServers []Server
	for _, groupName := range groupNames {
		// If groupName is a server name, add to matched and continue
		s, found := serverGroup[groupName]
		if found {
			matchedServers = append(matchedServers, s)
			continue
		}

		// If groupName is a group, then check if there's a range, otherwise return all servers in the group
		if containsRange(groupName) {
			targetRange, err := parseRange(groupName)
			if err != nil {
				return []Server{}, err
			}

			servers, err := evalTargetRange(groupServers, targetRange)
			if err != nil {
				return []Server{}, err
			}

			if len(groupServers[targetRange.Group]) > 0 {
				matchedServers = append(matchedServers, servers...)
			} else {
				return []Server{}, fmt.Errorf("cannot find server %s", groupName)
			}
		} else {
			if len(groupServers[groupName]) > 0 {
				matchedServers = append(matchedServers, groupServers[groupName]...)
			} else {
				return []Server{}, fmt.Errorf("cannot find server %s", groupName)
			}
		}
	}

	// Handle duplicates
	addedServer := make(map[string]bool)
	for _, s := range matchedServers {
		addedServer[s.Name] = false
	}
	var servers []Server
	for _, s := range matchedServers {
		if !addedServer[s.Name] {
			servers = append(servers, s)
			addedServer[s.Name] = true
		}
	}

	return servers, nil
}

func containsRange(s string) bool {
	return strings.Contains(s, "[")
}

type TargetRange struct {
	Group    string
	Start    int
	End      int
	Range    bool
	HasStart bool
	HasEnd   bool
}

// input: a, b, c, d
// [0] -> a
// [1:] -> b, c, d
// [1:2] -> b, c
// [:2] -> a, b, c
func parseRange(s string) (TargetRange, error) {
	if strings.Count(s, "[") != 1 || strings.Count(s, "]") != 1 {
		return TargetRange{}, fmt.Errorf("invalid server range")
	}
	if strings.Count(s, ":") > 1 {
		return TargetRange{}, fmt.Errorf("invalid server range")
	}

	state := 0
	group := ""
	start := ""
	hasStart := false
	end := ""
	hasEnd := false
	rrange := false
	i := 0
	for i < len(s) {
		if string(s[i]) == "]" {
			break
		} else if string(s[i]) == "[" {
			state = 1
			i += 1
			continue
		} else if string(s[i]) == ":" {
			rrange = true
			state = 2
			i += 1
			continue
		}

		switch state {
		case 0:
			group += string(s[i])
		case 1:
			if !core.IsDigit(string(s[i])) {
				return TargetRange{}, fmt.Errorf("only [0-9] allowed in server range")
			}

			hasStart = true
			start += string(s[i])
		case 2:
			if !core.IsDigit(string(s[i])) {
				return TargetRange{}, fmt.Errorf("only [0-9] allowed in server range")
			}
			hasEnd = true
			end += string(s[i])
		}
		i += 1
	}

	rr := TargetRange{
		Group:    group,
		HasStart: hasStart,
		HasEnd:   hasEnd,
		Range:    rrange,
	}

	if hasStart {
		ss, err := strconv.ParseInt(start, 10, 0)
		if err != nil {
			return TargetRange{}, err
		}
		rr.Start = int(ss)
	}

	if hasEnd {
		ee, err := strconv.ParseInt(end, 10, 0)
		if err != nil {
			return TargetRange{}, err
		}
		rr.End = int(ee)
	}

	return rr, nil
}

// input: a, b, c, d
//
// [0] -> a
//
// [1:2] -> b, c (start = 1, end = 2 | len(hm[tr.Group]))
// hasStart = true
// hasEnd = true
//
// [1:] -> b, c, d (start = 1, end = len(hm[tr.Group]))
// hasStart = true
// hasEnd = false
//
// [:2] -> a, b, c, (start = 0, end = 2 | len(hm[tr.Group]))
// hasStart = false
// hasEnd = true
func evalTargetRange(hm map[string][]Server, tr TargetRange) ([]Server, error) {
	s, found := hm[tr.Group]
	if !found {
		return []Server{}, fmt.Errorf("could not find server %s", tr.Group)
	}

	// Keep start/end within array boundary
	if tr.Start > len(hm[tr.Group])-1 {
		tr.Start = len(hm[tr.Group]) - 1
	}
	if tr.End > len(hm[tr.Group])-1 {
		tr.End = len(hm[tr.Group]) - 1
	}

	// Handle single elements
	if !tr.Range {
		return []Server{s[tr.Start]}, nil
	}

	start := 0
	end := 0
	if tr.HasStart && tr.HasEnd {
		start = tr.Start
		end = tr.End
	} else if tr.HasStart {
		start = tr.Start
		end = len(hm[tr.Group]) - 1
	} else if tr.HasEnd {
		start = 0
		end = tr.End
	} else {
		// Handle [:], to include all hosts
		start = 0
		end = len(hm[tr.Group]) - 1
	}

	var servers []Server
	for i := start; i <= end; i++ {
		servers = append(servers, hm[tr.Group][i])
	}

	return servers, nil
}

// Matches on server host
func (c *Config) GetServersByRegex(r string) ([]Server, error) {
	pattern, err := regexp.Compile(r)
	if err != nil {
		return []Server{}, err
	}

	// Find servers matching the flag
	var servers []Server
	for _, server := range c.Servers {
		match := pattern.MatchString(server.Host)
		if match {
			servers = append(servers, server)
		}
	}

	if len(servers) == 0 {
		return []Server{}, fmt.Errorf("cannot find server any servers matching regex %s", r)
	}

	return servers, nil
}

func (c *Config) GetRemoteServerNameAndDesc() []string {
	options := []string{}
	for _, server := range c.Servers {
		if !server.Local {
			options = append(options, fmt.Sprintf("%s\t%s", server.Name, server.Desc))
		}
	}

	return options
}

func GetFirstRemoteServer(servers []Server) (Server, error) {
	for _, s := range servers {
		if !s.Local {
			return s, nil
		}
	}

	return Server{}, &core.NoRemoteServerToAttach{}
}

func (c *Config) GetServerNameAndDesc() []string {
	options := []string{}
	hm := make(map[string][]Server, len(c.Servers))
	for _, s := range c.Servers {
		_, found := hm[s.Group]
		if found {
			options = append(options, fmt.Sprintf("%s\t%s", s.Name, s.Desc))
			hm[s.Group] = append(hm[s.Group], s)
		} else {
			// Has multiple servers, this only runs once per group
			if s.Name != s.Group {
				options = append(options, fmt.Sprintf("%s\t%s", s.Group, s.Desc))
			}
			options = append(options, fmt.Sprintf("%s\t%s", s.Name, s.Desc))
			hm[s.Group] = []Server{s}
		}
	}

	return options
}

func (c *Config) GetServerGroupsAndDesc() []string {
	options := []string{}
	hm := make(map[string][]Server, len(c.Servers))
	for _, s := range c.Servers {
		_, found := hm[s.Group]
		if !found {
			options = append(options, fmt.Sprintf("%s\t%s", s.Group, s.Desc))
		}
	}

	return options
}

// Servers must have all tags to match. For instance, if --tags frontend,backend
// is passed, then a server must have both tags.
// We only return error if the flags provided do not exist in the sake config.
func (c *Config) GetServersByTags(tags []string) ([]Server, error) {
	foundTags := make(map[string]bool)
	for _, tag := range tags {
		foundTags[tag] = false
	}

	// Find servers matching the flag
	var servers []Server
	for _, server := range c.Servers {
		// Variable use to check that all tags are matched
		numMatched := 0
		for _, tag := range tags {
			for _, serverTag := range server.Tags {
				if serverTag == tag {
					foundTags[tag] = true
					numMatched = numMatched + 1
				}
			}
		}

		if numMatched == len(tags) {
			servers = append(servers, server)
		}
	}

	nonExistingTags := []string{}
	for k, v := range foundTags {
		if !v {
			nonExistingTags = append(nonExistingTags, k)
		}
	}

	if len(nonExistingTags) > 0 {
		return []Server{}, &core.TagNotFound{Tags: nonExistingTags}
	}

	return servers, nil
}

func GetIntersectionServers(s ...[]Server) []Server {
	var count int
	for _, part := range s {
		if len(part) > 0 {
			count += 1
		}
	}

	var foundServers []Server
	ss := make(map[string]int)
	for _, part := range s {
		for _, server := range part {
			_, found := ss[server.Name]
			if found {
				ss[server.Name] += 1
			} else {
				foundServers = append(foundServers, server)
				ss[server.Name] = 1
			}
		}
	}

	var servers []Server
	for _, server := range foundServers {
		if ss[server.Name] == count {
			servers = append(servers, server)
		}
	}

	return servers
}

func GetInvertedServers(allServers []Server, excludeServers []Server) []Server {
	sm := make(map[string]bool, len(excludeServers))
	for _, s := range excludeServers {
		sm[s.Name] = true
	}

	var servers []Server
	for _, s := range allServers {
		_, found := sm[s.Name]
		if !found {
			servers = append(servers, s)
		}
	}

	return servers
}

func CreateInventoryServers(inputHost string, i int, server Server, userArgs []string) (Server, error) {
	user, host, port, err := core.ParseHostName(inputHost, server.User, server.Port)
	if err != nil {
		return server, err
	}

	serverEnvs := append(server.Envs, []string{
		fmt.Sprintf("S_HOST=%s", host),
		fmt.Sprintf("S_USER=%s", user),
		fmt.Sprintf("S_PORT=%d", port),
	}...)
	serverEnvs = append(serverEnvs, userArgs...)

	iServer := &Server{
		Name:         fmt.Sprintf("%s-%d", server.Name, i),
		Group:        server.Group,
		Desc:         server.Desc,
		Host:         host,
		User:         user,
		Port:         port,
		Local:        server.Local,
		Tags:         server.Tags,
		Shell:        server.Shell,
		WorkDir:      server.WorkDir,
		Envs:         serverEnvs,
		IdentityFile: server.IdentityFile,
		PubFile:      server.PubFile,
		Password:     server.Password,

		context:     server.context,
		contextLine: server.contextLine,
	}

	return *iServer, nil
}

func SortServers(order string, servers *[]Server) {
	switch order {
	case "inventory":
	case "reverse_inventory":
		for i, j := 0, len(*servers)-1; i < j; i, j = i+1, j-1 {
			(*servers)[i], (*servers)[j] = (*servers)[j], (*servers)[i]
		}
	case "sorted":
		sort.Slice(*servers, func(i, j int) bool {
			return (*servers)[i].Host < (*servers)[j].Host
		})
	case "reverse_sorted":
		sort.Slice(*servers, func(i, j int) bool {
			return (*servers)[i].Host > (*servers)[j].Host
		})
	case "random":
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(len((*servers)), func(i, j int) { (*servers)[i], (*servers)[j] = (*servers)[j], (*servers)[i] })
	}
}

func getBastionHosts(bastions []Bastion, splitOn string) string {
	if len(bastions) == 1 {
		return bastions[0].GetPrint()
	}
	output := ""
	for i, bastion := range bastions {
		b := bastion.GetPrint()

		if i < len(bastions)-1 {
			output += fmt.Sprintf("%s%s", b, splitOn)
		} else {
			output += b
		}
	}

	return output
}
