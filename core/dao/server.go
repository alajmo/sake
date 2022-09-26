package dao

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/sake/core"
)

type Server struct {
	Name         string
	Desc         string
	Host         string
	BastionHost  string
	BastionUser  string
	BastionPort  uint16
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

	context     string // config path
	contextLine int    // defined at
}

type ServerYAML struct {
	Name         string    `yaml:"-"`
	Desc         string    `yaml:"desc"`
	Host         string    `yaml:"host"`
	Hosts        yaml.Node `yaml:"hosts"`
	Bastion      string    `yaml:"bastion"`
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
	switch key {
	case "Server", "server":
		return s.Name
	case "Host", "host":
		return s.Host
	case "Bastion", "bastion":
		return s.BastionHost
	case "User", "user":
		return s.User
	case "Port", "port":
		return strconv.Itoa(int(s.Port))
	case "Local", "local":
		return strconv.FormatBool(s.Local)
	case "Desc", "desc", "Description", "description":
		return s.Desc
	case "Tag", "tag":
		return strings.Join(s.Tags, ", ")
	}

	return ""
}

func (s *Server) GetContext() string {
	return s.context
}

func (s *Server) GetContextLine() int {
	return s.contextLine
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
				panic(err)
			}

			if err != nil {
				serverErrors[j].Errors = append(serverErrors[j].Errors, err)
				continue
			}

			serverYAML.User = user.Username
		}

		if serverYAML.Port == 0 {
			serverYAML.Port = 22
		}

		hostDef, err := getServerHostDefinition(serverYAML)
		if err != nil {
			serverErrors[j].Errors = append(serverErrors[j].Errors, err)
			continue
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

		defaultEnvs := []string{
			fmt.Sprintf("SAKE_SERVER_NAME=%s", serverYAML.Name),
		}

		if serverYAML.Desc != "" {
			defaultEnvs = append(defaultEnvs, fmt.Sprintf("SAKE_SERVER_DESC=%s", serverYAML.Desc))
		}
		if len(serverYAML.Tags) > 0 {
			defaultEnvs = append(defaultEnvs, fmt.Sprintf("SAKE_SERVER_TAGS=%s", strings.Join(serverYAML.Tags, ",")))
		}
		if serverYAML.Bastion != "" {
			defaultEnvs = append(defaultEnvs, fmt.Sprintf("SAKE_SERVER_BASTION=%s", serverYAML.Bastion))
		}
		if serverYAML.Local {
			defaultEnvs = append(defaultEnvs, fmt.Sprintf("SAKE_SERVER_LOCAL=%t", serverYAML.Local))
		}

		// Same for all servers
		var bastionUser string
		var bastionHost string
		var bastionPort uint16
		if serverYAML.Bastion != "" {
			bUser, bHost, bPort, err := core.ParseHostName(serverYAML.Bastion, serverYAML.User, serverYAML.Port)
			if err != nil {
				serverErrors[j].Errors = append(serverErrors[j].Errors, err)
				continue
			}

			bastionHost = bHost
			bastionUser = bUser
			bastionPort = bPort
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

		// Return error if file not found

		// Same for all servers
		var password *string
		if serverYAML.Password != nil {
			password = serverYAML.Password
		}

		switch hostDef {
		case "host":
			// string to be evaluated and will result in list of hosts
			user, host, port, err := core.ParseHostName(serverYAML.Host, serverYAML.User, serverYAML.Port)
			if err != nil {
				serverErrors[j].Errors = append(serverErrors[j].Errors, err)
				continue
			}

			serverEnvs := append(defaultEnvs, []string{
				fmt.Sprintf("SAKE_SERVER_HOST=%s", host),
				fmt.Sprintf("SAKE_SERVER_USER=%s", user),
				fmt.Sprintf("SAKE_SERVER_PORT=%d", port),
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
				BastionHost:  bastionHost,
				BastionUser:  bastionUser,
				BastionPort:  bastionPort,
				IdentityFile: identityFile,
				PubFile:      pubKeyFile,
				Password:     password,

				context:     c.Path,
				contextLine: c.Servers.Content[i].Line,
			}

			servers = append(servers, *hServer)

		case "hosts":
			// list of servers

			for k, s := range serverYAML.Hosts.Content {
				user, host, port, err := core.ParseHostName(s.Value, serverYAML.User, serverYAML.Port)
				if err != nil {
					serverErrors[j].Errors = append(serverErrors[j].Errors, err)
					continue
				}

				serverEnvs := append(defaultEnvs, []string{
					fmt.Sprintf("SAKE_SERVER_HOST=%s", host),
					fmt.Sprintf("SAKE_SERVER_USER=%s", user),
					fmt.Sprintf("SAKE_SERVER_PORT=%d", port),
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
					BastionHost:  bastionHost,
					BastionUser:  bastionUser,
					BastionPort:  bastionPort,
					IdentityFile: identityFile,
					PubFile:      pubKeyFile,
					Password:     password,

					context:     c.Path,
					contextLine: c.Servers.Content[i].Line,
				}

				servers = append(servers, *hServer)
			}
		case "hosts-string":
			hosts, err := core.ExpandHostNames(server.context, serverYAML.Hosts.Value, defaultEnvs, envs)
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
					fmt.Sprintf("SAKE_SERVER_HOST=%s", host),
					fmt.Sprintf("SAKE_SERVER_USER=%s", user),
					fmt.Sprintf("SAKE_SERVER_PORT=%d", port),
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
					BastionHost:  bastionHost,
					BastionUser:  bastionUser,
					BastionPort:  bastionPort,
					IdentityFile: identityFile,
					PubFile:      pubKeyFile,
					Password:     password,

					context:     c.Path,
					contextLine: c.Servers.Content[i].Line,
				}

				servers = append(servers, *hServer)
			}
		}

		// servers = append(servers, *server)
	}

	return servers, serverErrors
}

