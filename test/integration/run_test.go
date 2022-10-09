package integration

import (
	"fmt"
	"testing"
)

var cases = []TemplateTest{
	// list tags
	{
		TestName: "List tags",
		TestCmd:  "go run ../../main.go list tags",
		WantErr:  false,
	},

	// list tasks
	{
		TestName: "List tasks",
		TestCmd:  "go run ../../main.go list tasks",
		WantErr:  false,
	},

	// list servers
	{
		TestName: "List servers",
		TestCmd:  "go run ../../main.go list servers",
		WantErr:  false,
	},
	{
		TestName: "List servers filter on list hosts",
		TestCmd:  "go run ../../main.go list servers list",
		WantErr:  false,
	},
	{
		TestName: "List servers filter on range hosts",
		TestCmd:  "go run ../../main.go list servers range",
		WantErr:  false,
	},
	{
		TestName: "List servers filter on inventory hosts",
		TestCmd:  "go run ../../main.go list servers inv",
		WantErr:  false,
	},

	// describe servers
	{
		TestName: "Describe servers",
		TestCmd:  "go run ../../main.go describe servers",
		WantErr:  false,
	},
	{
		TestName: "Describe servers filter on list hosts",
		TestCmd:  "go run ../../main.go describe servers list",
		WantErr:  false,
	},
	{
		TestName: "Describe servers filter on range hosts",
		TestCmd:  "go run ../../main.go describe servers range",
		WantErr:  false,
	},
	{
		TestName: "Describe servers filter on inventory hosts",
		TestCmd:  "go run ../../main.go describe servers inv",
		WantErr:  false,
	},

	// describe tasks
	{
		TestName: "Describe tasks",
		TestCmd:  "go run ../../main.go describe tasks",
		WantErr:  false,
	},

	// run basic tasks
	{
		TestName: "Ping all servers",
		TestCmd:  "go run ../../main.go run ping -S -t reachable",
		WantErr:  false,
	},
	{
		TestName: "Multiple commands",
		TestCmd:  "go run ../../main.go run info -S -t prod",
		WantErr:  false,
	},
	{
		TestName: "Filter by hosts server using server name",
		TestCmd:  "go run ../../main.go run info -S -s list-1",
		WantErr:  false,
	},
	{
		TestName: "Filter by hosts server using range index",
		TestCmd:  "go run ../../main.go run info -S -s 'list[0]'",
		WantErr:  false,
	},
	{
		TestName: "Filter by hosts server",
		TestCmd:  "go run ../../main.go run info -S -s 'list[0:2]'",
		WantErr:  false,
	},
	{
		TestName: "Filter by host regex",
		TestCmd:  "go run ../../main.go run info -S -r '172.24.2.(2|4)'",
		WantErr:  false,
	},
	{
		TestName: "Limit to 2 servers",
		TestCmd:  "go run ../../main.go run ping -S -t reachable -l 2",
		WantErr:  false,
	},
	{
		TestName: "Limit to 50 percent servers",
		TestCmd:  "go run ../../main.go run ping -S -t reachable -L 50",
		WantErr:  false,
	},
	{
		TestName: "Filter by inverting on tag unreachable",
		TestCmd:  "go run ../../main.go run ping -S -t unreachable -v",
		WantErr:  false,
	},

	// run tasks and display env
	{
		TestName: "Simple Envs",
		TestCmd:  "go run ../../main.go run env -S -t reachable",
		WantErr:  false,
	},
	{
		TestName: "Reference Envs",
		TestCmd:  "go run ../../main.go run env-complex -S -t reachable",
		WantErr:  false,
	},
	{
		TestName: "Default Envs",
		TestCmd:  "go run ../../main.go run env-default -S -t reachable",
		WantErr:  false,
	},

	// run nested tasks
	{
		TestName: "Nested tasks",
		TestCmd:  "go run ../../main.go run d -S -t reachable",
		WantErr:  false,
	},

	// run tasks and modify work dir
	{
		TestName: "Work Dir 1",
		TestCmd:  "go run ../../main.go run work-dir-1 -S -t reachable",
		WantErr:  false,
	},
	{
		TestName: "Work Dir 2",
		TestCmd:  "go run ../../main.go run work-dir-2 -S -t reachable",
		WantErr:  false,
	},
	{
		TestName: "Work Dir 3",
		TestCmd:  "go run ../../main.go run work-dir-3 -S -t reachable",
		WantErr:  false,
	},

	// Tests for running tasks with various specs
	{
		TestName: "fatal false",
		TestCmd:  "go run ../../main.go run fatal -S -t reachable",
		WantErr:  false,
	},
	{
		TestName: "fatal true",
		TestCmd:  "go run ../../main.go run fatal-true -S -t reachable",
		WantErr:  true,
	},
	{
		TestName: "ignore_errors false",
		TestCmd:  "go run ../../main.go run errors -S -t reachable",
		WantErr:  false,
	},
	{
		TestName: "ignore_errors true",
		TestCmd:  "go run ../../main.go run errors-true -S -t reachable",
		WantErr:  false,
	},
	{
		TestName: "unreachable false",
		TestCmd:  "go run ../../main.go run unreachable -S -a",
		WantErr:  true,
	},
	{
		TestName: "unreachable true",
		TestCmd:  "go run ../../main.go run unreachable-true -S -a",
		WantErr:  false,
	},
	{
		TestName: "omit_empty false",
		TestCmd:  "go run ../../main.go run empty -S -t reachable",
		WantErr:  false,
	},
	{
		TestName: "omit_empty true",
		TestCmd:  "go run ../../main.go run empty-true -S -t reachable",
		WantErr:  false,
	},
	{
		TestName: "output",
		TestCmd:  "go run ../../main.go run output -S -t reachable",
		WantErr:  false,
	},

	// exec
	{
		TestName: "Run exec command",
		TestCmd:  "go run ../../main.go exec 'echo 123' -S -t reachable",
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
