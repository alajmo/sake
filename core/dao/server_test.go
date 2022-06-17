package dao

import (
	"testing"
)

func TestFilterServers(t *testing.T) {
	s1 := Server{Name: "s1", Tags: []string{"t1", "t2"}}
	s2 := Server{Name: "s2", Tags: []string{"t3"}}
	s3 := Server{Name: "s3", Tags: []string{"t2", "t3"}}
	s4 := Server{Name: "s4", Tags: []string{"t8"}}
	s5 := Server{Name: "s5", Tags: []string{"t1", "t2"}}

	c := Config{
		Servers: []Server{s1, s2, s3, s4, s5},
	}

	ss, err := c.FilterServers(false, []string{"s1", "s2"}, []string{"t1"})
	if err != nil {
		t.Fatalf("%q", err)
	}
	wanted := []string{"s1", "s2", "s5"}
	for i, s := range ss {
		if s.Name != wanted[i] {
			t.Fatalf(`Wanted: %q, Found: %q`, wanted[i], s.Name)
		}
	}

	ss, err = c.FilterServers(false, []string{"s5"}, []string{})
	if err != nil {
		t.Fatalf("%q", err)
	}
	wanted = []string{"s5"}
	for i, s := range ss {
		if s.Name != wanted[i] {
			t.Fatalf(`Wanted: %q, Found: %q`, wanted[i], s.Name)
		}
	}

	ss, err = c.FilterServers(false, []string{}, []string{"t1"})
	if err != nil {
		t.Fatalf("%q", err)
	}
	wanted = []string{"s1", "s5"}
	for i, s := range ss {
		if s.Name != wanted[i] {
			t.Fatalf(`Wanted: %q, Found: %q`, wanted[i], s.Name)
		}
	}

	ss, err = c.FilterServers(true, []string{}, []string{})
	if err != nil {
		t.Fatalf("%q", err)
	}
	wanted = []string{"s1", "s2", "s3", "s4", "s5"}
	for i, s := range ss {
		if s.Name != wanted[i] {
			t.Fatalf(`Wanted: %q, Found: %q`, wanted[i], s.Name)
		}
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
	if err != nil {
		t.Fatalf("%q", err)
	}
	wanted := []string{"s1", "s5"}
	for i, s := range ss {
		if s.Name != wanted[i] {
			t.Fatalf(`Wanted: %q, Found: %q`, wanted[i], s.Name)
		}
	}

	ss, err = c.GetServersByTags([]string{"t1", "t2"})
	if err != nil {
		t.Fatalf("%q", err)
	}
	wanted = []string{"s1"}
	if len(ss) != len(wanted) {
		t.Fatalf("Wanted: %d servers, Found: %d servers", len(wanted), len(ss))
	}
	for i, s := range ss {
		if s.Name != wanted[i] {
			t.Fatalf(`Wanted: %q, Found: %q`, wanted[i], s.Name)
		}
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
	if err != nil {
		t.Fatalf("%q", err)
	}

	wanted := Tag{
		Name:    "t1",
		Servers: []string{"s1", "s5"},
	}
	if !Equal(ss[0].Servers, wanted.Servers) {
		t.Fatalf("Wanted: %q, Found: %q", wanted.Servers, ss[0].Servers)
	}
}

// Equal tells whether a and b contain the same elements.
// A nil argument is equivalent to an empty slice.
func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
