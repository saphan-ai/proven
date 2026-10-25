// Package base is the cooldown gate before the change: one cooldown per node.
package base

import "time"

type entry struct {
	tenant string
	until  time.Time
}

type place struct {
	node string
	slot int
}

// Gate holds the cooldown under one key per node.
type Gate struct {
	now      func() time.Time
	cooldown time.Duration
	slots    map[string]place // tenant -> its latest assignment
	cool     map[string]entry // node -> cooldown
}

func New(now func() time.Time, cooldown time.Duration) *Gate {
	return &Gate{now: now, cooldown: cooldown, slots: map[string]place{}, cool: map[string]entry{}}
}

// Assign places a tenant on a slot; a later assignment replaces the earlier one.
func (g *Gate) Assign(tenant, node string, slot int) {
	g.slots[tenant] = place{node, slot}
}

// Record notes a run. A run that used up the quota starts the node's cooldown;
// any other run on the node clears it.
func (g *Gate) Record(tenant, node string, slot int, exhausted bool) {
	if exhausted {
		g.cool[node] = entry{tenant, g.now().Add(g.cooldown)}
		return
	}
	delete(g.cool, node)
}

// Allow says whether a session may open on (node, slot): allowed, or the
// tenant whose cooldown applies.
func (g *Gate) Allow(node string, slot int) (bool, string) {
	e, ok := g.cool[node]
	if !ok || !g.now().Before(e.until) {
		return true, ""
	}
	return false, e.tenant
}
