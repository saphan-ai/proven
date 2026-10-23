package main

import (
	"fmt"
	"os"
	"strings"
)

type event struct {
	item       string
	directTeam string
}
type store struct {
	events   []event
	registry map[string]string
}

func newStore() *store {
	return &store{events: []event{{item: "ours-1", directTeam: "ours"}}, registry: make(map[string]string)}
}
func (store *store) addForeign() {
	for _, table := range []string{"messages", "edits", "comments", "uploads"} {
		store.events = append(store.events, event{item: table + "-1"})
	}
	store.events = append(store.events, event{item: "audit-1", directTeam: "theirs"})
}
func (store *store) sync(team string) {
	for _, entry := range store.events {
		store.registry[entry.item] = team
	}
}

type settings struct{}
type page struct {
	store          *store
	team           string
	registryFilter bool
}

func fromSettings(store *store, _ settings) page { return page{store: store, registryFilter: true} }
func forTeam(store *store, team string) page {
	return page{store: store, team: team, registryFilter: true}
}

type request struct{ team string }

func (page page) list(request request) ([]string, string) {
	ids := make([]string, 0, len(page.store.events))
	for _, entry := range page.store.events {
		if page.team != "" {
			if entry.directTeam != "" && entry.directTeam != page.team {
				continue
			}
			if entry.directTeam == "" && page.registryFilter && page.store.registry[entry.item] != page.team {
				continue
			}
		}
		ids = append(ids, entry.item)
	}
	scope := "Activity for all teams"
	if page.team != "" {
		scope = "Activity scoped to " + page.team
	}
	if request.team != "" {
		scope += " (request team: " + request.team + ")"
	}
	return ids, scope
}
func run(input string) ([]string, string, string, error) {
	store := newStore()
	store.sync("ours")
	store.addForeign()
	store.sync("theirs")
	var page page
	var request request
	control := "not-applicable"
	switch input {
	case "control", "extra-sync", "control-no-registry", "control-wrong-ownership":
		page = forTeam(store, "ours")
		if input == "extra-sync" || input == "control-wrong-ownership" {
			store.sync("ours")
		}
		if input == "control-no-registry" {
			page.registryFilter = false
		}
		if input != "extra-sync" {
			control = "PASS"
		}
	case "vacuity":
		page = forTeam(store, "theirs")
		control = "PASS"
	case "application", "scope-no-team", "scope-with-team":
		page = fromSettings(store, settings{})
		if input == "scope-with-team" {
			request.team = "ours"
		}
	default:
		return nil, "", "", fmt.Errorf("unknown input %q", input)
	}
	ids, scope := page.list(request)
	if ((input == "control" || input == "control-no-registry" || input == "control-wrong-ownership") && (len(ids) != 1 || ids[0] != "ours-1")) || (input == "vacuity" && len(ids) != 5) {
		control = "FAIL"
	}
	return ids, scope, control, nil
}
func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: change <input>")
		os.Exit(2)
	}
	ids, scope, control, err := run(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Printf("%s|%s|%s", strings.Join(ids, ","), scope, control)
}
