package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const ANSI = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var RE = regexp.MustCompile(ANSI)

func Strip(str string) string {
	return RE.ReplaceAllString(str, "")
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Intersection(a []string, b []string) []string {
	var i []string
	for _, s := range a {
		if StringInSlice(s, b) {
			i = append(i, s)
		}
	}

	return i
}

func FindFileInParentDirs(path string, files []string) (string, error) {
	for _, file := range files {
		pathToFile := filepath.Join(path, file)

		if _, err := os.Stat(pathToFile); err == nil {
			return pathToFile, nil
		}
	}

	parentDir := filepath.Dir(path)

	if parentDir == "/" {
		return "", &ConfigNotFound{files}
	}

	return FindFileInParentDirs(parentDir, files)
}

// Get the absolute path
// Need to support following path types:
//
//	lala/land
//	./lala/land
//	../lala/land
//	/lala/land
//	$HOME/lala/land
//	~/lala/land
//	~root/lala/land
func GetAbsolutePath(configDir string, path string, name string) (string, error) {
	path = os.ExpandEnv(path)

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	homeDir := usr.HomeDir

	// TODO: Remove any .., make path absolute and then cut of configDir
	if path == "~" {
		path = homeDir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	} else if len(path) > 0 && filepath.IsAbs(path) { // TODO: Rewrite this
	} else if len(path) > 0 {
		path = filepath.Join(configDir, path)
	} else {
		path = filepath.Join(configDir, name)
	}

	return path, nil
}

// FormatShell returns the shell program and associated command flag
func FormatShell(shell string) string {
	s := strings.Split(shell, " ")

	if len(s) > 1 { // User provides correct flag, bash -c, /bin/bash -c, /bin/sh -c
		return shell
	} else if strings.Contains(shell, "bash") { // bash, /bin/bash
		return shell + " -c"
	} else if strings.Contains(shell, "zsh") { // zsh, /bin/zsh
		return shell + " -c"
	} else if strings.Contains(shell, "sh") { // sh, /bin/sh
		return shell + " -c"
	} else if strings.Contains(shell, "python") { // python, /bin/python
		return shell + " -c"
	} else if strings.Contains(shell, "node") { // node, /bin/node
		return shell + " -e"
	}

	return shell
}

// Used when creating pointers to literal. Useful when you want set/unset attributes.
func Ptr[T any](t T) *T {
	return &t
}

func StringsToErrors(str []string) []error {
	errs := []error{}
	for _, s := range str {
		errs = append(errs, errors.New(s))
	}

	return errs
}

func AnyToString(s any) string {
	var v string

	switch s := s.(type) {
	case nil:
		v = ""
	case int:
		v = strconv.Itoa(s)
	case bool:
		v = strconv.FormatBool(s)
	case string:
		v = s
	default:
		return ""
	}

	return v
}

func StringToBool(s string) bool {
	ss := strings.ToLower(strings.TrimSpace(s))
	return ss == "true" || ss == "yes"
}

// TODO: Don't include this in build
func DebugPrint(data any) {
	s, _ := json.MarshalIndent(data, "", "\t")
	fmt.Print(string(s))
	fmt.Println()
}

// Parse host, for instance : user@hostname:22
func ParseHostName(hostname string, defaultUser string, defaultPort uint16) (string, string, uint16, error) {
	if strings.Contains(hostname, "/") {
		return "", "", 22, fmt.Errorf("unexpected slash in the host %s", hostname)
	}

	var user string
	host := hostname
	var port uint16

	// Remove extra "ssh://" schema
	if len(host) > 6 && host[:6] == "ssh://" {
		host = host[6:]
	}

	// Split by the last "@", since there may be an "@" in the username
	if at := strings.LastIndex(host, "@"); at != -1 {
		user = host[:at]
		host = host[at+1:]
	} else {
		user = defaultUser
	}

	// Checks only host, not user or port
	// Valid:
	//  - 192.168.0.1
	//  - 2001:3984:3989::10
	//
	// Invalid:
	//  - user@192.168.0.1:port
	//  - user@[2001:3984:3989::10]:port
	ip := net.ParseIP(host)
	if ip != nil {
		ipp := ip.To4()
		if ipp != nil {
			return user, ipp.String(), defaultPort, nil
		}

		ipp = ip.To16()
		if ipp != nil {
			return user, ipp.String(), defaultPort, nil
		}
	}

	// We check this when user has specified port in ip address:
	// Valid:
	//  - 192.168.0.1:port
	//  - [2001:3984:3989::10]:port
	//
	// Invalid:
	//  - [192.168.0.1]:port
	//  - 2001:3984:3989::10:port
	switch getIPType(host) {
	case 4:
		if strings.Contains(host, ":") {
			lastInd := strings.LastIndex(host, ":")
			p, err := strconv.ParseUint(host[lastInd+1:], 10, 16)
			if err != nil {
				return "", "", 22, err
			}
			host = host[:lastInd]
			port = uint16(p)
		} else {
			port = defaultPort
		}
		return user, host, port, nil
	case 6:
		if strings.Contains(host, "[") && strings.Contains(host, "]") {
			if at := strings.LastIndex(host, ":"); at != -1 {
				p, err := strconv.ParseUint(host[at+1:], 10, 16)
				if err != nil {
					return "", "", 22, fmt.Errorf("failed to parse %s", hostname)
				}
				host = host[1 : at-1]
				port = uint16(p)
			}
		}

		// Check if has brackets and port, remove brackets and add new port
		return user, host, port, nil
	}

	if port == 0 {
		port = defaultPort
	}

	return user, host, port, nil
}

func getIPType(ip string) uint {
	for i := 0; i < len(ip); i++ {
		switch ip[i] {
		case '.':
			return 4
		case ':':
			return 6
		}
	}

	return 0
}

func IsDigit(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func SplitString(s, sep string) []string {
	if len(s) == 0 {
		return []string{}
	}
	return strings.Split(s, sep)
}

func GetFirstExistingFile(files ...string) string {
	for _, file := range files {
		expandedFile := os.ExpandEnv(file)
		expandedFile, err := expandTilde(expandedFile)
		if err != nil {
			continue
		}
		if _, err := os.Stat(expandedFile); err == nil {
			return expandedFile
		}
	}
	return ""
}

func expandTilde(path string) (string, error) {
	if path[0] != '~' {
		return path, nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}
