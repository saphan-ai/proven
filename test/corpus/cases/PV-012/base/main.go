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
	store *store
	team  string
}

func fromSettings(store *store, _ settings) page { return page{store: store} }
func forTeam(store *store, team string) page     { return page{store: store, team: team} }

type request struct{ team string }

func (page page) list(_ request) ([]string, string) {
	ids := make([]string, 0, len(page.store.events))
	for _, entry := range page.store.events {
		ids = append(ids, entry.item)
	}
	return ids, "Activity for all teams"
}
func run(input string) ([]string, string, error) {
	store := newStore()
	store.sync("ours")
	store.addForeign()
	store.sync("theirs")
	var page page
	var request request
	switch input {
	case "control", "extra-sync", "control-no-registry", "control-wrong-ownership":
		page = forTeam(store, "ours")
		if input == "extra-sync" || input == "control-wrong-ownership" {
			store.sync("ours")
		}
	case "vacuity":
		page = forTeam(store, "theirs")
	case "application", "scope-no-team", "scope-with-team":
		page = fromSettings(store, settings{})
		if input == "scope-with-team" {
			request.team = "ours"
		}
	default:
		return nil, "", fmt.Errorf("unknown input %q", input)
	}
	ids, scope := page.list(request)
	return ids, scope, nil
}
func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: base <input>")
		os.Exit(2)
	}
	ids, scope, err := run(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Printf("%s|%s|not-present", strings.Join(ids, ","), scope)
}
