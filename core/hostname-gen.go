package core

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Separate hosts with newline and space/tab
func EvaluateInventory(
	shell string,
	context string,
	input string,
	serverEnvs []string,
	userEnvs []string,
) ([]string, error) {
	args := strings.SplitN(shell, " ", 2)
	shellProgram := args[0]
	shellFlag := append(args[1:], input)

	cmd := exec.Command(shellProgram, shellFlag...)
	cmd.Env = append(os.Environ(), serverEnvs...)
	cmd.Env = append(cmd.Env, userEnvs...)

	cmd.Dir = filepath.Dir(context)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return []string{}, &InventoryEvalFailed{Err: string(out)}
	}

	trimmedOutput := strings.TrimSpace(string(out))
	trimmedOutput = strings.Trim(trimmedOutput, "\t")
	trimmedOutput = strings.Trim(trimmedOutput, "\r")
	trimmedOutput = strings.TrimPrefix(trimmedOutput, "\n")
	trimmedOutput = strings.TrimSuffix(trimmedOutput, "\n")

	output := strings.Fields(trimmedOutput)
	hosts := []string{}
	for _, h := range output {
		if h != "" {
			hosts = append(hosts, h)
		}
	}

	return hosts, nil
}

type HostPart struct {
	TokenType string
	Literal   any
}

type HostString struct {
	Value string
}

type HostRange struct {
	Start string
	End   string
	Step  string
}

func EvaluateRange(input string) ([]string, error) {
	ast, err := buildRangeAST(input)
	if err != nil {
		return []string{}, err
	}

	hosts, err := buildHosts(ast)
	if err != nil {
		return []string{}, err
	}

	return hosts, nil
}

func buildRangeAST(input string) ([]any, error) {
	parts := []any{}
	part := HostString{Value: ""}
	i := 0
	for i < len(input) {
		if string(input[i]) == "[" {
			r, j, err := readRange(input, i+1)
			if err != nil {
				return []any{}, err
			}

			parts = append(parts, part)
			parts = append(parts, r)
			part.Value = ""

			i = j
			continue
		} else {
			part.Value += string(input[i])
		}

		i += 1
	}

	if part.Value != "" {
		parts = append(parts, part)
	}

	return parts, nil
}

func buildHosts(parts []any) ([]string, error) {
	host := ""
	hosts := []string{host}

	for i := 0; i < len(parts); i++ {
		switch v := parts[i].(type) {
		case HostString:
			// append to all hosts
			for j := 0; j < len(hosts); j++ {
				hosts[j] += v.Value
			}
		case HostRange:
			subs, err := expandRange(v)
			if err != nil {
				return []string{}, err
			}
			var temp = []string{}
			for j := 0; j < len(subs); j++ {
				for k := 0; k < len(hosts); k++ {
					temp = append(temp, hosts[k]+subs[j])
				}
			}
			hosts = temp
		}
	}

	return hosts, nil
}

type RangeState uint64

const (
	Start = 0
	End   = 1
	Step  = 2
)

func readRange(input string, i int) (HostRange, int, error) {
	r := HostRange{
		Start: "",
		End:   "",
		Step:  "1",
	}

	var state RangeState
	for i < len(input) {
		if state > 2 {
			// malformed range
			return HostRange{}, i, errors.New("parsing hosts failed")
		}

		if string(input[i]) == "]" {
			// end of range
			i += 1
			break
		}

		if IsDigit(string(input[i])) {
			switch state {
			case Start:
				r.Start += string(input[i])
			case End:
				r.End += string(input[i])
			case Step:
				r.Step = string(input[i])
			}
		} else if string(input[i]) == ":" {
			state += 1
		} else {
			return HostRange{}, i, fmt.Errorf("parsing hosts failed, found %s in range", string(input[i]))
		}

		i += 1
	}

	if r.Start == "" {
		return HostRange{}, i, errors.New("parsing hosts failed, missing start range")
	}

	if r.End == "" {
		return HostRange{}, i, errors.New("parsing hosts failed, missing end range")
	}

	s, err := strconv.Atoi(r.Start)
	if err != nil {
		return HostRange{}, i, errors.New("start is not a number")
	}

	e, err := strconv.Atoi(r.End)
	if err != nil {
		return HostRange{}, i, errors.New("end is not a number")
	}

	if s > e {
		return HostRange{}, i, errors.New("parsing hosts failed, start cannot be greater than end")
	}

	return r, i, nil
}

func expandRange(hr HostRange) ([]string, error) {
	padLen, err := countPadding(hr.Start)
	if err != nil {
		return []string{}, err
	}

	dStart, err := strconv.Atoi(hr.Start)
	if err != nil {
		return []string{}, err
	}
	dEnd, err := strconv.Atoi(hr.End)
	if err != nil {
		return []string{}, err
	}
	dStep, err := strconv.Atoi(hr.Step)
	if err != nil {
		return []string{}, err
	}

	if dStep <= 0 {
		return []string{}, errors.New("parsing hosts failed, step less than 1")
	}

	if dEnd < dStart {
		return []string{}, errors.New("parsing hosts failed, end lower than start")
	}

	hosts := make([]string, 0, 1+(dEnd-dStart)/dStep)
	for dStart <= dEnd {
		if padLen == 0 {
			hosts = append(hosts, strconv.Itoa(dStart))
		} else {
			h := strconv.Itoa(dStart)
			if padLen >= len(h) {
				hosts = append(hosts, leftZeroPad(h, padLen+1))
			} else {
				hosts = append(hosts, h)
			}
		}

		dStart += dStep
	}

	return hosts, nil
}

func countPadding(s string) (int, error) {
	origLen := len(s)
	digit, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	digitLen := lenDigit(digit)

	return origLen - digitLen, nil
}

func lenDigit(i int) int {
	if i == 0 {
		return 1
	}
	count := 0
	for i != 0 {
		i /= 10
		count++
	}
	return count
}

func leftZeroPad(s string, padWidth int) string {
	return fmt.Sprintf(fmt.Sprintf("%%0%ds", padWidth), s)
}
