// Command check runs every input of the PV-003 contract against base, change
// and the reference gate, and prints one line per input.
package main

import (
	"fmt"
	"time"

	"github.com/saphan-ai/proven/test/corpus/cases/PV-003/base"
	"github.com/saphan-ai/proven/test/corpus/cases/PV-003/change"
)

type gate interface {
	Assign(tenant, node string, slot int)
	Record(tenant, node string, slot int, exhausted bool)
	Allow(node string, slot int) (bool, string)
}

type input struct {
	name string
	run  func(g gate) (bool, string)
}

var inputs = []input{
	{"own-case-cooled-slot", func(g gate) (bool, string) {
		// A on slot 1 uses up its quota; B on slot 2 then runs without using up its own.
		g.Assign("A", "n1", 1)
		g.Assign("B", "n1", 2)
		g.Record("A", "n1", 1, true)
		g.Record("B", "n1", 2, false)
		return g.Allow("n1", 1)
	}},
	{"own-case-other-slot", func(g gate) (bool, string) {
		g.Assign("A", "n1", 1)
		g.Assign("B", "n1", 2)
		g.Record("A", "n1", 1, true)
		g.Record("B", "n1", 2, false)
		return g.Allow("n1", 2)
	}},
	{"move-moved-tenant", func(g gate) (bool, string) {
		g.Assign("X", "n1", 1)
		g.Record("X", "n1", 1, true)
		g.Assign("X", "n1", 2)
		return g.Allow("n1", 2)
	}},
	{"vacated-slot", func(g gate) (bool, string) {
		g.Assign("X", "n1", 1)
		g.Record("X", "n1", 1, true)
		g.Assign("X", "n1", 2)
		g.Assign("Y", "n1", 1)
		return g.Allow("n1", 1)
	}},
	{"control-no-cooldown", func(g gate) (bool, string) {
		g.Assign("X", "n1", 1)
		g.Assign("Y", "n1", 2)
		g.Record("X", "n1", 1, false)
		return g.Allow("n1", 1)
	}},
}

func show(ok bool, tenant string) string {
	if ok {
		return "allowed"
	}
	return "refused(" + tenant + ")"
}

func main() {
	const cooldown = time.Minute
	for _, in := range inputs {
		clock := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC) // fake clock, never advanced
		now := func() time.Time { return clock }
		b := show(in.run(base.New(now, cooldown)))
		c := show(in.run(change.New(now, cooldown)))
		r := show(in.run(newReference(now, cooldown)))
		fmt.Printf("%s | base=%s | change=%s | reference=%s\n", in.name, b, c, r)
	}
}
