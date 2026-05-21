package dto

import (
	"errors"
	"fmt"
	"strings"
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

// ParseAmount converts a positive decimal string into integer minor units for
// the given currency precision. Rejects negatives, scientific notation, and
// fractional digits beyond the currency precision.
//   ParseAmount("50.00", 2) -> 5000
//   ParseAmount("0.5",   2) -> 50
//   ParseAmount("1",     2) -> 100
func ParseAmount(s string, precision int) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("amount is empty")
	}
	if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+") {
		return 0, errors.New("amount must be unsigned")
	}

	parts := strings.SplitN(s, ".", 2)
	wholeStr := parts[0]
	fracStr := ""
	if len(parts) == 2 {
		fracStr = parts[1]
	}

	if wholeStr == "" && fracStr == "" {
		return 0, errors.New("amount has no digits")
	}
	if len(fracStr) > precision {
		return 0, fmt.Errorf("amount has more than %d decimal places", precision)
	}
	for _, c := range wholeStr {
		if c < '0' || c > '9' {
			return 0, errors.New("amount has non-numeric characters")
		}
	}
	for _, c := range fracStr {
		if c < '0' || c > '9' {
			return 0, errors.New("amount has non-numeric characters")
		}
	}

	// pad fractional part to full precision
	for len(fracStr) < precision {
		fracStr += "0"
	}

	combined := wholeStr + fracStr
	if combined == "" {
		return 0, errors.New("amount is empty")
	}

	var n int64
	for _, c := range combined {
		d := int64(c - '0')
		// overflow guard
		if n > (1<<62)/10 {
			return 0, errors.New("amount too large")
		}
		n = n*10 + d
	}
	return n, nil
}
