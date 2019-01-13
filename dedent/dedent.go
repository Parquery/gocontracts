package dedent

import (
	"math"
)

// TrimEmptyLines removes the prefix and suffix empty lines.
func TrimEmptyLines(lines []string) (result []string) {
	start := 0
	for i, line := range lines {
		if len(line) == 0 {
			start = i + 1
		} else {
			break
		}
	}

	end := len(lines)
	for i := len(lines) - 1; i >= 0; i-- {
		if len(lines[i]) == 0 {
			end--
		} else {
			break
		}
	}

	if start == len(lines) {
		return []string{}
	}

	return lines[start:end]
}

// allLinesEmpty returns true if all given lines are empty.
func allLinesEmpty(lines []string) bool {
	result := true
	for _, line := range lines {
		if len(line) > 0 {
			result = false
			break
		}
	}

	return result
}

// maxPrefixLength the maximum length of the possible common whitespace prefix.
func maxPrefixLength(lines []string) int {
	result := math.MaxInt64
	for _, line := range lines {
		// Empty lines are ignored in dedention.
		if len(line) == 0 {
			continue
		}

		if len(line) < result {
			result = len(line)
		}
	}

	return result
}

// commonWhitespacePrefix determines the end of the common whitespace prefix
// between the lines.
func commonWhitespacePrefix(lines []string) int {
	////
	// Edge cases
	////

	if len(lines) == 0 || allLinesEmpty(lines) {
		return 0
	}

	////
	// Determine the common whitespace prefix
	////

	maxPreLen := maxPrefixLength(lines)

	result := 0

	for i := 0; i < maxPreLen; i++ {
		var needToMatch *byte

		matched := true
		for _, line := range lines {
			// Empty lines are not dedented.
			if len(line) == 0 {
				continue
			}

			if needToMatch == nil {
				b := line[i]

				// Stop at the first non-space character
				if b != ' ' && b != '\t' {
					matched = false
					break
				}

				needToMatch = &b
			} else {
				matched = line[i] == *needToMatch
				if !matched {
					break
				}
			}
		}

		if !matched {
			break
		} else {
			result++
		}
	}

	return result
}

// Dedent removes the common whitespace prefix from the lines.
//
// Both tabs and spaces are considered as "whitespace".
func Dedent(lines []string) (result []string) {
	prefixEnd := commonWhitespacePrefix(lines)

	result = make([]string, len(lines))
	for lineNo, line := range lines {
		switch {
		case len(line) == 0:
			result[lineNo] = line
		case len(line) == prefixEnd:
			result[lineNo] = ""
		case len(line) > prefixEnd:
			result[lineNo] = line[prefixEnd:]
		default:
			panic("assertion violated: unexpected dedention case")
		}
	}

	return
}
