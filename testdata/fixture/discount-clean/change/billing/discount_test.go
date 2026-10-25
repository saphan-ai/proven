package billing

import "testing"

func TestDiscounted(t *testing.T) {
	got := Discounted(100, 20)
	if got != 80 {
		t.Fatalf("got %d want 80", got)
	}
}

func TestDiscountedRoundsTheDiscountDown(t *testing.T) {
	got := Discounted(99, 10)
	if got != 90 {
		t.Fatalf("got %d want 90", got)
	}
}
