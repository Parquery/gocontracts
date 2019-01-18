package parsebody_test

import (
	"testing"

	"github.com/Parquery/gocontracts/parsebody"
)

func TestToContract_NoContract_CommentStatement(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	// do something
	return
}`

	expected := parsebody.Contract{NextNodePos: 74}
	checkContract(t, text, expected)
}

func TestToContract_EmptyMultilineFunction(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
}`

	expected := parsebody.Contract{}
	checkContract(t, text, expected)
}

func TestToContract_EmptySinglelineFunction(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {}`

	expected := parsebody.Contract{}
	checkContract(t, text, expected)
}
