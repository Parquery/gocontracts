package testcases

// ConditionsRemoved tests that non-empty functions without conditions in the documentation are stripped of the
// condition checking code.
var ConditionsRemoved = Case{
	ID: "conditions_removed",
	Text: `package somepkg

// SomeFunc does something.
func SomeFunc(x int, y int) (result string, err error) {
	// Pre-condition
	if !(y > 4) {
		panic("Violated: y > 4")
	}

	// do something
	return
}
`,
	Expected: `package somepkg

// SomeFunc does something.
func SomeFunc(x int, y int) (result string, err error) {
	// do something
	return
}
`}
