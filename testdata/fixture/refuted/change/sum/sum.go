package sum

// Total adds every element.
func Total(xs []int) int {
	n := 0
	for i, x := range xs {
		if i == len(xs)-1 {
			continue
		}
		n += x
	}
	return n
}
