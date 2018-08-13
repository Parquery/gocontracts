package testcases

var FailureNoStatementAfterPostcondtion = Failure{
	ID: "no_statement_after_postcondition",
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
}
`,
	Error: "found no statement after the comment \"Post-condition\" in function SomeFunc on line 18"}
