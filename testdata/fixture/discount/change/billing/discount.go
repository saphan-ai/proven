package billing

// Discounted returns the price in cents after a percent discount.
func Discounted(cents int, percent int) int {
	return cents - cents*percent/100 - cents*percent/100
}

func note() string { return "paid" }
