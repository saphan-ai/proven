package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/store"
)

// runData imports administrator-provided source tables, never model-written SQL.
// Each SQL table replacement and its snapshot identity commit atomically.
func runData(args []string, env func(string) string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprintln(stdout, "usage: proven data import <dataset.json>")
		return exitOK
	}
	if len(args) != 2 || args[0] != "import" {
		fmt.Fprintln(stderr, "usage: proven data import <dataset.json>")
		return exitFatal
	}
	f, err := os.Open(args[1])
	if err != nil {
		fmt.Fprintln(stderr, "data import: cannot open dataset")
		return exitFatal
	}
	defer f.Close()
	var dataset struct {
		Version int
		Tables  []store.Table
	}
	dec := json.NewDecoder(io.LimitReader(f, (256<<20)+1))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&dataset); err != nil || dataset.Version != 1 || len(dataset.Tables) == 0 {
		fmt.Fprintln(stderr, "data import: expected version 1 and a non-empty tables array")
		return exitConfig
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		fmt.Fprintln(stderr, "data import: trailing or oversized input")
		return exitConfig
	}
	choice, err := store.Load(env)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitConfig
	}
	var db *store.SQLStore
	switch choice.Name {
	case store.NameSQLite:
		db, err = store.OpenSQLite(choice.File)
	case store.NamePostgres:
		policy, perr := network.Load(env)
		if perr != nil {
			fmt.Fprintln(stderr, perr)
			return exitConfig
		}
		db, err = store.OpenPostgres(env(store.EnvPostgresURL), policy)
	default:
		fmt.Fprintln(stderr, "data import: choose sqlite or postgres; other adapters retain their native import tools")
		return exitConfig
	}
	if err != nil {
		fmt.Fprintln(stderr, "data import: could not open database")
		return exitFatal
	}
	defer db.Close()
	for _, table := range dataset.Tables {
		if err := db.Replace(context.Background(), table); err != nil {
			fmt.Fprintln(stderr, "data import: "+err.Error())
			return exitFatal
		}
		fmt.Fprintln(stdout, "imported "+table.Name)
	}
	return exitOK
}
