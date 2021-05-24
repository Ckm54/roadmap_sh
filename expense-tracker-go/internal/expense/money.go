package expense

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Cents is an amount of money in the smallest currency unit (1/100), so
// 12.50 is stored as 1250. Integers add up exactly; float64 cannot represent
// most decimals (0.1 + 0.2 != 0.3), and those errors pile up in totals.
type Cents int64

// ParseCents turns user input such as "12", "12.5" or "12.50" into Cents.
// It parses the digits itself instead of going through float64, so no
// rounding happens. More than two decimal places is an error rather than
// being silently rounded.
func ParseCents(s string) (Cents, error) {
	s = strings.TrimSpace(s)
	invalid := fmt.Errorf("invalid amount %q: use a positive number like 12 or 12.50", s)

	whole, frac, hasDot := strings.Cut(s, ".")
	if whole == "" && frac == "" {
		return 0, invalid
	}
	if !isDigits(whole) || (hasDot && !isDigits(frac)) {
		return 0, invalid
	}
	if len(frac) > 2 {
		return 0, fmt.Errorf("invalid amount %q: at most 2 decimal places", s)
	}

	w, err := parseUint(whole)
	if err != nil {
		return 0, errTooLarge
	}
	f, _ := parseUint(frac + strings.Repeat("0", 2-len(frac))) // "5" -> "50"

	if w > (math.MaxInt64-f)/100 {
		return 0, errTooLarge
	}
	return Cents(w*100 + f), nil
}

var errTooLarge = errors.New("amount is too large")

// String formats c the way users type it, e.g. 1250 -> "12.50".
func (c Cents) String() string {
	sign := ""
	if c < 0 {
		sign, c = "-", -c
	}
	return fmt.Sprintf("%s%d.%02d", sign, c/100, c%100)
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// parseUint treats an empty string as 0, so ".5" means 0.50.
func parseUint(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseInt(s, 10, 64)
}
