package utils

import (
	"fmt"
)

// Bytes convert byte size to human-readable size.
func Bytes[T int64 | int](b T) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

// ToStrings converts a slice of any custom string-like type (T ~string)
// into a regular []string so it can be used with functions like strings.Join.
func ToStrings[T ~string](in []T) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = string(v) // cast each element to string
	}
	return out
}
