// TODO: Refactor
package core

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gobwas/glob"
	"github.com/kevinburke/ssh_config"
)

// Parse reads and parses the file in the given path.
func ParseSSHConfig(path string) (map[string](Endpoint), error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config: %w", err)
	}
	defer f.Close()

	endpoints, err := ParseReader(f, path)
	if err != nil {
		return nil, err
	}

	hosts := make(map[string]Endpoint)
	for _, e := range endpoints {
		hosts[e.Name] = *e
	}

	return hosts, nil
}

// ParseReader reads and parses the given reader.
func ParseReader(r io.Reader, cfg string) ([]*Endpoint, error) {
	infos, err := parseInternal(r, cfg)
	if err != nil {
		return nil, err
	}

	wildcards, hosts := split(infos)

	endpoints := make([]*Endpoint, 0, infos.length())
	if err := hosts.forEach(func(name string, info hostinfo, err error) error {
		if err != nil {
			return err
		}
		if err := wildcards.forEach(func(k string, v hostinfo, err error) error {
			if err != nil {
				return err
			}

			g, err := glob.Compile(k)
			if err != nil {
				return fmt.Errorf("%s: invalid Host: %q: %w", cfg, k, err)
			}

			if g.Match(name) || (info.HostName != "" && g.Match(info.HostName)) {
				info = mergeHostinfo(info, v)
			}

			return nil
		}); err != nil {
			return err
		}

		endpoints = append(endpoints, &Endpoint{
			Name:          name,
			HostName:      firstNonEmpty(info.HostName, name),
			Port:          info.Port,
			User:          info.User,
			ProxyJump:     info.ProxyJump,
			IdentityFiles: info.IdentityFiles,
			ForwardAgent:  StringToBool(info.ForwardAgent),
			RequestTTY:    StringToBool(info.RequestTTY),
			RemoteCommand: info.RemoteCommand,
			SetEnv:        info.SetEnv,
			SendEnv:       info.SendEnv,
		})
		return nil
	}); err != nil {
		return nil, err
	}

	return endpoints, nil
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

type Endpoint struct {
	Name          string
	HostName      string
	Port          string
	User          string
	ProxyJump     string
	ForwardAgent  bool
	RequestTTY    bool
	RemoteCommand string
	SendEnv       []string
	SetEnv        []string
	IdentityFiles []string
}

type hostinfo struct {
	HostName      string
	Port          string
	User          string
	ProxyJump     string
	ForwardAgent  string
	RequestTTY    string
	RemoteCommand string
	SendEnv       []string
	SetEnv        []string
	IdentityFiles []string
}

type hostinfoMap struct {
	inner map[string]hostinfo
	keys  []string
	lock  sync.Mutex
}

func (m *hostinfoMap) length() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return len(m.keys)
}

func (m *hostinfoMap) set(k string, v hostinfo) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.inner[k]; !ok {
		m.keys = append(m.keys, k)
	}
	m.inner[k] = v
}

func (m *hostinfoMap) get(k string) (hostinfo, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	v, ok := m.inner[k]
	return v, ok
}

func (m *hostinfoMap) forEach(fn func(string, hostinfo, error) error) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	var err error
	for _, k := range m.keys {
		err = fn(k, m.inner[k], err)
	}
	return err
}

func newHostinfoMap() *hostinfoMap {
	return &hostinfoMap{
		inner: map[string]hostinfo{},
	}
}

