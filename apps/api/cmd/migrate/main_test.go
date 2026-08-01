package main

import "testing"

func TestMigrateRejectsDestructiveOrImplicitCommands(t *testing.T) {
	for _, arguments := range [][]string{nil, {}, {"down"}, {"up", "down"}} {
		if err := run(arguments); err == nil {
			t.Fatalf("arguments %#v must be rejected", arguments)
		}
	}
}
