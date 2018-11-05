package testcases

// CurlyBracketsOnSameLine tests that condition checks are correctly generated in an empty function where
// curly brackets are written on the same line.
var CurlyBracketsOnSameLine = Case{
	ID: "curly_brackets_on_same_line",
	Text: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
//  * x > 0
//  * x < 100
//  * some condition: y > 3
//
// SomeFunc ensures:
//  * strings.HasPrefix(result, "hello")
//
// Some text here.
func SomeFunc(x int, y int) (result string, err error) {}
`,
	Expected: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
//  * x > 0
//  * x < 100
//  * some condition: y > 3
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

	// Post-condition
	defer func() {
		if !(strings.HasPrefix(result, "hello")) {
			panic("Violated: strings.HasPrefix(result, \"hello\")")
		}
	}()
}
`}
