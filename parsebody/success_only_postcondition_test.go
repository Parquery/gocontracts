package parsebody_test

import (
	"testing"

	"github.com/Parquery/gocontracts/parsebody"
)

func TestToContract_OnlyPostcondition_NoNext(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	// Post-condition
	defer func() {
		if !(strings.HasPrefix(result, "hello")) {
			panic("Violated: strings.HasPrefix(result, \"hello\")")
		}
	}()
}`

	expected := parsebody.Contract{Start: 74, End: 220}
	checkContract(t, text, expected)
}

func TestToContract_OnlyPostcondition_WithNextOnSeparateLine(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	// Post-condition
	defer func() {
		if !(strings.HasPrefix(result, "hello")) {
			panic("Violated: strings.HasPrefix(result, \"hello\")")
		}
	}()

	return
}`

	expected := parsebody.Contract{Start: 74, End: 220, NextNodePos: 223}
	checkContract(t, text, expected)
}

func TestToContract_OnlyPostcondition_WithNextOnSameLine(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	// Post-condition
	defer func() {
		if !(strings.HasPrefix(result, "hello")) {
			panic("Violated: strings.HasPrefix(result, \"hello\")")
		}
	}(); return
}`

	expected := parsebody.Contract{Start: 74, End: 220, NextNodePos: 222}
	checkContract(t, text, expected)
}

func TestToContract_OnlyPostconditions(t *testing.T) {
	// We intentionally do not test for post-conditions without the next statement or
	// with statement on the same line or the next line since it is assumed that
	// these two cases (namely, a single post-condition and multiple post-conditions) share
	// most of the execution paths.
	//
	// Observe the code coverage and add additional tests if the code coverage ever
	// drops due to a refactoring.

	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	// Post-conditions
	defer func() {
		switch {
		case !(x > 0):
			panic("Violated: x > 0")
		case !(x < 100):
			panic("Violated: x < 100")
		case !(y > 3):
			panic("Violated: some condition: y > 3")
		default:
			// Pass
		}
	}()

	return
}`

	expected := parsebody.Contract{Start: 74, End: 305, NextNodePos: 308}
	checkContract(t, text, expected)
}