func getServerHostDefinition(serverYAML *ServerYAML) (string, error) {
	hostDef := ""
	numDefined := 0
	if serverYAML.Host != "" {
		hostDef = "host"
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

func ServerInSlice(name string, list []Server) bool {
	for _, s := range list {
		if s.Name == name {
			return true
		}
	}
	return false
}

// FilterServers returns servers matching filters, it does a union select.
func (c Config) FilterServers(
	allServersFlag bool,
	serversFlag []string,
	tagsFlag []string,
) ([]Server, error) {
	var finalServers []Server
	if allServersFlag {
		finalServers = c.Servers
	} else {
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

		finalServers = GetUnionServers(servers, tagServers)
	}

	return finalServers, nil
}

func (c Config) GetServer(name string) (*Server, error) {
	for _, server := range c.Servers {
		if name == server.Name {
			return &server, nil
		}
	}

	return nil, &core.ServerNotFound{Name: []string{name}}
}

func (c Config) GetServersByName(serverNames []string) ([]Server, error) {
	var matchedServers []Server

	foundServerNames := make(map[string]bool)
	for _, s := range serverNames {
		foundServerNames[s] = false
	}

	for _, v := range serverNames {
		for _, s := range c.Servers {
			if v == s.Group {
				foundServerNames[s.Group] = true
				matchedServers = append(matchedServers, s)
			}
		}
	}

	nonExistingServers := []string{}
	for k, v := range foundServerNames {
		if !v {
			nonExistingServers = append(nonExistingServers, k)
		}
	}

	if len(nonExistingServers) > 0 {
		return []Server{}, &core.ServerNotFound{Name: nonExistingServers}
	}

	return matchedServers, nil
}

func (c Config) GetRemoteServerNameAndDesc() []string {
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

func (c Config) GetServerNameAndDesc() []string {
	options := []string{}
	for _, server := range c.Servers {
		options = append(options, fmt.Sprintf("%s\t%s", server.Name, server.Desc))
	}

	return options
}

// Servers must have all tags to match. For instance, if --tags frontend,backend
// is passed, then a server must have both tags.
// We only return error if the flags provided do not exist in the sake config.
func (c Config) GetServersByTags(tags []string) ([]Server, error) {
	if len(tags) == 0 {
		return c.Servers, nil
	}

	foundTags := make(map[string]bool)
	for _, tag := range tags {
		foundTags[tag] = false
	}

	// Find servers matching the flag
	var servers []Server
	for _, server := range c.Servers {
		// Variable use to check that all tags are matched
		var numMatched int = 0
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

func (c Config) GetServerNames() []string {
	names := []string{}
	for _, server := range c.Servers {
		names = append(names, server.Name)
	}

	return names
}

func GetUnionServers(s ...[]Server) []Server {
	servers := []Server{}
	for _, part := range s {
		for _, server := range part {
			if !ServerInSlice(server.Name, servers) {
				servers = append(servers, server)
			}
		}
	}

	return servers
}
