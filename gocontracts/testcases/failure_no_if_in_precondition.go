package testcases

// FailureNoIfInPrecondition tests that we correctly detect when pre-condition comment is not followed by an If
// statement in the function body.
var FailureNoIfInPrecondition = Failure{
	ID: "no_if_in_precondition",
	Text: `package somepkg

// SomeFunc does something.
//
// SomeFunc requires:
//  * x > 0
//
// SomeFunc ensures:
//  * result == "oi"
//
// Some text here.
func SomeFunc(x int, y int) (result string, err error) {
	// Pre-condition
	panic("hello")

	return
}
`,
	Error: "expected an 'if' statement after the comment \"Pre-condition\" in function SomeFunc on line 14"}
