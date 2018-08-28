package testcases

// RemoveInCodeWithSemicolon tests the case when pre and postconditions are left in the documentation,
// but removed from code in a function whose body continues with a semi-colon just after the conditions.
var RemoveInCodeWithSemicolon = Case{
	ID:     "remove_in_code_with_semicolon",
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
	}(); return // do something
}
`,
	Expected: `package somepkg

// SomeFunc does something.
//
// SomeFunc ensures:
// * strings.HasPrefix(result, "hello")
//
// Some text here.
func SomeFunc(x int, y int) (result string, err error) {; return // do something
}
`}
