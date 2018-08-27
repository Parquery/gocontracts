package testcases

// ConditionsRemovedEmptyFunction tests that empty functions without conditions in the documentation are stripped of
// the condition checking code.
var ConditionsRemovedEmptyFunction = Case{
	ID: "conditions_removed_empty_function",
	Text: `package somepkg

// SomeFunc does something.
func SomeFunc(x int, y int) () {
	// Pre-condition
	if !(y > 4) {
		panic("Violated: y > 4")
	}
}
`,
	Expected: `package somepkg

// SomeFunc does something.
func SomeFunc(x int, y int) () {}
`}
