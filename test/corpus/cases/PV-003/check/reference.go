package main

import "time"

type refPlace struct {
	node string
	slot int
}

// reference holds the cooldown under the tenant identity: a cooldown belongs
// to the tenant whose quota ran out. Used only by this check.
type reference struct {
	now      func() time.Time
	cooldown time.Duration
	slots    map[string]refPlace
	until    map[string]time.Time
}

func newReference(now func() time.Time, cooldown time.Duration) *reference {
	return &reference{now: now, cooldown: cooldown, slots: map[string]refPlace{}, until: map[string]time.Time{}}
}

func (r *reference) Assign(tenant, node string, slot int) {
	r.slots[tenant] = refPlace{node, slot}
}

func (r *reference) Record(tenant, node string, slot int, exhausted bool) {
	if exhausted {
		r.until[tenant] = r.now().Add(r.cooldown)
		return
	}
	delete(r.until, tenant)
}

// Allow looks up the tenant now assigned to (node, slot) and applies that
// tenant's own cooldown.
func (r *reference) Allow(node string, slot int) (bool, string) {
	for tenant, p := range r.slots {
		if p.node == node && p.slot == slot {
			if u, ok := r.until[tenant]; ok && r.now().Before(u) {
				return false, tenant
			}
		}
	}
	return true, ""
}
