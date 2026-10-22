package billing

import "testing"

func TestDiscounted(t *testing.T) {
	got := Discounted(100, 20)
	if got != 80 {
		t.Fatalf("got %d want 80", got)
	}
}
