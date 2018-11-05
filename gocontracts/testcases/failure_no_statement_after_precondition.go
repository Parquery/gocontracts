package testcases

// FailureNoStatementAfterPrecondtion tests that we correctly detect when pre-condition comment is not followed by
// a statement in the function body.
var FailureNoStatementAfterPrecondtion = Failure{
	ID: "no_statement_after_precondition",
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
}
`,
	Error: "found no statement after the comment Pre-condition in function SomeFunc on line 13"}