func parseInternal(r io.Reader, cfg string) (*hostinfoMap, error) {
	bts, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	var rb bytes.Buffer
	for _, line := range bytes.Split(bts, []byte("\n")) {
		if bytes.HasPrefix(bytes.TrimSpace(bytes.ToLower(line)), []byte("match")) {
			continue
		}
		if _, err := rb.Write(line); err != nil {
			return nil, fmt.Errorf("failed to parse: %w", err)
		}
		if _, err := rb.Write([]byte("\n")); err != nil {
			return nil, fmt.Errorf("failed to parse: %w", err)
		}
	}

	config, err := ssh_config.Decode(&rb)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	infos := newHostinfoMap()

	for _, h := range config.Hosts {
		for _, pattern := range h.Patterns {
			name := pattern.String()
			info, _ := infos.get(name)
			for _, n := range h.Nodes {
				node := strings.TrimSpace(n.String())
				if node == "" {
					continue // ignore empty nodes
				}

				if strings.HasPrefix(node, "#") {
					continue
				}

				parts := strings.SplitN(node, " ", 2)
				if len(parts) != 2 {
					return nil, fmt.Errorf("%s: invalid node on app %q: %q", cfg, name, node)
				}

				key := strings.ToLower(strings.TrimSpace(parts[0]))
				value := strings.TrimSpace(parts[1])

				switch key {
				case "hostname":
					info.HostName = value
				case "user":
					info.User = value
				case "port":
					info.Port = value
				case "proxyjump":
					info.ProxyJump = value
				case "identityfile":
					info.IdentityFiles = append(info.IdentityFiles, value)
				case "forwardagent": // not used
					info.ForwardAgent = value
				case "requesttty": // not used
					info.RequestTTY = value
				case "remotecommand": // not used
					info.RemoteCommand = value
				case "sendenv": // not used
					info.SendEnv = append(info.SendEnv, value)
				case "setenv": // not used
					info.SetEnv = append(info.SetEnv, value)
				case "include":
					// TODO: Handle glob (Include supports dir/* format)
					if strings.Contains(value, "*") {
						continue
					}

					path, err := ExpandPath(value)
					if err != nil {
						return nil, err
					}
					included, err := parseFileInternal(path)
					if err != nil {
						return nil, err
					}
					infos.set(name, info)
					infos = merge(infos, included)
					info, _ = infos.get(name)
				}
			}

			infos.set(name, info)
		}
	}

	return infos, nil
}

func split(m *hostinfoMap) (*hostinfoMap, *hostinfoMap) {
	wildcards := newHostinfoMap()
	hosts := newHostinfoMap()
	_ = m.forEach(func(k string, v hostinfo, _ error) error {
		// FWIW the lib always returns at least one * section... no idea why.
		if strings.Contains(k, "*") {
			wildcards.set(k, v)
			return nil
		}
		hosts.set(k, v)
		return nil
	})
	return wildcards, hosts
}

func merge(m1, m2 *hostinfoMap) *hostinfoMap {
	result := newHostinfoMap()

	_ = m1.forEach(func(k string, v hostinfo, _ error) error {
		vv, ok := m2.get(k)
		if !ok {
			result.set(k, v)
			return nil
		}
		result.set(k, mergeHostinfo(v, vv))
		return nil
	})

	_ = m2.forEach(func(k string, v hostinfo, _ error) error {
		if _, ok := m1.get(k); !ok {
			result.set(k, v)
		}
		return nil
	})
	return result
}

func mergeHostinfo(h1 hostinfo, h2 hostinfo) hostinfo {
	if h1.Port != "" {
		h2.Port = h1.Port
	}
	if h1.HostName != "" {
		h2.HostName = h1.HostName
	}
	if h1.User != "" {
		h2.User = h1.User
	}
	if h1.ProxyJump != "" {
		h2.ProxyJump = h1.ProxyJump
	}

	h2.IdentityFiles = append(h2.IdentityFiles, h1.IdentityFiles...)
	if h1.ForwardAgent != "" {
		h2.ForwardAgent = h1.ForwardAgent
	}
	if h1.RequestTTY != "" {
		h2.RequestTTY = h1.RequestTTY
	}
	if h1.RemoteCommand != "" {
		h2.RemoteCommand = h1.RemoteCommand
	}
	h2.SendEnv = append(h2.SendEnv, h1.SendEnv...)
	h2.SetEnv = append(h2.SetEnv, h1.SetEnv...)

	return h2
}

func parseFileInternal(path string) (*hostinfoMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config: %w", err)
	}
	defer f.Close()
	return parseInternal(f, path)
}

func ExpandPath(p string) (string, error) {
	if !strings.HasPrefix(p, "~/") {
		return p, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to expand path: %q: %w", p, err)
	}

	return filepath.Join(home, strings.TrimPrefix(p, "~/")), nil
}
