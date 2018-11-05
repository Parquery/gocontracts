package testcases

// MultipleFunctions tests that conditions are correctly generated in a file containing more than one function.
var MultipleFunctions = Case{
	ID: "multiple_functions",
	Text: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
//  * x > 0
//
// Some text here.
func SomeFunc(x int, y int) (result string, err error) {
	// do something
	return
}

// AnotherFunc does something.
//
// AnotherFunc requires:
//  * x > 0
//
// Some text here.
func AnotherFunc(x int, y int) (result string, err error) {
	// do something
	return
}
`,
	Expected: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
//  * x > 0
//
// Some text here.
func SomeFunc(x int, y int) (result string, err error) {
	// Pre-condition
	if !(x > 0) {
		panic("Violated: x > 0")
	}

	// do something
	return
}

// AnotherFunc does something.
//
// AnotherFunc requires:
//  * x > 0
//
// Some text here.
func AnotherFunc(x int, y int) (result string, err error) {
	// Pre-condition
	if !(x > 0) {
		panic("Violated: x > 0")
	}

	// do something
	return
}
`}
