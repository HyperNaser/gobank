package util

const (
	USD = "USD"
	EUR = "EUR"
	BHD = "BHD"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, BHD:
		return true
	}

	return false
}
