package parsecomment_test

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Parquery/gocontracts/parsecomment"
)

func checkFailure(t *testing.T, name string, text string, expected string) {
	if expected == "" {
		panic("Expected error message must not be empty.")
	}

	lines := strings.Split(text, "\n")

	_, err := parsecomment.ToContract(name, lines)

	var errStr string
	if err == nil {
		errStr = "nil"
	} else {
		errStr = fmt.Sprintf("%#v", err.Error())
	}

	if err == nil || err.Error() != expected {
		_, file, line, _ := runtime.Caller(1)

		fmt.Printf("\033[31m%s:%d:\n"+
			"\tExpected error %#v,\n"+
			"\tgot %s\033[39m\n",
			filepath.Base(file), line, expected, errStr)
		t.FailNow()
	}
}

func TestToContract_InvalidNameInPreconditionBlock(t *testing.T) {
	text := `SomeFunc does something.

AnotherFunc requires:
 * x > 0`

	checkFailure(t, "SomeFunc", text,
		"expected function name \"SomeFunc\" in pre-condition block, "+
			"but got \"AnotherFunc\"")
}

func TestToContract_InvalidNameInPostconditionBlock(t *testing.T) {
	text := `SomeFunc does something.

AnotherFunc ensures:
* x > 0`

	checkFailure(t, "SomeFunc", text,
		"expected function name \"SomeFunc\" in post-condition block, "+
			"but got \"AnotherFunc\"")
}

func TestToContract_InvalidNameInPreambleBlock(t *testing.T) {
	text := `SomeFunc does something.

AnotherFunc preamble:
	y := 1`

	checkFailure(t, "SomeFunc", text,
		"expected function name \"SomeFunc\" in preamble block, "+
			"but got \"AnotherFunc\"")
}

func TestToContract_MultiplePreconditionBlocks(t *testing.T) {
	text := `SomeFunc does something.

SomeFunc requires:
* x == 1

SomeFunc requires:
* y == 1`

	checkFailure(t, "SomeFunc", text,
		"multiple pre-condition blocks")
}

func TestToContract_MultiplePostconditionBlocks(t *testing.T) {
	text := `SomeFunc does something.

SomeFunc ensures:
* x == 1

SomeFunc ensures:
* y == 1`

	checkFailure(t, "SomeFunc", text,
		"multiple post-condition blocks")
}

func TestToContract_MultiplePreambles(t *testing.T) {
	text := `SomeFunc does something.

SomeFunc preamble:
	print("hello")

SomeFunc preamble:
	print("world")`

	checkFailure(t, "SomeFunc", text,
		"multiple preambles")
}

func TestToContract_FailedToParsePrecondition(t *testing.T) {
	text := `SomeFunc does something.

SomeFunc requires:
 * x ==`

	checkFailure(t, "SomeFunc", text,
		"failed to parse a pre-condition: "+
			"1:5: expected operand, found 'EOF'")
}

func TestToContract_FailedToParsePostcondition(t *testing.T) {
	text := `SomeFunc does something.

SomeFunc ensures:
 * x ==`

	checkFailure(t, "SomeFunc", text,
		"failed to parse a post-condition: "+
			"1:5: expected operand, found 'EOF'")
}
