package parsebody_test

import (
	"github.com/Parquery/gocontracts/parsebody"
	"testing"
)

func TestToContract_PrePreamblePost_NoNext(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int, a []string) (result string, err error) {
	// Pre-condition
	if !(y > 4) {
		panic("Violated: y > 4")
	}

	// Preamble starts.
	old_first := a[0]
	// Preamble ends.

	// Post-condition
	defer func() {
		if !(strings.HasPrefix(result, "hello")) {
			panic("Violated: strings.HasPrefix(result, \"hello\")")
		}
	}()
}`

	expected := parsebody.Contract{Start: 86, End: 356}
	checkContract(t, text, expected)
}

func TestToContract_PrePreamblePost_WithNextOnSeparateLine(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int, a []string) (result string, err error) {
	// Pre-condition
	if !(y > 4) {
		panic("Violated: y > 4")
	}

	// Preamble starts.
	old_first := a[0]
	// Preamble ends.

	// Post-condition
	defer func() {
		if !(strings.HasPrefix(result, "hello")) {
			panic("Violated: strings.HasPrefix(result, \"hello\")")
		}
	}()

	return
}`

	expected := parsebody.Contract{Start: 86, End: 356, NextNodePos: 359}
	checkContract(t, text, expected)
}

func TestToContract_PrePreamblePost_WithNextOnSameLine(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int, a []string) (result string, err error) {
	// Pre-condition
	if !(y > 4) {
		panic("Violated: y > 4")
	}

	// Preamble starts.
	old_first := a[0]
	// Preamble ends.

	// Post-condition
	defer func() {
		if !(strings.HasPrefix(result, "hello")) {
			panic("Violated: strings.HasPrefix(result, \"hello\")")
		}
	}(); return
}`

	expected := parsebody.Contract{Start: 86, End: 356, NextNodePos: 358}
	checkContract(t, text, expected)
}
