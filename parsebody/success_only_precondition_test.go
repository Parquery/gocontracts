package parsebody_test

import (
	"testing"

	"github.com/Parquery/gocontracts/parsebody"
)

func TestToContract_OnlyPrecondition_NoNext(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	// Pre-condition
	if !(y > 4) {
		panic("Violated: y > 4")
	}
}`

	expected := parsebody.Contract{Start: 74, End: 135}
	checkContract(t, text, expected)
}

func TestToContract_OnlyPrecondition_WithNextOnSeparateLine(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	// Pre-condition
	if !(y > 4) {
		panic("Violated: y > 4")
	}

	return
}`

	expected := parsebody.Contract{Start: 74, End: 135, NextNodePos: 138}
	checkContract(t, text, expected)
}

func TestToContract_OnlyPrecondition_WithNextOnSameLine(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	// Pre-condition
	if !(y > 4) {
		panic("Violated: y > 4")
	}; return
}`

	expected := parsebody.Contract{Start: 74, End: 135, NextNodePos: 137}
	checkContract(t, text, expected)
}

func TestToContract_OnlyPreconditions(t *testing.T) {
	// We intentionally do not test for pre-conditions without the next statement or
	// with statement on the same line or the next line since it is assumed that
	// these two cases (namely, a single pre-condition and multiple pre-conditions) share
	// most of the execution paths.
	//
	// Observe the code coverage and add additional tests if the code coverage ever
	// drops due to a refactoring.

	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	// Pre-conditions
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

	return
}`

	expected := parsebody.Contract{Start: 74, End: 273, NextNodePos: 276}
	checkContract(t, text, expected)
}
