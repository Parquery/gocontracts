package testcases

// NoPreviousConditions tests handling of functions with conditions defined in the documentation, but not yet
// generated in the function body.
var NoPreviousConditions = Case{
	ID: "no_previous_conditions",
	Text: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
// * x > 0
// * x < 100
// * some condition: y > 3
//
// SomeFunc ensures:
// * strings.HasPrefix(result, "hello")
//
// Some text here.
func SomeFunc(x int, y int) (result string, err error) {
	// do something
	return
}
`,
	Expected: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
// * x > 0
// * x < 100
// * some condition: y > 3
//
// SomeFunc ensures:
// * strings.HasPrefix(result, "hello")
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
