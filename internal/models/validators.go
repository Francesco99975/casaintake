package models

import (
	"log/slog"
	"regexp"
	"strings"
	// "github.com/disposable/disposable"
)

// func IsDisposableEmail(email string) bool {
// 	email = strings.ToLower(strings.TrimSpace(email))

// 	parts := strings.Split(email, "@")
// 	if len(parts) != 2 {
// 		return true // invalid format → treat as bad
// 	}

// 	domain := parts[1]

// 	return disposable.Domain(domain)
// }

// Optional: also reject common patterns
func IsSuspiciousEmail(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))

	// if IsDisposableEmail(email) {
	// 	return true
	// }

	// Extra rules you can add:
	if strings.Count(email, ".") > 5 { // too many dots
		return true
	}
	if strings.Contains(email, "+") && len(email) > 40 { // long plus-addressing
		return true
	}

	return false
}

// separatorRegex matches whitespace and hyphens used to visually group
// OHIP number digits (e.g. "1234-567-890" or "1234 567 890").
var separatorRegex = regexp.MustCompile(`[\s-]+`)

// ohipPattern matches exactly 10 digits followed by exactly 2 letters,
// once separators have been stripped out.
var ohipPattern = regexp.MustCompile(`^(\d{10})([A-Za-z]{2})$`)

// PassesLuhnChecksum reports whether the given string of digits satisfies
// the Luhn (mod 10) checksum. Every second digit, counting from the
// rightmost, is doubled; if that doubling exceeds 9, 9 is subtracted.
func PassesLuhnChecksum(digits string) bool {
	if len(digits) == 0 {
		slog.Error("invalid Luhn checksum", "digits", digits)
		return false
	}
	sum := 0
	doubleDigit := false
	for i := len(digits) - 1; i >= 0; i-- {
		c := digits[i]
		if c < '0' || c > '9' {
			slog.Error("invalid Luhn checksum", "digits", digits)
			return false
		}
		d := int(c - '0')
		if doubleDigit {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		doubleDigit = !doubleDigit
	}
	return sum%10 == 0
}

// ParseOHIPNumber normalizes an OHIP input (stripping spaces/hyphens,
// uppercasing) and splits it into the 10-digit health number and the
// required 2-letter version code.
//
// Accepted formats:
//
//	1234-567-890-AB
//	1234567890AB
//	1234 567 890 AB
//
// ok is false if the input doesn't structurally match one of these formats.
func ParseOHIPNumber(input string) (healthNumber string, versionCode string, ok bool) {
	normalized := strings.ToUpper(separatorRegex.ReplaceAllString(strings.TrimSpace(input), ""))
	m := ohipPattern.FindStringSubmatch(normalized)
	if m == nil {
		slog.Error("invalid OHIP number", "input", input)
		return "", "", false
	}
	return m[1], m[2], true
}

// IsValidOHIPNumber reports whether input is a valid Ontario Health Card
// number: 10 digits (passing the Luhn checksum) followed by exactly a
// 2-letter version code, in any of the accepted separator formats.
func IsValidOHIP(input string) bool {
	healthNumber, versionCode, ok := ParseOHIPNumber(input)
	if !ok {
		slog.Error("invalid OHIP number", "input", input)
		return false
	}
	if len(versionCode) != 2 {
		slog.Error("invalid OHIP version code", "versionCode", versionCode)
		return false
	}
	return PassesLuhnChecksum(healthNumber)
}
