// Package change is the cooldown gate under review: one cooldown per (node, slot).
package change

import "time"

type entry struct {
	tenant string
	until  time.Time
}

type place struct {
	node string
	slot int
}

// Gate holds the cooldown under (node, slot), one table per slot. The tenant
// on a run record is kept only as the name in a refusal.
type Gate struct {
	now      func() time.Time
	cooldown time.Duration
	slots    map[string]place
	cool     map[place]entry
}

func New(now func() time.Time, cooldown time.Duration) *Gate {
	return &Gate{now: now, cooldown: cooldown, slots: map[string]place{}, cool: map[place]entry{}}
}

// Assign places a tenant on a slot; a later assignment replaces the earlier one.
func (g *Gate) Assign(tenant, node string, slot int) {
	g.slots[tenant] = place{node, slot}
}

// Record notes a run. A run that used up the quota starts the slot's cooldown;
// any other run on the slot clears it. The tenant identity is not consulted.
func (g *Gate) Record(tenant, node string, slot int, exhausted bool) {
	k := place{node, slot}
	if exhausted {
		g.cool[k] = entry{tenant, g.now().Add(g.cooldown)}
		return
	}
	delete(g.cool, k)
}

// Allow says whether a session may open on (node, slot): allowed, or the
// tenant whose cooldown applies.
func (g *Gate) Allow(node string, slot int) (bool, string) {
	e, ok := g.cool[place{node, slot}]
	if !ok || !g.now().Before(e.until) {
		return true, ""
	}
	return false, e.tenant
}
