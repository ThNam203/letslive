package dto

import "testing"

func TestFormatAmount(t *testing.T) {
	cases := []struct {
		amount    int64
		precision int
		want      string
	}{
		{1500, 2, "15.00"},
		{50, 2, "0.50"},
		{1, 2, "0.01"},
		{0, 2, "0.00"},
		{-100, 2, "-1.00"},
		{-1, 2, "-0.01"},
		{7, 0, "7"},
		{0, 0, "0"},
		{1234567, 2, "12345.67"},
		{1, 4, "0.0001"},
	}
	for _, c := range cases {
		got := FormatAmount(c.amount, c.precision)
		if got != c.want {
			t.Errorf("FormatAmount(%d, %d) = %q, want %q", c.amount, c.precision, got, c.want)
		}
	}
}

func TestParseAmount(t *testing.T) {
	cases := []struct {
		input     string
		precision int
		want      int64
	}{
		{"50.00", 2, 5000},
		{"50", 2, 5000},
		{"0.5", 2, 50},
		{"0.50", 2, 50},
		{"1", 2, 100},
		{"1.01", 2, 101},
		{"12345.67", 2, 1234567},
		{"7", 0, 7},
		{"0.0001", 4, 1},
	}
	for _, c := range cases {
		got, err := ParseAmount(c.input, c.precision)
		if err != nil {
			t.Errorf("ParseAmount(%q, %d) error: %v", c.input, c.precision, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseAmount(%q, %d) = %d, want %d", c.input, c.precision, got, c.want)
		}
	}
}

func TestParseAmountRejects(t *testing.T) {
	cases := []struct {
		input     string
		precision int
		reason    string
	}{
		{"-1", 2, "negative sign"},
		{"+1", 2, "positive sign"},
		{"", 2, "empty"},
		{"abc", 2, "non-numeric whole"},
		{"1.ab", 2, "non-numeric fraction"},
		{"1.234", 2, "exceeds precision"},
		{"1e3", 2, "scientific notation"},
	}
	for _, c := range cases {
		_, err := ParseAmount(c.input, c.precision)
		if err == nil {
			t.Errorf("ParseAmount(%q, %d) accepted but should fail (%s)", c.input, c.precision, c.reason)
		}
	}
}

func TestFormatParseRoundTrip(t *testing.T) {
	values := []int64{0, 1, 100, 99, 12345, -1, -100, -12345}
	for _, v := range values {
		s := FormatAmount(v, 2)
		// FormatAmount may emit signed; ParseAmount rejects negatives. Round-trip on unsigned only.
		if v < 0 {
			continue
		}
		back, err := ParseAmount(s, 2)
		if err != nil {
			t.Errorf("round-trip ParseAmount(%q) error: %v", s, err)
			continue
		}
		if back != v {
			t.Errorf("round-trip mismatch: original=%d formatted=%q parsed=%d", v, s, back)
		}
	}
}
