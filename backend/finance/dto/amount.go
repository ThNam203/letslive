package dto

import (
	"fmt"
)

// FormatAmount converts an integer minor-unit amount (e.g. cents) into a decimal
// string with the given currency precision. Negative values keep their sign.
//   FormatAmount(1500, 2)  -> "15.00"
//   FormatAmount(-100, 2)  -> "-1.00"
//   FormatAmount(7, 0)     -> "7"
func FormatAmount(amount int64, precision int) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}

	div := int64(1)
	for i := 0; i < precision; i++ {
		div *= 10
	}

	whole := amount / div
	frac := amount % div

	sign := ""
	if negative {
		sign = "-"
	}

	if precision == 0 {
		return fmt.Sprintf("%s%d", sign, whole)
	}
	return fmt.Sprintf("%s%d.%0*d", sign, whole, precision, frac)
}
