package dao

import (
	"testing"

	"github.com/alajmo/sake/core/test"
)

func TestFilterServers(t *testing.T) {
	s1 := Server{Name: "s1", Group: "s1", Tags: []string{"t1", "t2"}}
	s2 := Server{Name: "s2", Group: "s2", Tags: []string{"t3"}}
	s3 := Server{Name: "s3", Group: "s3", Tags: []string{"t2", "t3"}}
	s4 := Server{Name: "s4", Group: "s4", Tags: []string{"t8"}}
	s5 := Server{Name: "s5", Group: "s5", Tags: []string{"t1", "t2"}}

	c := Config{
		Servers: []Server{s1, s2, s3, s4, s5},
	}

	// TODO: Fails
	ss, err := c.FilterServers(false, []string{"s1", "s2"}, []string{"t1"})
	test.CheckErr(t, err)
	wanted := []string{"s1", "s2", "s5"}
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	ss, err = c.FilterServers(false, []string{"s5"}, []string{})
	test.CheckErr(t, err)
	wanted = []string{"s5"}
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	ss, err = c.FilterServers(false, []string{}, []string{"t1"})
	test.CheckErr(t, err)
	wanted = []string{"s1", "s5"}
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	ss, err = c.FilterServers(true, []string{}, []string{})
	test.CheckErr(t, err)
	wanted = []string{"s1", "s2", "s3", "s4", "s5"}
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}
}

func TestGetServersByTags(t *testing.T) {
	s1 := Server{Name: "s1", Tags: []string{"t1", "t2"}}
	s2 := Server{Name: "s2", Tags: []string{"t3"}}
	s3 := Server{Name: "s3", Tags: []string{"t2", "t3"}}
	s4 := Server{Name: "s4", Tags: []string{"t8"}}
	s5 := Server{Name: "s5", Tags: []string{"t1"}}

	c := Config{
		Servers: []Server{s1, s2, s3, s4, s5},
	}

	ss, err := c.GetServersByTags([]string{"t1"})
	test.CheckErr(t, err)
	wanted := []string{"s1", "s5"}
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	ss, err = c.GetServersByTags([]string{"t1", "t2"})
	test.CheckErr(t, err)
	wanted = []string{"s1"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}
}

func TestGetTagAssocations(t *testing.T) {
	s1 := Server{Name: "s1", Tags: []string{"t1", "t2"}}
	s2 := Server{Name: "s2", Tags: []string{"t3"}}
	s3 := Server{Name: "s3", Tags: []string{"t2", "t3"}}
	s4 := Server{Name: "s4", Tags: []string{"t8"}}
	s5 := Server{Name: "s5", Tags: []string{"t1"}}

	c := Config{
		Servers: []Server{s1, s2, s3, s4, s5},
	}

	ss, err := c.GetTagAssocations([]string{"t1"})
	test.CheckErr(t, err)

	wanted := Tag{
		Name:    "t1",
		Servers: []string{"s1", "s5"},
	}
	test.CheckEqualStringArr(t, ss[0].Servers, wanted.Servers)
}
