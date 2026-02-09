// Package term provides ANSI terminal color utilities.
//
// Colors are automatically disabled when stdout is not a terminal
// or when the NO_COLOR environment variable is set.
package term

import "os"

//nolint:gochecknoglobals // package-level color state is idiomatic for terminal libraries
var colorEnabled = detectColor()

func detectColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return fi.Mode()&os.ModeCharDevice != 0
}

func apply(code, s string) string {
	if !colorEnabled {
		return s
	}

	return "\033[" + code + "m" + s + "\033[0m"
}

// Bold returns s with bold formatting.
func Bold(s string) string { return apply("1", s) }

// Dim returns s with dim formatting.
func Dim(s string) string { return apply("2", s) }

// Green returns s colored green.
func Green(s string) string { return apply("32", s) }

// Yellow returns s colored yellow.
func Yellow(s string) string { return apply("33", s) }

// Cyan returns s colored cyan.
func Cyan(s string) string { return apply("36", s) }

// BoldRed returns s with bold red formatting.
func BoldRed(s string) string { return apply("1;31", s) }

// BoldYellow returns s with bold yellow formatting.
func BoldYellow(s string) string { return apply("1;33", s) }

// BoldCyan returns s with bold cyan formatting.
func BoldCyan(s string) string { return apply("1;36", s) }
