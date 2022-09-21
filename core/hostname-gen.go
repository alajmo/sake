package core

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func ExpandHostNames(input string) ([]string, error) {
	// TODO:
	// - [x] Check if enclosed by $(), if run cmd
	// - If contains brackets, expand to list of strings
	// - else return single host

	if IsCmd(input) {
		// TODO: Handle cmd
		return EvaluateInventory(input)
	} else if IsRange(input) {
		// TODO: Handle range
		return EvaluateRange(input)
	}

	return []string{input}, nil
}

func IsCmd(input string) bool {
	if strings.HasPrefix(input, "$(") && strings.HasSuffix(input, ")") {
		return true
	}
	return false
}

func IsRange(input string) bool {
	return strings.Contains(input, "[")
}

// Separate hosts with newline and space/tab
func EvaluateInventory(input string) ([]string, error) {
	// TODO: Add in env variables from server
	input = strings.TrimPrefix(input, "$(")
	input = strings.TrimSuffix(input, ")")

	cmd := exec.Command("sh", "-c", input)
	cmd.Env = os.Environ()
	stdout, err := cmd.Output()
	if err != nil {
		return []string{}, &InventoryEvalFailed{Err: err.Error()}
	}

	trimmedOutput := strings.TrimSpace(string(stdout))
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
	// leftBrackets := strings.Count(input, "[")
	// rightBrackets := strings.Count(input, "]")
	// if leftBrackets != rightBrackets {
	// 	return []string{}, fmt.Errorf("missing brackets, found %d '[' and %d ']'", leftBrackets, rightBrackets)
	// }

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
			// TODO: Only allow alpha/digit/. Valid characters for hostnames are ASCII(7)
			// letters from a to z,
			// the digits from 0 to 9,
			// and the hyphen (-).
			// A hostname may not start with a hyphen.
			// @ and : for user/port

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

func (s RangeState) String() string {
	switch s {
	case Start:
		return "Start"
	case End:
		return "End"
	case Step:
		return "Step"
	}

	return "unknown"
}

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

		if isDigit(string(input[i])) {
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

	if r.Start > r.End {
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

func isDigit(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
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
