package parsebody_test

import (
	"fmt"
	"github.com/Parquery/gocontracts/parsebody"
	"path/filepath"
	"runtime"
	"testing"
)

func checkFailure(t *testing.T, text string, expected string) {
	if expected == "" {
		panic("Expected error message must not be empty.")
	}

	fset, fn, bodyCmtMap, err := parse(text)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = parsebody.ToContract(fset, fn, bodyCmtMap)

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

func TestToContract_StatementBefore(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	print("hello")

	// Pre-condition
	if !(y > 4) {
		panic("Violated: y > 4")
	}
}`

	checkFailure(t, text,
		"unexpected statement before the contract in function SomeFunc on line 4")
}

func TestToContract_StatementInbetween(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	// Pre-condition
	if !(x > 0) {
		panic("Violated: x > 0")
	}

	panic("hello")

	// Post-condition
	defer func() {
		if !(result == "oi") {
			panic("Violated: result == \"oi\"")
		}
	}();

	return
}`

	checkFailure(t, text,
		"unexpected statement between the contract blocks in function SomeFunc on line 9")
}

func TestToContract_NoSwitchInPreconditions(t *testing.T) {
	text := `package somepkg

func SomeFunc(x int, y int) (result string, err error) {
	// Pre-conditions
	panic("hello")

	return
}`

	checkFailure(t, text,
		"expected a 'switch' statement after the comment \"Pre-conditions\" in function SomeFunc on line 5")
}

func TestToContract_NoDeferInPostcondition(t *testing.T) {
	text := `package somepkg

func SomeFunc(x int, y int) (result string, err error) {
	// Pre-condition
	if !(x > 0) {
		panic("Violated: x > 0")
	}
	
	// Post-condition
	panic("hello")

	return
}`

	checkFailure(t, text,
		"expected a defer statement after the comment \"Post-condition\" in function SomeFunc on line 10")
}

func TestToContract_NoIfInPrecondition(t *testing.T) {
	text := `package somepkg

func SomeFunc(x int, y int) (result string, err error) {
	// Pre-condition
	panic("hello")

	return
}`

	checkFailure(t, text,
		"expected an 'if' statement after the comment \"Pre-condition\" in function SomeFunc on line 5")
}

func TestToContract_NoStatementAfterPrecondition(t *testing.T) {
	text := `package somepkg

func SomeFunc(x int, y int) (result string, err error) {
	// Pre-condition
}`

	checkFailure(t, text,
		"found no statement after the comment Pre-condition in function SomeFunc on line 4")
}

func TestToContract_NoStatementAfterPostcondition(t *testing.T) {
	text := `package somepkg

func SomeFunc(x int, y int) (result string, err error) {
	// Pre-condition
	if !(x > 0) {
		panic("Violated: x > 0")
	}

	// Post-condition
}`

	checkFailure(t, text,
		"found no statement after the comment \"Post-condition\" in function SomeFunc on line 9")
}

func TestToContract_PreambleMissingStart(t *testing.T) {
	text := `package somepkg

func SomeFunc(x int, y int) (result string, err error) {
	// Preamble ends.
}`

	checkFailure(t, text,
		"preamble end marker without the start marker in function SomeFunc on line 4")
}

func TestToContract_PreambleMissingEnd(t *testing.T) {
	text := `package somepkg

func SomeFunc(x int, y int) (result string, err error) {
	// Preamble starts.
}`

	checkFailure(t, text,
		"preamble start marker without the end marker in function SomeFunc on line 4")
}

func TestToContract_PreambleMissingEndDueToTypo(t *testing.T) {
	text := `package somepkg

func SomeFunc(x int, y int) (result string, err error) {
	// Preamble starts.

	// Preamble ends..
}`

	checkFailure(t, text,
		"preamble start marker without the end marker in function SomeFunc on line 4")
}

func TestToContract_PreambleInverted(t *testing.T) {
	text := `package somepkg

func SomeFunc(x int, y int) (result string, err error) {
	// Preamble ends.

	// Preamble starts.
}`

	checkFailure(t, text,
		"preamble start marker after the end marker in function SomeFunc on line 6")
}

func TestToContract_PreambleInvertedWithoutNewlines(t *testing.T) {
	text := `package somepkg

func SomeFunc(x int, y int) (result string, err error) {
	// Preamble ends.
	// Preamble starts.
}`

	checkFailure(t, text,
		"preamble start marker after the end marker in function SomeFunc on line 4")
}

func TestToContract_BlockOverlap(t *testing.T) {
	text := `package somepkg

func SomeFunc(x int, y int) (result string, err error) {
	// Preamble starts.
	print("hello")

	// Post-condition
	defer func() {
		if !(strings.HasPrefix(result, "hello")) {
			panic("Violated: strings.HasPrefix(result, \"hello\")")
		}
		// Preamble ends.
	}()

	return
}`

	checkFailure(t, text,
		"unexpected overlap in contract blocks "+
			"in function SomeFunc starting on lines 4 and 7, respectively")
}

func TestToContract_DuplicatePrecondition(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	// Pre-condition
	if !(y > 4) {
		panic("Violated: y > 4")
	}

	// Pre-condition
	if !(x > 4) {
		panic("Violated: x > 4")
	}
}`

	checkFailure(t, text,
		"duplicate pre-condition block found in function SomeFunc on line 9")
}

func TestToContract_DuplicatePreambleStart(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	// Preamble starts.
	print("hello")

	// Preamble starts.
	print("world")
	// Preamble ends.
}`

	checkFailure(t, text,
		"duplicate preamble start found in function SomeFunc on line 7")
}

func TestToContract_DuplicatePreambleEnd(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	// Preamble starts.
	print("hello")

	// Preamble ends.

	print("world")
	// Preamble ends.
}`

	checkFailure(t, text,
		"duplicate preamble end found in function SomeFunc on line 10")
}

func TestToContract_DuplicatePostcondition(t *testing.T) {
	text := `package dummy

func SomeFunc(x int, y int) (result string, err error) {
	// Post-condition
	defer func() {
		if !(strings.HasPrefix(result, "hello")) {
			panic("Violated: strings.HasPrefix(result, \"hello\")")
		}
	}()

	// Post-condition
	defer func() {
		if !(strings.HasPrefix(result, "hello")) {
			panic("Violated: strings.HasPrefix(result, \"hello\")")
		}
	}()

	return
}`

	checkFailure(t, text,
		"duplicate post-condition block found in function SomeFunc on line 11")
}
