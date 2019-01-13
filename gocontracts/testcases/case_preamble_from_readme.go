package testcases

// FromReadme tests the code generation with the preamble from Readme.
var FromReadme = Case{
	ID: "from_readme",
	Text: `package somepackage

// increaseFirst increases the first element of the array.
//
// increaseFirst requires:
//  * len(a) > 0
//
// increaseFirst preamble:
//  oldFirst := a[0]
//
// increaseFirst ensures:
//  * a[0] == oldFirst + 1
func increaseFirst(a []int) {
	a[0]++
}
`,
	Expected: `package somepackage

// increaseFirst increases the first element of the array.
//
// increaseFirst requires:
//  * len(a) > 0
//
// increaseFirst preamble:
//  oldFirst := a[0]
//
// increaseFirst ensures:
//  * a[0] == oldFirst + 1
func increaseFirst(a []int) {
	// Pre-condition
	if !(len(a) > 0) {
		panic("Violated: len(a) > 0")
	}

	// Preamble starts.
	oldFirst := a[0]
	// Preamble ends.

	// Post-condition
	defer func() {
		if !(a[0] == oldFirst + 1) {
			panic("Violated: a[0] == oldFirst + 1")
		}
	}()

	a[0]++
}
`}
