package utils

import "fmt"

// StringSlice is a custom flag type for collecting multiple values.
// It can be used with flag.Var() to allow specifying a flag multiple times.
type StringSlice []string

func (s *StringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *StringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// DeduplicateStrings removes duplicate strings from a slice.
func DeduplicateStrings(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	result := make([]string, 0, len(input))

	for _, item := range input {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}
