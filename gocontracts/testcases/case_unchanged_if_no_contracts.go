package testcases

// UnchangedIfNoContracts tests that the function is unchanged if there were no contracts in the documentation
// nor in the body of the function.
var UnchangedIfNoContracts = Case{
	ID: "unchanged_if_no_contracts",
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
