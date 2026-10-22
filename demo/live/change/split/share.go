package split

// Share divides cents across n people.
// each is one share. left is the cents that do not divide evenly.
func Share(cents, n int) (each, left int) {
	if n <= 0 {
		return 0, cents
	}
	return cents / n, 0
}
