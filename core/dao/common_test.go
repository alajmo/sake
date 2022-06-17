package dao

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseEnvsYAML(t *testing.T) {
	var data = `
env:
  foo: "bar"
  hello: hello world
  script: $(echo 123)
  script: $echo 123)
`
	configYAML := ConfigYAML{}
	err := yaml.Unmarshal([]byte(data), &configYAML)
	if err != nil {
		t.Fatalf("%q", err)
	}

	envs := ParseNodeEnv(configYAML.Env)

	wanted := []string{
		"foo=bar",
		"hello=hello world",
		"script=$(echo 123)",
		"script=$echo 123)",
	}

	for i := range wanted {
		if envs[i] != wanted[i] {
			t.Fatalf(`Wanted: %q, Found: %q`, wanted[i], envs[i])
		}
	}
}

func TestEvaluateEnv(t *testing.T) {
	envs := []string{
		"foo=bar",
		"hello=hello world",
		"script=$(echo 123)",
		"script=$echo 123)",
	}

	found, err := EvaluateEnv(envs)
	if err != nil {
		t.Fatalf("%q", err)
	}

	wanted := []string{
		"foo=bar",
		"hello=hello world",
		"script=123\n",
		"script=$echo 123)",
	}

	for i := range wanted {
		if found[i] != wanted[i] {
			t.Fatalf(`Wanted: %q, Found: %q`, wanted[i], found[i])
		}
	}
}

func TestMergeEnvs(t *testing.T) {
	envsA := []string{
		"foo=bar",
		"hello=world",
	}

	envsB := []string{
		"foo=hej",
		"script=$(echo 123)",
		"xyz=zyx",
	}

	merged := MergeEnvs(envsA, envsB)

	wanted := []string{
		"foo=bar",
		"hello=world",
		"script=$(echo 123)",
		"xyz=zyx",
	}

	for i := range wanted {
		if merged[i] != wanted[i] {
			t.Fatalf(`Wanted: %q, Found: %q`, wanted[i], merged[i])
		}
	}
}

func TestSelectFirstNonEmpty(t *testing.T) {
	values := []string{
		"",
		"",
		"foo",
		"hello",
		"",
	}

	firstNonEmpty := SelectFirstNonEmpty(values...)

	if firstNonEmpty != "foo" {
		t.Fatalf(`Wanted: %q, Found: %q`, "foo", firstNonEmpty)
	}
}
