package testcases

// ConditionsRemovedInCommentOfEmptyFunction tests that empty functions without conditions in the documentation are
// stripped of the condition checking code.
var ConditionsRemovedInCommentOfEmptyFunction = Case{
	ID: "conditions_removed_in_comment_of_empty_function",
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
