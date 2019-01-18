package testcases

// HasInitialization tests that condition initializations are correctly handled.
var HasInitialization = Case{
	ID: "initialization",
	Text: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
//  * _, ok := someMap[3]; ok
//
// SomeFunc ensures:
//  * _, ok := someMap[3]; ok
func SomeFunc(someMap map[string]bool) {
	// do something
	return
}
`,
	Expected: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
//  * _, ok := someMap[3]; ok
//
// SomeFunc ensures:
//  * _, ok := someMap[3]; ok
func SomeFunc(someMap map[string]bool) {
	// Pre-condition
	if _, ok := someMap[3]; !ok {
		panic("Violated: _, ok := someMap[3]; ok")
	}

	// Post-condition
	defer func() {
		if _, ok := someMap[3]; !ok {
			panic("Violated: _, ok := someMap[3]; ok")
		}
	}()

	// do something
	return
}
`}
