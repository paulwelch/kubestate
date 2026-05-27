package main

import "testing"

func TestNewAppRegistersCoreCommands(t *testing.T) {
	app := newApp()

	want := map[string]bool{
		"get":   false,
		"list":  false,
		"top":   false,
		"watch": false,
	}
	for _, c := range app.Commands {
		if _, ok := want[c.Name]; ok {
			want[c.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("expected command %q to be registered", name)
		}
	}
}

func TestTopSubcommandsHaveUniqueNamesAndAliases(t *testing.T) {
	app := newApp()

	var topFound bool
	for _, c := range app.Commands {
		if c.Name != "top" {
			continue
		}
		topFound = true
		seen := map[string]string{}
		for _, sc := range c.Subcommands {
			if owner, ok := seen[sc.Name]; ok {
				t.Fatalf("duplicate top subcommand name %q between %q and %q", sc.Name, owner, sc.Name)
			}
			seen[sc.Name] = sc.Name
			for _, alias := range sc.Aliases {
				if owner, ok := seen[alias]; ok {
					t.Fatalf("duplicate top alias/name %q between %q and %q", alias, owner, sc.Name)
				}
				seen[alias] = sc.Name
			}
		}
	}

	if !topFound {
		t.Fatal("expected top command to be registered")
	}
}
