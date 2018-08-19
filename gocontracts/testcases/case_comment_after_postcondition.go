package testcases

// CommentAfterPostcondition tests that we correctly handle a function whose body contains a comment following the
// post-condition.
var CommentAfterPostcondition = Case{
	ID: "comment_after_postcondition",
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
	// Post-condition
	defer func() {
		if !(strings.HasPrefix(result, "hello")) {
			panic("Violated: strings.HasPrefix(result, \"hello\")")
		}
	}();  // some comment here

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
	}();  // some comment here

	return
}
`}
