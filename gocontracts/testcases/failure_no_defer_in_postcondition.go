package testcases

// FailureNoDeferInPostcondition tests that we correctly detect when post-condition comment is not followed by a defer
// statement in the function body.
var FailureNoDeferInPostcondition = Failure{
	ID: "no_defer_in_postcondition",
	Text: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
// * x > 0
//
// SomeFunc ensures:
// * result == "oi"
//
// Some text here.
func SomeFunc(x int, y int) (result string, err error) {
	// Pre-condition
	if !(x > 0) {
		panic("Violated: x > 0")
	}
	
	// Post-condition
	panic("hello")

	return
}
`,
	Error: "expected a defer statement after the comment \"Post-condition\" in function SomeFunc on line 19"}
