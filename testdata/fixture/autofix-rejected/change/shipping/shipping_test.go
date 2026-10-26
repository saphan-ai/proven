package shipping

import "testing"

func TestFee(t *testing.T) {
	got := Fee(1000)
	if got != 500 {
		t.Fatalf("got %d want 500", got)
	}
}

func TestFreeShipping(t *testing.T) {
	got := Fee(6000)
	if got != 0 {
		t.Fatalf("got %d want 0", got)
	}
}
