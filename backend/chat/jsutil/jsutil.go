// Package jsutil reproduces the few JavaScript string and number semantics the
// Node chat service relied on, so the Go port accepts and rejects exactly the
// same inputs.
package jsutil

import (
	"strings"
	"unicode"
	"unicode/utf16"
)

// Length is JavaScript's String.prototype.length: UTF-16 code units, not bytes
// or runes. Length limits are compared against this value.
func Length(s string) int {
	return len(utf16.Encode([]rune(s)))
}

// Trim is String.prototype.trim, which also strips the byte order mark that
// unicode.IsSpace does not treat as space.
func Trim(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || r == 0xFEFF
	})
}

// Truncate is String.prototype.substring(0, n) for n UTF-16 code units.
func Truncate(s string, n int) string {
	units := utf16.Encode([]rune(s))
	if len(units) <= n {
		return s
	}
	return string(utf16.Decode(units[:n]))
}

// ParseInt mirrors parseInt(s, 10): leading whitespace and sign are allowed and
// parsing stops at the first non-digit. ok is false where parseInt gives NaN.
func ParseInt(s string) (value int, ok bool) {
	s = Trim(s)
	negative := false
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		negative = s[0] == '-'
		s = s[1:]
	}

	digits := 0
	for digits < len(s) && s[digits] >= '0' && s[digits] <= '9' {
		// clamp instead of overflowing; callers clamp to small ranges anyway
		if value < 1<<31 {
			value = value*10 + int(s[digits]-'0')
		}
		digits++
	}
	if digits == 0 {
		return 0, false
	}
	if negative {
		value = -value
	}
	return value, true
}

// IntOr is the `parseInt(s) || fallback` idiom: NaN and 0 both fall back.
func IntOr(s string, fallback int) int {
	if value, ok := ParseInt(s); ok && value != 0 {
		return value
	}
	return fallback
}
