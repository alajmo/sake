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
	User         string
	Port         uint8
	Local        bool
	Tags         []string
	Envs         []string
	WorkDir      string
	IdentityFile *string
	Password     *string

	context     string // config path
	contextLine int    // defined at
}

type ServerYAML struct {
	Name         string    `yaml:"-"`
	Desc         string    `yaml:"desc"`
	Host         string    `yaml:"host"`
	User         string    `yaml:"user"`
	Port         uint8     `yaml:"port"`
	Local        bool      `yaml:"local"`
	Tags         []string  `yaml:"tags"`
	Env          yaml.Node `yaml:"env"`
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
			fmt.Sprintf("SAKE_SERVER_DESC=%s", serverYAML.Desc),
			fmt.Sprintf("SAKE_SERVER_TAGS=%s", strings.Join(serverYAML.Tags, ",")),
			fmt.Sprintf("SAKE_SERVER_HOST=%s", serverYAML.Host),
			fmt.Sprintf("SAKE_SERVER_USER=%s", serverYAML.User),
			fmt.Sprintf("SAKE_SERVER_PORT=%d", serverYAML.Port),
			fmt.Sprintf("SAKE_SERVER_LOCAL=%t", serverYAML.Local),
		}

		server.Name = serverYAML.Name
		server.Desc = serverYAML.Desc
		server.Host = serverYAML.Host
		server.User = serverYAML.User
		server.Port = serverYAML.Port
		server.Local = serverYAML.Local
		server.Tags = serverYAML.Tags
		server.WorkDir = serverYAML.WorkDir
		server.Envs = append(envs, defaultEnvs...)

		if serverYAML.IdentityFile != nil {
			identityFile := os.ExpandEnv(*serverYAML.IdentityFile)
			serverYAML.IdentityFile = &identityFile

			if strings.HasPrefix(*serverYAML.IdentityFile, "~/") {
				// Expand tilde ~
				home, err := os.UserHomeDir()
				if err != nil {
					panic(err)
				}

				identityFile := *serverYAML.IdentityFile
				identityFile = filepath.Join(home, identityFile[2:])
				server.IdentityFile = &identityFile
			} else {
				// Absolute filepath
				if filepath.IsAbs(*serverYAML.IdentityFile) {
					server.IdentityFile = serverYAML.IdentityFile
				} else {
					// Relative filepath
					identityFile := filepath.Join(c.Dir, *serverYAML.IdentityFile)
					server.IdentityFile = &identityFile
					fmt.Println(*server.IdentityFile)
				}
			}

		}

		if serverYAML.Password != nil {
			server.Password = serverYAML.Password
		}

		servers = append(servers, *server)
	}

	return servers, serverErrors
}

func ServerInSlice(name string, list []Server) bool {
	for _, s := range list {
		if s.Name == name {
			return true
		}
	}
	return false
}

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

		finalServers = GetUnionServers(tagServers, servers)
	}

	return finalServers, nil
}

func (c Config) GetServerHosts() []string {
	hosts := []string{}
	for _, server := range c.Servers {
		if server.Host != "" {
			hosts = append(hosts, server.Host)
		}
	}

	return hosts
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
			if v == s.Name {
				foundServerNames[s.Name] = true
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
