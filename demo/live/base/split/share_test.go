package split

import "testing"

func TestShare(t *testing.T) {
	each, left := Share(100, 3)
	if each != 33 || left != 1 {
		t.Fatalf("got %d left %d, want 33 left 1", each, left)
	}
}
