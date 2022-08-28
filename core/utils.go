package core

import (
	"encoding/json"
	"errors"
	"fmt"
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
//		lala/land
//		./lala/land
//		../lala/land
//		/lala/land
//		$HOME/lala/land
//		~/lala/land
//		~root/lala/land
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
	} else if strings.Contains(shell, "node") { // node, /bin/node
		return shell + " -e"
	} else if strings.Contains(shell, "python") { // python, /bin/python
		return shell + " -c"
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

func DebugPrint(data any) {
	s, _ := json.MarshalIndent(data, "", "\t")
	fmt.Print(string(s))
	fmt.Println()
}

// Parse host, for instance : user@hostname
func ParseHostName(hostname string, defaultUser string, defaultPort uint16) (string, string, uint16, error) {
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

	if strings.Contains(host, "/") {
		return "", "", 22, fmt.Errorf("unexpected slash in the host %s", hostname)
	}

	if strings.Contains(hostname, ":") {
		lastInd := strings.LastIndex(host, ":")
		p, err := strconv.ParseInt(host[lastInd+1:], 0, 16)
		if err != nil {
			return "", "", 22, err
		}
		port = uint16(p)

		host = host[:lastInd]
	} else {
		port = defaultPort
	}

	return user, host, port, nil
}
