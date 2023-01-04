package integration

import (
	"flag"
	"fmt"
	"testing"
)

var par = flag.Bool("par", true, "run tests in parallel")

var cases = []TemplateTest{
	// list tags
	{
		TestName: "List tags",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go list tags`,
		WantErr:  false,
	},

	// list specs
	{
		TestName: "List specs",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go list specs`,
		WantErr:  false,
	},

	// list targets
	{
		TestName: "List targets",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go list targets`,
		WantErr:  false,
	},

	// list tasks
	{
		TestName: "List tasks",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go list tasks`,
		WantErr:  false,
	},

	// list servers
	{
		TestName: "List servers",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go list servers`,
		WantErr:  false,
	},
	{
		TestName: "List servers filter on list hosts",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go list servers list`,
		WantErr:  false,
	},
	{
		TestName: "List servers filter on range hosts",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go list servers range`,
		WantErr:  false,
	},
	{
		TestName: "List servers filter on inventory hosts",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go list servers inv`,
		WantErr:  false,
	},

	// describe specs
	{
		TestName: "Describe specs",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go describe specs`,
		WantErr:  false,
	},

	// describe targets
	{
		TestName: "Describe targets",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go describe targets`,
		WantErr:  false,
	},

	// describe servers
	{
		TestName: "Describe servers",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go describe servers`,
		WantErr:  false,
	},
	{
		TestName: "Describe servers filter on list hosts",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go describe servers list`,
		WantErr:  false,
	},
	{
		TestName: "Describe servers filter on range hosts",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go describe servers range`,
		WantErr:  false,
	},
	{
		TestName: "Describe servers filter on inventory hosts",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go describe servers inv`,
		WantErr:  false,
	},

	// describe tasks
	{
		TestName: "Describe tasks",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go describe tasks`,
		WantErr:  false,
	},

	// run basic tasks
	{
		TestName: "Ping all servers",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run ping -q -t reachable`,
		WantErr:  false,
	},
	{
		TestName: "Multiple commands",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run info -q -t prod`,
		WantErr:  false,
	},
	{
		TestName: "Filter by hosts server using server name",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run info -q -s list-1`,
		WantErr:  false,
	},
	{
		TestName: "Filter by hosts server using range index",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run info -q -s 'list[0]'`,
		WantErr:  false,
	},
	{
		TestName: "Filter by hosts server",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run info -q -s 'list[0:2]'`,
		WantErr:  false,
	},
	{
		TestName: "Filter by host regex",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run info -q -r '172.24.2.(2|4)'`,
		WantErr:  false,
	},
	{
		TestName: "Limit to 2 servers",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run ping -q -t reachable -l 2`,
		WantErr:  false,
	},
	{
		TestName: "Limit to 50 percent servers",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run ping -q -t reachable -L 50`,
		WantErr:  false,
	},
	{
		TestName: "Filter by inverting on tag unreachable",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run ping -q -t unreachable -v`,
		WantErr:  false,
	},

	// run tasks and display env
	{
		TestName: "Simple Envs",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run env -q -t reachable`,
		WantErr:  false,
	},
	{
		TestName: "Reference Envs",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run env-complex -q -t reachable`,
		WantErr:  false,
	},
	{
		TestName: "Default Envs",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run env-default -q -t reachable`,
		WantErr:  false,
	},

	// run nested tasks
	{
		TestName: "Nested tasks",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run d -q -t reachable`,
		WantErr:  false,
	},

	// run tasks and modify work dir
	{
		TestName: "Work Dir 1",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run work-dir-1 -q -t reachable`,
		WantErr:  false,
	},
	{
		TestName: "Work Dir 2",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run work-dir-2 -q -t reachable`,
		WantErr:  false,
	},
	{
		TestName: "Work Dir 3",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run work-dir-3 -q -t reachable`,
		WantErr:  false,
	},

	// run tasks and register variables
	{
		TestName: "Register 1",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run register-1 -q -t reachable`,
		WantErr:  false,
	},
	{
		TestName: "Register 2",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run register-2 -q -t reachable`,
		WantErr:  false,
	},

	// Tests for running tasks with various specs
	{
		TestName: "fatal false",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run fatal -q -t reachable`,
		WantErr:  true,
	},
	{
		TestName: "fatal true",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run fatal-true -q -t reachable`,
		WantErr:  true,
	},
	{
		TestName: "ignore_errors false",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run errors -q -t reachable`,
		WantErr:  true,
	},
	{
		TestName: "ignore_errors true",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run errors-true -q -t reachable`,
		WantErr:  false,
	},
	{
		TestName: "unreachable false",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run unreachable -q -a`,
		WantErr:  true,
	},
	{
		TestName: "unreachable true",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run unreachable-true -o table -q -a`,
		WantErr:  false,
	},
	{
		TestName: "omit_empty false",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run empty -q -t reachable`,
		WantErr:  false,
	},
	{
		TestName: "omit_empty true",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run empty-true -q -t reachable`,
		WantErr:  false,
	},
	{
		TestName: "output",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go run output -q -t reachable`,
		WantErr:  false,
	},

	// exec
	{
		TestName: "Run exec command",
		TestCmd:  `export SAKE_USER_CONFIG="$PWD/../user-config.yaml" && go run ../../main.go exec 'echo 123' -q -t reachable`,
		WantErr:  false,
	},
}

func TestRunCmd(t *testing.T) {
	for i := range cases {
		i := i
		cases[i].Golden = fmt.Sprintf("golden-%d.stdout", i)
		cases[i].Index = i
		t.Run(cases[i].TestName, func(t *testing.T) {
			if *par {
				t.Parallel()
			}

			Run(t, cases[i])
		})
	}
}
