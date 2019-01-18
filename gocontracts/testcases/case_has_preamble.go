package testcases

// HasPreamble tests that conditions are correctly generated in a function whose body already contains
// a preamble.
var HasPreamble = Case{
	ID: "has_preamble",
	Text: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
//  * x > 0
//  * x < 100
//  * some condition: y > 3
//
// SomeFunc preamble:
//	print("hello")
//
// SomeFunc ensures:
//  * strings.HasPrefix(result, "hello")
//
// Some text here.
func SomeFunc(x int, y int) (result string, err error) {
	// Preamble starts.
	print("world")
	// Preamble ends.

	// do something
	return
}
`,
	Expected: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
//  * x > 0
//  * x < 100
//  * some condition: y > 3
//
// SomeFunc preamble:
//	print("hello")
//
// SomeFunc ensures:
//  * strings.HasPrefix(result, "hello")
//
// Some text here.
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

	// Preamble starts.
	print("hello")
	// Preamble ends.

	// Post-condition
	defer func() {
		if !(strings.HasPrefix(result, "hello")) {
			panic("Violated: strings.HasPrefix(result, \"hello\")")
		}
	}()

	// do something
	return
}
`}
