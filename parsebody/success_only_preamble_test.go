package parsebody_test

import (
	"testing"

	"github.com/Parquery/gocontracts/parsebody"
)

func TestToContract_OnlyPreamble_NoNext(t *testing.T) {
	text := `package dummy

func SomeFunc(a []string) {
	// Preamble starts.
	old_first := a[0]
	// Preamble ends.
}`

	expected := parsebody.Contract{Start: 45, End: 85}
	checkContract(t, text, expected)
}

func TestToContract_OnlyPreamble_WithNextOnSeparateLine(t *testing.T) {
	text := `package dummy

func SomeFunc(a []string) {
	// Preamble starts.
	old_first := a[0]
	// Preamble ends.

	return
}`

	expected := parsebody.Contract{Start: 45, End: 85, NextNodePos: 105}
	checkContract(t, text, expected)
}

func TestToContract_PreambleWithoutNewline(t *testing.T) {
	text := `package somepkg

func SomeFunc(x int, y int) (result string, err error) {
	// Preamble starts.
	// Preamble ends.

	return
}`

	expected := parsebody.Contract{Start: 76, End: 76, NextNodePos: 117}
	checkContract(t, text, expected)
}
