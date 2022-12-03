package core

import (
	"testing"

	"github.com/alajmo/sake/core/test"
)

func TestEvaluateInventory(t *testing.T) {
	// IP inventory 1 host
	input := `echo 192.168.0.1`
	hosts, err := EvaluateInventory("sh -c", "", input, []string{}, []string{})
	test.CheckErr(t, err)
	wanted := []string{"192.168.0.1"}
	for i := range wanted {
		test.CheckEqS(t, hosts[i], wanted[i])
	}

	// IP inventory 2 hosts splitted on space
	input = `echo "192.168.0.1 192.168.0.2"`
	hosts, err = EvaluateInventory("sh -c", "", input, []string{}, []string{})
	test.CheckErr(t, err)
	wanted = []string{"192.168.0.1", "192.168.0.2"}
	for i := range wanted {
		test.CheckEqS(t, hosts[i], wanted[i])
	}

	// IP inventory 2 hosts splitted on newline
	input = `echo "192.168.0.1\n192.168.0.2"`
	hosts, err = EvaluateInventory("sh -c", "", input, []string{}, []string{})
	test.CheckErr(t, err)
	wanted = []string{"192.168.0.1", "192.168.0.2"}
	for i := range wanted {
		test.CheckEqS(t, hosts[i], wanted[i])
	}

	// IP inventory 2 hosts splitted on tab
	input = `echo "192.168.0.1\t192.168.0.2"`
	hosts, err = EvaluateInventory("sh -c", "", input, []string{}, []string{})
	test.CheckErr(t, err)
	wanted = []string{"192.168.0.1", "192.168.0.2"}
	for i := range wanted {
		test.CheckEqS(t, hosts[i], wanted[i])
	}
}

func TestEvaluateRange(t *testing.T) {
	// SINGLE IP
	input := "192.168.0.1"
	hosts, err := EvaluateRange(input)
	test.CheckErr(t, err)
	wanted := []string{"192.168.0.1"}
	for i := range wanted {
		test.CheckEqS(t, hosts[i], wanted[i])
	}

	// RANGE

	// IP 1 range
	input = "192.168.0.[1:2]"
	hosts, err = EvaluateRange(input)
	test.CheckErr(t, err)
	wanted = []string{"192.168.0.1", "192.168.0.2"}
	for i := range wanted {
		test.CheckEqS(t, hosts[i], wanted[i])
	}

	// IP 1 range
	input = "192.168.0.[1:2].33"
	hosts, err = EvaluateRange(input)
	test.CheckErr(t, err)
	wanted = []string{"192.168.0.1.33", "192.168.0.2.33"}
	for i := range wanted {
		test.CheckEqS(t, hosts[i], wanted[i])
	}

	// IP 1 range with padding
	input = "192.168.0.[09:12].33"
	hosts, err = EvaluateRange(input)
	test.CheckErr(t, err)
	wanted = []string{"192.168.0.09.33", "192.168.0.10.33", "192.168.0.11.33", "192.168.0.12.33"}
	for i := range wanted {
		test.CheckEqS(t, hosts[i], wanted[i])
	}

	// IP 1 range with step
	input = "192.168.0.[1:4:2].33"
	hosts, err = EvaluateRange(input)
	test.CheckErr(t, err)
	wanted = []string{"192.168.0.1.33", "192.168.0.3.33"}
	for i := range wanted {
		test.CheckEqS(t, hosts[i], wanted[i])
	}

	// IP 1 range with step and padding
	input = "192-[01:4:2].33"
	hosts, err = EvaluateRange(input)
	test.CheckErr(t, err)
	wanted = []string{"192-01.33", "192-03.33"}
	for i := range wanted {
		test.CheckEqS(t, hosts[i], wanted[i])
	}

	// IP 2 ranges
	input = "192.[0:1].0.[2:3]"
	hosts, err = EvaluateRange(input)
	test.CheckErr(t, err)
	wanted = []string{
		"192.0.0.2",
		"192.1.0.2",
		"192.0.0.3",
		"192.1.0.3",
	}
	for i := range wanted {
		test.CheckEqS(t, hosts[i], wanted[i])
	}

	// Malformed ranges
	_, err = EvaluateRange("192.[2:1].0")
	test.IsError(t, err)

	_, err = EvaluateRange("192.[0].0")
	test.IsError(t, err)

	_, err = EvaluateRange("192.[0.0")
	test.IsError(t, err)

	_, err = EvaluateRange("192.[1:2:1:2]")
	test.IsError(t, err)
}
