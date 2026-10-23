package sum

import "testing"

func TestTotal(t *testing.T) {
	if Total(nil) != 0 {
		t.Fatal(Total(nil))
	}
}
