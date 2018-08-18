package testcases

var NoConditions = Case{
	ID: "no_conditions",
	Text: `package somepkg

// SomeFunc does something.
func SomeFunc(x int, y int) (result string, err error) {
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