package billing

// Discounted returns the price in cents after a percent discount.
// The discount is rounded down to a whole cent.
func Discounted(cents int, percent int) int {
	off := cents * percent / 100
	return cents - off
}
