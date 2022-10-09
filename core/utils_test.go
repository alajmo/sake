package core

import (
	"testing"

	"github.com/alajmo/sake/core/test"
)

func checkEqHost(t *testing.T, hostname string, defaultUser string, defaultPort uint16, wantedHost string, wantedUser string, wantedPort uint16) {
	foundUser, foundHost, foundPort, err := ParseHostName(hostname, defaultUser, defaultPort)
	test.CheckErr(t, err)

	if foundHost != wantedHost {
		t.Fatalf(`Wanted: %q, Found: %q`, wantedHost, foundHost)
	}
	if foundUser != wantedUser {
		t.Fatalf(`Wanted: %q, Found: %q`, wantedUser, foundUser)
	}
	if foundPort != wantedPort {
		t.Fatalf(`Wanted: %q, Found: %q`, wantedPort, foundPort)
	}
}

func TestParseHostName(t *testing.T) {
	// IPv4
	checkEqHost(t, "192.168.0.1", "test", 33, "192.168.0.1", "test", 33)
	checkEqHost(t, "user@192.168.0.1", "test", 33, "192.168.0.1", "user", 33)
	checkEqHost(t, "192.168.0.1:44", "test", 33, "192.168.0.1", "test", 44)
	checkEqHost(t, "user@192.168.0.1:44", "test", 33, "192.168.0.1", "user", 44)

	// IPv6
	checkEqHost(t, "2001:3984:3989::10", "test", 33, "2001:3984:3989::10", "test", 33)
	checkEqHost(t, "user@2001:3984:3989::10", "test", 33, "2001:3984:3989::10", "user", 33)
	checkEqHost(t, "[2001:3984:3989::10]:44", "test", 33, "2001:3984:3989::10", "test", 44)
	checkEqHost(t, "user@[2001:3984:3989::10]:44", "test", 33, "2001:3984:3989::10", "user", 44)

	// Resolved hostname
	checkEqHost(t, "resolved", "test", 33, "resolved", "test", 33)
	checkEqHost(t, "resolved.com", "test", 33, "resolved.com", "test", 33)
	checkEqHost(t, "user@resolved", "test", 33, "resolved", "user", 33)
}
