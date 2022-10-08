package test

import (
	"testing"
)


func IsError(t *testing.T, err error) {
	if err == nil {
		t.Fatalf("%q", err)
	}
}

func CheckErr(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("%q", err)
	}
}

func WantErr(t *testing.T, err error) {
	if err == nil {
		t.Fatalf("Wanted error, got nil")
	}
}

func CheckEqS(t *testing.T, found string, wanted string) {
	if found != wanted {
		t.Fatalf(`Wanted: %q, Found: %q`, wanted, found)
	}
}

func CheckEqN(t *testing.T, found int, wanted int) {
	if found != wanted {
		t.Fatalf(`Wanted: %d, Found: %d`, wanted, found)
	}
}

// Equal tells whether a and b contain the same elements.
// A nil argument is equivalent to an empty slice.
func CheckEqualStringArr(t *testing.T, found []string, wanted []string) {
	equal := true

	if len(found) != len(wanted) {
		equal = false
	}
	for i, v := range found {
		if v != wanted[i] {
			equal = false
		}
	}

	if !equal {
		t.Fatalf("Wanted: %q, Found: %q", wanted, found[0])
	}
}
