package testcases

// RemoveInCode tests the case when pre and postconditions are left in the documentation, but removed from code.
var RemoveInCode = Case{
	ID:     "remove_in_code",
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

	// do something
	return
}
`,
	Expected: `package somepkg

// SomeFunc does something.
//
// SomeFunc ensures:
// * strings.HasPrefix(result, "hello")
//
// Some text here.
func SomeFunc(x int, y int) (result string, err error) {
	// do something
	return
}
`}
