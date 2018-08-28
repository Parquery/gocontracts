package testcases

// RemoveInCodeOfEmptyFunction tests the case when pre and postconditions are left in the documentation,
// but are removed from code of a function whose body is empty.
var RemoveInCodeOfEmptyFunction = Case{
	ID:     "remove_in_code_of_empty_function",
	Remove: true,
	Text: `package somepkg

// SomeFunc does something.
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
	}()
}
`,
	Expected: `package somepkg

// SomeFunc does something.
//
// SomeFunc ensures:
// * strings.HasPrefix(result, "hello")
//
// Some text here.
func SomeFunc(x int, y int) (result string, err error) {}
`}
