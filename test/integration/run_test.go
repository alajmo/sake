package integration

import (
	"fmt"
	"testing"
)

var cases = []TemplateTest{
	// list/describe
	{
		TestName: "List tags",
		TestCmd:  "go run ../../main.go list tags",
		WantErr:  false,
	},
	{
		TestName: "List servers",
		TestCmd:  "go run ../../main.go list servers",
		WantErr:  false,
	},
	{
		TestName: "Describe servers",
		TestCmd:  "go run ../../main.go describe servers",
		WantErr:  false,
	},
	{
		TestName: "List tasks",
		TestCmd:  "go run ../../main.go list tasks",
		WantErr:  false,
	},
	{
		TestName: "Describe tasks",
		TestCmd:  "go run ../../main.go describe tasks",
		WantErr:  false,
	},

	// basic
	{
		TestName: "Ping all servers",
		TestCmd:  "go run ../../main.go run ping -t reachable",
		WantErr:  false,
	},
	{
		TestName: "Multiple commands",
		TestCmd:  "go run ../../main.go run info -t prod",
		WantErr:  false,
	},

	// env
	{
		TestName: "Simple Envs",
		TestCmd:  "go run ../../main.go run env -t reachable",
		WantErr:  false,
	},
	{
		TestName: "Reference Envs",
		TestCmd:  "go run ../../main.go run env-complex -t reachable",
		WantErr:  false,
	},
	{
		TestName: "Default Envs",
		TestCmd:  "go run ../../main.go run env-default -t reachable",
		WantErr:  false,
	},

	// nested tasks
	{
		TestName: "Nested tasks",
		TestCmd:  "go run ../../main.go run d -t reachable",
		WantErr:  false,
	},

	// work_dir
	{
		TestName: "Work Dir 1",
		TestCmd:  "go run ../../main.go run work-dir-1 -t reachable",
		WantErr:  false,
	},
	{
		TestName: "Work Dir 2",
		TestCmd:  "go run ../../main.go run work-dir-2 -t reachable",
		WantErr:  false,
	},
	{
		TestName: "Work Dir 3",
		TestCmd:  "go run ../../main.go run work-dir-3 -t reachable",
		WantErr:  false,
	},

	// spec
	{
		TestName: "fatal false",
		TestCmd:  "go run ../../main.go run fatal -t reachable",
		WantErr:  false,
	},
	{
		TestName: "fatal true",
		TestCmd:  "go run ../../main.go run fatal-true -t reachable",
		WantErr:  true,
	},
	{
		TestName: "ignore_errors false",
		TestCmd:  "go run ../../main.go run errors -t reachable",
		WantErr:  false,
	},
	{
		TestName: "ignore_errors true",
		TestCmd:  "go run ../../main.go run errors-true -t reachable",
		WantErr:  false,
	},
	{
		TestName: "unreachable false",
		TestCmd:  "go run ../../main.go run unreachable -a",
		WantErr:  true,
	},
	{
		TestName: "unreachable true",
		TestCmd:  "go run ../../main.go run unreachable-true -a",
		WantErr:  false,
	},
	{
		TestName: "omit_empty false",
		TestCmd:  "go run ../../main.go run empty -t reachable",
		WantErr:  false,
	},
	{
		TestName: "omit_empty true",
		TestCmd:  "go run ../../main.go run empty-true -t reachable",
		WantErr:  false,
	},
	{
		TestName: "output",
		TestCmd:  "go run ../../main.go run output -t reachable",
		WantErr:  false,
	},

	// exec
	{
		TestName: "Run exec command",
		TestCmd:  "go run ../../main.go exec 'echo 123' -t reachable",
		WantErr:  false,
	},
}

func TestRunCmd(t *testing.T) {
	for i := range cases {
		cases[i].Golden = fmt.Sprintf("golden-%d.stdout", i)
		cases[i].Index = i

		t.Run(cases[i].TestName, func(t *testing.T) {
			Run(t, cases[i])
		})
	}
}
