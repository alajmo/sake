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
	s6 := Server{Name: "s6-1", Group: "s6", Host: "192.168", Tags: []string{}}
	s7 := Server{Name: "s6-2", Group: "s6", Host: "192.169", Tags: []string{}}

	c := Config{
		Servers: []Server{s1, s2, s3, s4, s5, s6, s7},
	}

	// Server + Tag
	ss, err := c.FilterServers(false, []string{"s1", "s2"}, []string{"t1"}, "", false)
	test.CheckErr(t, err)
	wanted := []string{"s1"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	// Invert
	ss, err = c.FilterServers(false, []string{"s1", "s2"}, []string{"t1"}, "", true)
	test.CheckErr(t, err)
	wanted = []string{"s2", "s3", "s4", "s5", "s6-1", "s6-2"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	// Server
	ss, err = c.FilterServers(false, []string{"s5"}, []string{}, "", false)
	test.CheckErr(t, err)
	wanted = []string{"s5"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	// Tag
	ss, err = c.FilterServers(false, []string{}, []string{"t1"}, "", false)
	test.CheckErr(t, err)
	wanted = []string{"s1", "s5"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	// All
	ss, err = c.FilterServers(true, []string{}, []string{}, "", false)
	test.CheckErr(t, err)
	wanted = []string{"s1", "s2", "s3", "s4", "s5", "s6-1", "s6-2"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	// Regex
	ss, err = c.FilterServers(false, []string{}, []string{}, "192.(168|169)", false)
	test.CheckErr(t, err)
	wanted = []string{"s6-1", "s6-2"}
	test.CheckEqN(t, len(ss), len(wanted))
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
		Servers: []string{"s1", "s5"},
	}
	test.CheckEqualStringArr(t, ss[0].Servers, wanted.Servers)
}

func TestTargetRange(t *testing.T) {
	s1 := Server{Name: "s1", Group: "s1"}
	s2 := Server{Name: "s2", Group: "s2"}
	s3 := Server{Name: "s3", Group: "s3"}
	s4 := Server{Name: "s4", Group: "s4"}
	s5 := Server{Name: "s5-0", Group: "s5"}
	s6 := Server{Name: "s5-1", Group: "s5"}

	c := Config{
		Servers: []Server{s1, s2, s3, s4, s5, s6},
	}

	ss, err := c.FilterServers(false, []string{"s4"}, []string{}, "", false)
	test.CheckErr(t, err)
	wanted := []string{"s4"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	ss, err = c.FilterServers(false, []string{"s5"}, []string{}, "", false)
	test.CheckErr(t, err)
	wanted = []string{"s5-0", "s5-1"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	ss, err = c.FilterServers(false, []string{"s5[:]"}, []string{}, "", false)
	test.CheckErr(t, err)
	wanted = []string{"s5-0", "s5-1"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	ss, err = c.FilterServers(false, []string{"s5[0:]"}, []string{}, "", false)
	test.CheckErr(t, err)
	wanted = []string{"s5-0", "s5-1"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	ss, err = c.FilterServers(false, []string{"s5[0:10]"}, []string{}, "", false)
	test.CheckErr(t, err)
	wanted = []string{"s5-0", "s5-1"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	ss, err = c.FilterServers(false, []string{"s5[:10]"}, []string{}, "", false)
	test.CheckErr(t, err)
	wanted = []string{"s5-0", "s5-1"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	ss, err = c.FilterServers(false, []string{"s5[:3]"}, []string{}, "", false)
	test.CheckErr(t, err)
	wanted = []string{"s5-0", "s5-1"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	ss, err = c.FilterServers(false, []string{"s5[0]"}, []string{}, "", false)
	test.CheckErr(t, err)
	wanted = []string{"s5-0"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	ss, err = c.FilterServers(false, []string{"s5[1]"}, []string{}, "", false)
	test.CheckErr(t, err)
	wanted = []string{"s5-1"}
	test.CheckEqN(t, len(ss), len(wanted))
	for i, s := range ss {
		test.CheckEqS(t, s.Name, wanted[i])
	}

	// Want errors when input malformed
	_, err = c.FilterServers(false, []string{"s4[1"}, []string{}, "", false)
	test.WantErr(t, err)
	_, err = c.FilterServers(false, []string{"s4[1]]"}, []string{}, "", false)
	test.WantErr(t, err)
	_, err = c.FilterServers(false, []string{"s4[1::]"}, []string{}, "", false)
	test.WantErr(t, err)
	_, err = c.FilterServers(false, []string{"s4[01:a]"}, []string{}, "", false)
	test.WantErr(t, err)
	_, err = c.FilterServers(false, []string{"s4[-1]"}, []string{}, "", false)
	test.WantErr(t, err)
}
